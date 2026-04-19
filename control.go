package payfake

import (
	"context"
	"fmt"
	"net/http"
)

// ControlService wraps the /control namespace.
// These endpoints require JWT auth (dashboard token) not a secret key.
// Pass the JWT token to each method explicitly, control operations
// are dashboard actions, not application-level API calls.
type ControlService struct {
	client *Client
}

// GetScenario fetches the current scenario config for the merchant.
func (s *ControlService) GetScenario(ctx context.Context, token string) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := s.client.doWithJWT(ctx, http.MethodGet, "/api/v1/control/scenario", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateScenario updates the scenario config.
// Only non-nil fields are updated, use pointer fields to send partial updates.
func (s *ControlService) UpdateScenario(ctx context.Context, token string, input UpdateScenarioInput) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := s.client.doWithJWT(ctx, http.MethodPut, "/api/v1/control/scenario", input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResetScenario resets the scenario config to defaults.
// After reset all transactions succeed with no delay.
func (s *ControlService) ResetScenario(ctx context.Context, token string) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := s.client.doWithJWT(ctx, http.MethodPost, "/api/v1/control/scenario/reset", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWebhooks returns paginated webhook events.
func (s *ControlService) ListWebhooks(ctx context.Context, token string, opts ListOptions) ([]WebhookEvent, error) {
	var out struct {
		Webhooks []WebhookEvent `json:"webhooks"`
	}
	path := fmt.Sprintf("/api/v1/control/webhooks?page=%d&per_page=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if err := s.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return out.Webhooks, nil
}

// RetryWebhook manually re-triggers delivery for a webhook event.
func (s *ControlService) RetryWebhook(ctx context.Context, token, id string) error {
	return s.client.doWithJWT(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/control/webhooks/%s/retry", id), nil, nil, token)
}

// GetWebhookAttempts returns delivery attempts for a webhook event.
func (s *ControlService) GetWebhookAttempts(ctx context.Context, token, id string) ([]WebhookAttempt, error) {
	var out struct {
		Attempts []WebhookAttempt `json:"attempts"`
	}
	if err := s.client.doWithJWT(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/control/webhooks/%s/attempts", id), nil, &out, token); err != nil {
		return nil, err
	}
	return out.Attempts, nil
}

// ForceTransaction forces a pending transaction to a specific terminal state.
func (s *ControlService) ForceTransaction(ctx context.Context, token, reference string, input ForceTransactionInput) (*Transaction, error) {
	var out Transaction
	if err := s.client.doWithJWT(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/control/transactions/%s/force", reference), input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLogs returns paginated request/response logs.
func (s *ControlService) GetLogs(ctx context.Context, token string, opts ListOptions) ([]RequestLog, error) {
	var out struct {
		Logs []RequestLog `json:"logs"`
	}
	path := fmt.Sprintf("/api/v1/control/logs?page=%d&per_page=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if err := s.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return out.Logs, nil
}

// ClearLogs permanently deletes all logs for the merchant.
func (s *ControlService) ClearLogs(ctx context.Context, token string) error {
	return s.client.doWithJWT(ctx, http.MethodDelete, "/api/v1/control/logs", nil, nil, token)
}

// GetOTPLogs returns OTP codes generated during charge flows.
// Use reference to filter for a specific transaction.
// This is the primary way to get OTPs during testing without a real phone.
//
// Example:
//
//	logs, err := client.Control.GetOTPLogs(ctx, token, "TXN_xxx")
//	fmt.Println("OTP:", logs[0].OTPCode)
func (s *ControlService) GetOTPLogs(ctx context.Context, token string, reference string) ([]OTPLog, error) {
	path := "/api/v1/control/otp-logs"
	if reference != "" {
		path += "?reference=" + reference
	}

	var out struct {
		OTPLogs []OTPLog `json:"otp_logs"`
	}
	if err := s.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return out.OTPLogs, nil
}
