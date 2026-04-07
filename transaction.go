package payfake

import (
	"context"
	"fmt"
	"net/http"
)

type TransactionService struct {
	client *Client
}

// Initialize creates a new pending transaction and returns the
// authorization URL and access code for the payment popup.
func (s *TransactionService) Initialize(ctx context.Context, input InitializeInput) (*InitializeResponse, error) {
	var out InitializeResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/transaction/initialize", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Verify retrieves a transaction by reference.
// Call this after the payment popup closes to confirm the outcome.
func (s *TransactionService) Verify(ctx context.Context, reference string) (*Transaction, error) {
	var out Transaction
	if err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/transaction/verify/%s", reference), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single transaction by ID.
func (s *TransactionService) Get(ctx context.Context, id string) (*Transaction, error) {
	var out Transaction
	if err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/transaction/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of transactions.
func (s *TransactionService) List(ctx context.Context, opts ListOptions) (*TransactionList, error) {
	var out TransactionList
	path := fmt.Sprintf("/api/v1/transaction?page=%d&per_page=%d", pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Refund marks a successful transaction as reversed.
func (s *TransactionService) Refund(ctx context.Context, id string) (*Transaction, error) {
	var out Transaction
	if err := s.client.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/transaction/%s/refund", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
