package payfake

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ControlNamespace wraps /api/v1/control endpoints.
// These are Payfake-specific, no Paystack equivalent.
// Auth: Bearer JWT (from Auth.Login)
type ControlNamespace struct {
	client *Client
}

// GetStats returns aggregated overview numbers for the dashboard.
func (n *ControlNamespace) GetStats(ctx context.Context, token string) (*MerchantStats, error) {
	var out MerchantStats
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/control/stats", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetScenario returns the current scenario config for the merchant.
func (n *ControlNamespace) GetScenario(ctx context.Context, token string) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/control/scenario", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateScenario updates the scenario config.
// Only non-nil pointer fields are updated, pass nil to leave a field unchanged.
//
// Common patterns:
//
//	// Force all charges to fail with insufficient funds
//	status := "failed"
//	code := "CHARGE_INSUFFICIENT_FUNDS"
//	client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
//	    ForceStatus: &status,
//	    ErrorCode:   &code,
//	})
//
//	// 30% random failure rate with 2 second delay
//	rate := 0.3
//	delay := 2000
//	client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
//	    FailureRate: &rate,
//	    DelayMS:     &delay,
//	})
func (n *ControlNamespace) UpdateScenario(ctx context.Context, token string, input UpdateScenarioInput) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := n.client.doWithJWT(ctx, http.MethodPut, "/api/v1/control/scenario", input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResetScenario resets the scenario config to defaults.
// After reset all charges succeed with no delay.
func (n *ControlNamespace) ResetScenario(ctx context.Context, token string) (*ScenarioConfig, error) {
	var out ScenarioConfig
	if err := n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/control/scenario/reset", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTransactions returns a JWT-authenticated transaction list for the dashboard.
// Supports search by reference or customer email and status filter.
func (n *ControlNamespace) ListTransactions(ctx context.Context, token string, opts ListOptions, status, search string) (*TransactionList, error) {
	path := fmt.Sprintf("/api/v1/control/transactions?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if status != "" {
		path += "&status=" + status
	}
	if search != "" {
		path += "&search=" + search
	}
	var out TransactionList
	if err := n.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListCustomers returns a JWT-authenticated customer list for the dashboard.
func (n *ControlNamespace) ListCustomers(ctx context.Context, token string, opts ListOptions) (*CustomerList, error) {
	path := fmt.Sprintf("/api/v1/control/customers?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	var out CustomerList
	if err := n.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ForceTransaction forces a pending transaction to a specific terminal state.
// Bypasses the scenario engine entirely, useful for deterministic test cases.
func (n *ControlNamespace) ForceTransaction(ctx context.Context, token, reference string, input ForceTransactionInput) (*Transaction, error) {
	var out Transaction
	path := fmt.Sprintf("/api/v1/control/transactions/%s/force", reference)
	if err := n.client.doWithJWT(ctx, http.MethodPost, path, input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWebhooks returns paginated webhook events.
func (n *ControlNamespace) ListWebhooks(ctx context.Context, token string, opts ListOptions) (*WebhookList, error) {
	path := fmt.Sprintf("/api/v1/control/webhooks?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	var out WebhookList
	if err := n.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// RetryWebhook manually re-triggers delivery for a webhook event.
func (n *ControlNamespace) RetryWebhook(ctx context.Context, token, id string) error {
	return n.client.doWithJWT(ctx, http.MethodPost,
		"/api/v1/control/webhooks/"+id+"/retry", nil, nil, token)
}

// GetWebhookAttempts returns all delivery attempts for a webhook event.
func (n *ControlNamespace) GetWebhookAttempts(ctx context.Context, token, id string) ([]WebhookAttempt, error) {
	var out struct {
		Data []WebhookAttempt `json:"data"`
	}
	if err := n.client.doWithJWT(ctx, http.MethodGet,
		"/api/v1/control/webhooks/"+id+"/attempts", nil, &out, token); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// GetLogs returns paginated request/response introspection logs.
func (n *ControlNamespace) GetLogs(ctx context.Context, token string, opts ListOptions) (*LogList, error) {
	path := fmt.Sprintf("/api/v1/control/logs?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	var out LogList
	if err := n.client.doWithJWT(ctx, http.MethodGet, path, nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// ClearLogs permanently deletes all request logs for the merchant.
func (n *ControlNamespace) ClearLogs(ctx context.Context, token string) error {
	return n.client.doWithJWT(ctx, http.MethodDelete, "/api/v1/control/logs", nil, nil, token)
}

// GetOTPLogs returns OTP codes generated during charge flows.
// This is the primary way to get OTPs during testing without a real phone.
//
//	logs, _ := client.Control.GetOTPLogs(ctx, token, "TXN_xxx", payfake.ListOptions{})
//	otp := logs[0].OTPCode
//
// Pass reference to filter for a specific transaction.
// Pass empty string to get all OTP logs paginated.
func (n *ControlNamespace) GetOTPLogs(ctx context.Context, token, reference string, opts ListOptions) ([]OTPLog, error) {
	var path string
	if reference != "" {
		path = "/api/v1/control/otp-logs?reference=" + reference
	} else {
		path = fmt.Sprintf("/api/v1/control/otp-logs?page=%d&perPage=%d",
			pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	}

	// The response shape differs based on whether reference is set:
	// with reference: { data: [] } (flat array, no meta)
	// without reference: { data: [], meta: {} }
	// We handle both by trying the paginated shape first.
	var raw json.RawMessage
	if err := n.client.doWithJWT(ctx, http.MethodGet, path, nil, &raw, token); err != nil {
		return nil, err
	}

	// Try paginated shape first
	var paginated OTPLogList
	if err := json.Unmarshal(raw, &paginated); err == nil && paginated.Data != nil {
		return paginated.Data, nil
	}

	// Fall back to flat array
	var flat []OTPLog
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("payfake: parse otp logs: %w", err)
	}
	return flat, nil
}
