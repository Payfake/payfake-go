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

const defaultBaseURL = "https://api.payfake.co"

// Client is the root Payfake SDK client.
// All API namespaces hang off this struct.
//
//	client := payfake.New(payfake.Config{
//	    SecretKey: "sk_test_xxx",
//	})
//
// For self-hosted instances:
//
//	client := payfake.New(payfake.Config{
//	    SecretKey: "sk_test_xxx",
//	    BaseURL:   "http://localhost:8080",
//	})
type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client

	// Namespaces
	Auth        *AuthNamespace
	Transaction *TransactionNamespace
	Charge      *ChargeNamespace
	Customer    *CustomerNamespace
	Merchant    *MerchantNamespace
	Control     *ControlNamespace
}

// Config holds the client configuration.
type Config struct {
	// SecretKey is the merchant's sk_test_xxx key.
	// Required for all Paystack-compatible endpoints.
	SecretKey string
	// BaseURL defaults to https://api.payfake.co.
	// Override for self-hosted: "http://localhost:8080"
	BaseURL string
	// Timeout defaults to 30 seconds.
	Timeout time.Duration
}

// New creates a new Payfake SDK client.
func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	c := &Client{
		baseURL:   baseURL,
		secretKey: cfg.SecretKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	c.Auth = &AuthNamespace{client: c}
	c.Transaction = &TransactionNamespace{client: c}
	c.Charge = &ChargeNamespace{client: c}
	c.Customer = &CustomerNamespace{client: c}
	c.Merchant = &MerchantNamespace{client: c}
	c.Control = &ControlNamespace{client: c}

	return c
}

// do executes an HTTP request authenticated with the secret key.
// Used for all Paystack-compatible endpoints (/transaction, /charge, /customer).
func (c *Client) do(ctx context.Context, method, path string, body, target any) error {
	return c.request(ctx, method, path, body, target, c.secretKey)
}

// doWithJWT executes an HTTP request authenticated with a JWT token.
// Used for Payfake-specific endpoints (/api/v1/auth, /api/v1/control, /api/v1/merchant).
func (c *Client) doWithJWT(ctx context.Context, method, path string, body, target any, token string) error {
	return c.request(ctx, method, path, body, target, token)
}

// doPublic executes an unauthenticated HTTP request.
// Used for public checkout endpoints (/api/v1/public/*).
func (c *Client) doPublic(ctx context.Context, method, path string, body, target any) error {
	return c.request(ctx, method, path, body, target, "")
}

func (c *Client) request(ctx context.Context, method, path string, body, target any, token string) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("payfake: marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("payfake: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("payfake: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("payfake: read response body: %w", err)
	}

	var env envelope
	if err := json.Unmarshal(respBytes, &env); err != nil {
		return fmt.Errorf("payfake: parse response (status %d): %w", resp.StatusCode, err)
	}

	// status is a boolean in the Paystack/Payfake envelope.
	// false means the API returned an error regardless of HTTP status code.
	if !env.Status {
		// Flatten the errors map into our ErrorField slice.
		var fields []ErrorField
		for fieldName, rules := range env.Errors {
			for _, rule := range rules {
				fields = append(fields, ErrorField{
					Field:   fieldName,
					Rule:    rule.Rule,
					Message: rule.Message,
				})
			}
		}
		// Error code travels in the X-Payfake-Code header, not the body.
		// This keeps the body 100% Paystack-compatible.
		return &SDKError{
			Code:       resp.Header.Get("X-Payfake-Code"),
			Message:    env.Message,
			Fields:     fields,
			HTTPStatus: resp.StatusCode,
		}
	}

	if target != nil && env.Data != nil {
		// Re-marshal the data field and unmarshal into the typed target.
		// envelope.Data is any{} so we need this round-trip to get
		// the concrete type the caller expects.
		dataBytes, err := json.Marshal(env.Data)
		if err != nil {
			return fmt.Errorf("payfake: re-marshal data: %w", err)
		}
		if err := json.Unmarshal(dataBytes, target); err != nil {
			return fmt.Errorf("payfake: unmarshal data into target: %w", err)
		}
	}

	return nil
}

// helpers

func pageOrDefault(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func perPageOrDefault(perPage int) int {
	if perPage < 1 {
		return 50
	}
	if perPage > 100 {
		return 100
	}
	return perPage
}
