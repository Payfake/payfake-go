package payfake

import (
	"context"
	"net/http"
)

// MerchantNamespace wraps /api/v1/merchant endpoints.
// These are Payfake-specific, no Paystack equivalent.
// Auth: Bearer JWT (from Auth.Login)
type MerchantNamespace struct {
	client *Client
}

// GetProfile returns the full merchant profile.
func (n *MerchantNamespace) GetProfile(ctx context.Context, token string) (*MerchantProfile, error) {
	var out MerchantProfile
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/merchant", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProfile updates the merchant's business name and/or webhook URL.
func (n *MerchantNamespace) UpdateProfile(ctx context.Context, token string, input UpdateProfileInput) (*MerchantProfile, error) {
	var out MerchantProfile
	if err := n.client.doWithJWT(ctx, http.MethodPut, "/api/v1/merchant", input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWebhookURL returns the merchant's current webhook URL and whether it's configured.
func (n *MerchantNamespace) GetWebhookURL(ctx context.Context, token string) (*WebhookConfig, error) {
	var out WebhookConfig
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/merchant/webhook", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateWebhookURL sets the merchant's webhook URL.
func (n *MerchantNamespace) UpdateWebhookURL(ctx context.Context, token, webhookURL string) error {
	return n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/merchant/webhook",
		map[string]string{"webhook_url": webhookURL}, nil, token)
}

// TestWebhook fires a test webhook to verify the endpoint is reachable.
// Rate limited to 5 requests per minute per merchant.
func (n *MerchantNamespace) TestWebhook(ctx context.Context, token string) (*WebhookTestResult, error) {
	var out WebhookTestResult
	if err := n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/merchant/webhook/test", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}
