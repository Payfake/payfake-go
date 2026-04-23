package payfake

import (
	"context"
	"net/http"
)

// AuthNamespace wraps /api/v1/auth endpoints.
// These are Payfake-specific, no Paystack equivalent.
// Use the returned access_token with Control and Merchant methods.
type AuthNamespace struct {
	client *Client
}

// Register creates a new merchant account.
// Returns the merchant profile and an access token.
func (n *AuthNamespace) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	var out AuthResponse
	if err := n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/auth/register", input, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}

// Login authenticates a merchant and returns an access token.
func (n *AuthNamespace) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	var out AuthResponse
	if err := n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/auth/login", input, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}

// Me returns the currently authenticated merchant profile.
// Requires a valid access token.
func (n *AuthNamespace) Me(ctx context.Context, token string) (*MerchantProfile, error) {
	var out MerchantProfile
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/auth/me", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetKeys returns the merchant's public and secret keys.
func (n *AuthNamespace) GetKeys(ctx context.Context, token string) (*KeysResponse, error) {
	var out KeysResponse
	if err := n.client.doWithJWT(ctx, http.MethodGet, "/api/v1/auth/keys", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}

// RegenerateKeys rotates the merchant's key pair.
// All requests using the old secret key will fail immediately after this call.
func (n *AuthNamespace) RegenerateKeys(ctx context.Context, token string) (*KeysResponse, error) {
	var out KeysResponse
	if err := n.client.doWithJWT(ctx, http.MethodPost, "/api/v1/auth/keys/regenerate", nil, &out, token); err != nil {
		return nil, err
	}
	return &out, nil
}
