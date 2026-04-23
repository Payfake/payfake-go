package payfake

import (
	"context"
	"fmt"
	"net/http"
)

// TransactionNamespace wraps /transaction endpoints.
// These match https://api.paystack.co/transaction exactly.
// Auth: Bearer sk_test_xxx
type TransactionNamespace struct {
	client *Client
}

// Initialize creates a new pending transaction.
// Returns the authorization_url, access_code and reference.
// Redirect your customer to authorization_url to open the checkout page.
//
// Response matches Paystack exactly:
//
//	{ "authorization_url": "...", "access_code": "...", "reference": "..." }
func (n *TransactionNamespace) Initialize(ctx context.Context, input InitializeInput) (*InitializeResponse, error) {
	var out InitializeResponse
	if err := n.client.do(ctx, http.MethodPost, "/transaction/initialize", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Verify retrieves a transaction by reference.
// Call this after the checkout completes to confirm the outcome.
// Always verify before delivering value, never trust the callback URL alone.
func (n *TransactionNamespace) Verify(ctx context.Context, reference string) (*Transaction, error) {
	var out Transaction
	if err := n.client.do(ctx, http.MethodGet, "/transaction/verify/"+reference, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a single transaction by ID.
func (n *TransactionNamespace) Fetch(ctx context.Context, id string) (*Transaction, error) {
	var out Transaction
	if err := n.client.do(ctx, http.MethodGet, "/transaction/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of transactions.
// Filter by status: "success", "failed", "pending", "abandoned".
// Uses Paystack's perPage param (not per_page).
func (n *TransactionNamespace) List(ctx context.Context, opts ListOptions, status string) (*TransactionList, error) {
	path := fmt.Sprintf("/transaction?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if status != "" {
		path += "&status=" + status
	}
	var out TransactionList
	if err := n.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Refund marks a successful transaction as reversed.
func (n *TransactionNamespace) Refund(ctx context.Context, id string) (*Transaction, error) {
	var out Transaction
	if err := n.client.do(ctx, http.MethodPost, "/transaction/"+id+"/refund", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PublicFetch loads transaction details for the checkout page using the access code.
// No secret key required, authenticated via access_code in the URL.
// Returns merchant branding, amount, currency and current charge flow status.
// Call this on checkout page mount.
func (n *TransactionNamespace) PublicFetch(ctx context.Context, accessCode string) (*PublicTransactionResponse, error) {
	var out PublicTransactionResponse
	if err := n.client.doPublic(ctx, http.MethodGet, "/api/v1/public/transaction/"+accessCode, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PublicVerify polls transaction status for MoMo pay_offline state.
// No secret key required. Poll every 3 seconds, stop when Status is
// "success" or "failed".
//
//	for {
//	    result, _ := client.Transaction.PublicVerify(ctx, reference)
//	    if result.Status == "success" || result.Status == "failed" {
//	        break
//	    }
//	    time.Sleep(3 * time.Second)
//	}
func (n *TransactionNamespace) PublicVerify(ctx context.Context, reference string) (*PublicVerifyResponse, error) {
	var out PublicVerifyResponse
	if err := n.client.doPublic(ctx, http.MethodGet, "/api/v1/public/transaction/verify/"+reference, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
