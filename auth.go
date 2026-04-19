package payfake

import (
	"context"
	"net/http"
)

// AuthService wraps the /auth endpoints.
// Register and Login require no authentication, they are how
// you get credentials in the first place.
// GetKeys and RegenerateKeys require a JWT token from Login().
type AuthService struct {
	client *Client
}

// Register creates a new merchant account.
// Returns merchant data and a JWT token on success.
// Returns ErrEmailTaken (code AUTH_EMAIL_TAKEN) if the email is already registered.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResponse, error) {
	var out RegisterResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/auth/register", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Login authenticates a merchant and returns a JWT token.
// Store the token, you need it for all control and key endpoints.
// The token expires after JWT_EXPIRY_HOURS (default 24 hours).
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResponse, error) {
	var out LoginResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/auth/login", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetKeys fetches the merchant's current public and secret keys.
// Requires a valid JWT token from Login().
func (s *AuthService) GetKeys(ctx context.Context, token string) (*KeysResponse, error) {
	var out KeysResponse
	if err := s.client.doWithJWT(ctx, http.MethodGet, "/api/v1/auth/keys", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// RegenerateKeys generates a new key pair for the merchant.
// The old secret key is immediately invalid after this call.
// Update your environment variables before calling this.
// Requires a valid JWT token from Login().
func (s *AuthService) RegenerateKeys(ctx context.Context, token string) (*KeysResponse, error) {
	var out KeysResponse
	if err := s.client.doWithJWT(ctx, http.MethodPost, "/api/v1/auth/keys/regenerate", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProfile fetches the merchant's full profile.
// Requires JWT token from Login().
func (s *AuthService) GetProfile(ctx context.Context, token string) (*MerchantProfile, error) {
	var out MerchantProfile
	if err := s.client.doWithJWT(ctx, http.MethodGet, "/api/v1/merchant", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProfile updates the merchant's business name and/or webhook URL.
func (s *AuthService) UpdateProfile(ctx context.Context, token string, input UpdateProfileInput) (*MerchantProfile, error) {
	var out MerchantProfile
	if err := s.client.doWithJWT(ctx, http.MethodPut, "/api/v1/merchant", input, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWebhookURL fetches the merchant's current webhook URL.
func (s *AuthService) GetWebhookURL(ctx context.Context, token string) (*WebhookConfig, error) {
	var out WebhookConfig
	if err := s.client.doWithJWT(ctx, http.MethodGet, "/api/v1/merchant/webhook", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateWebhookURL sets the merchant's webhook URL.
func (s *AuthService) UpdateWebhookURL(ctx context.Context, token, webhookURL string) error {
	return s.client.doWithJWT(ctx, http.MethodPost, "/api/v1/merchant/webhook",
		map[string]string{"webhook_url": webhookURL}, nil, token)
}

// TestWebhook fires a test webhook to the merchant's configured URL.
func (s *AuthService) TestWebhook(ctx context.Context, token string) (*WebhookTestResult, error) {
	var out WebhookTestResult
	if err := s.client.doWithJWT(ctx, http.MethodPost, "/api/v1/merchant/webhook/test", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}
