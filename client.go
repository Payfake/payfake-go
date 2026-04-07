package payfake

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "http://localhost:8080"

// Client is the root Payfake SDK client.
// All API namespaces hang off this struct, callers access them
// via client.Transaction, client.Charge, client.Customer etc.
// This mirrors how official SDKs (Stripe, Paystack) are structured
// one client, multiple namespaces, no global state.
type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client

	// Namespace services, each wraps a slice of related API endpoints.
	Auth        *AuthService
	Transaction *TransactionService
	Charge      *ChargeService
	Customer    *CustomerService
	Control     *ControlService
}

// Config holds the configuration for the SDK client.
type Config struct {
	// SecretKey is the merchant's sk_test_ key.
	// Required for all API calls.
	SecretKey string
	// BaseURL is the Payfake server URL.
	// Defaults to http://localhost:8080 if not set.
	// Override this if you're running Payfake on a different port
	// or have deployed it to a server.
	BaseURL string
	// Timeout is the HTTP client timeout.
	// Defaults to 30 seconds if not set.
	Timeout time.Duration
}

// New creates a new Payfake SDK client.
// This is the only constructor, all fields have sensible defaults
// so the minimum viable setup is just a secret key.
//
// Example:
//
//	client := payfake.New(payfake.Config{
//	    SecretKey: "sk_test_xxx",
//	})
func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		// 30 seconds covers MoMo async resolution delays
		// MoMo charges can take up to the configured delay_ms
		// before the webhook fires. Card charges are near-instant.
		timeout = 30 * time.Second
	}

	c := &Client{
		baseURL:   baseURL,
		secretKey: cfg.SecretKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	// Wire up namespace services, each gets a reference to the
	// parent client so they can call c.do() for HTTP requests.
	c.Auth = &AuthService{client: c}
	c.Transaction = &TransactionService{client: c}
	c.Charge = &ChargeService{client: c}
	c.Customer = &CustomerService{client: c}
	c.Control = &ControlService{client: c}

	return c
}

// do is the single HTTP execution method for the entire SDK.
// Every namespace method builds a request and calls this.
// Centralizing HTTP here means auth headers, error parsing,
// and response unwrapping happen exactly once, not scattered
// across 27 different endpoint methods.
func (c *Client) do(ctx context.Context, method, path string, body any, target any) error {
	// Serialize the request body to JSON if one was provided.
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Attach the secret key as a Bearer token, same as calling
	// the API directly. The SDK doesn't change the auth scheme.
	if c.secretKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.secretKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the Payfake envelope regardless of status code.
	// We need the envelope to extract the error code and message
	// even on failure, raw HTTP status codes alone aren't enough.
	var envelope apiResponse
	if err := json.Unmarshal(respBytes, &envelope); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Any non-success status means the API returned an error envelope.
	// We convert it to an SDKError so callers get typed error handling.
	if envelope.Status != "success" {
		return &SDKError{
			Code:       envelope.Code,
			Message:    envelope.Message,
			Fields:     envelope.Errors,
			HTTPStatus: resp.StatusCode,
		}
	}

	// Success, unmarshal the data field into the target struct.
	// We re-marshal just the data field and unmarshal into target
	// because envelope.Data is any, we need the concrete type.
	if target != nil && envelope.Data != nil {
		dataBytes, err := json.Marshal(envelope.Data)
		if err != nil {
			return fmt.Errorf("failed to re-marshal data: %w", err)
		}
		if err := json.Unmarshal(dataBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal data into target: %w", err)
		}
	}

	return nil
}

// doWithJWT is like do but attaches a JWT token instead of a secret key.
// Used for dashboard/control endpoints that require JWT auth
// rather than secret key auth.
func (c *Client) doWithJWT(ctx context.Context, method, path string, body any, target any, token string) error {
	original := c.secretKey
	c.secretKey = token
	err := c.do(ctx, method, path, body, target)
	c.secretKey = original
	return err
}
