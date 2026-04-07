package payfake

import (
	"context"
	"fmt"
	"net/http"
)

type CustomerService struct {
	client *Client
}

// Create creates a new customer under the merchant account.
func (s *CustomerService) Create(ctx context.Context, input CreateCustomerInput) (*Customer, error) {
	var out Customer
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/customer", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of customers.
func (s *CustomerService) List(ctx context.Context, opts ListOptions) (*CustomerList, error) {
	var out CustomerList
	path := fmt.Sprintf("/api/v1/customer?page=%d&per_page=%d", pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a customer by their code (CUS_xxxxxxxx).
func (s *CustomerService) Get(ctx context.Context, code string) (*Customer, error) {
	var out Customer
	if err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/customer/%s", code), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update partially updates a customer record.
// Only non-nil fields are updated.
func (s *CustomerService) Update(ctx context.Context, code string, input UpdateCustomerInput) (*Customer, error) {
	var out Customer
	if err := s.client.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/customer/%s", code), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Transactions returns paginated transactions for a customer.
func (s *CustomerService) Transactions(ctx context.Context, code string, opts ListOptions) (*TransactionList, error) {
	var out TransactionList
	path := fmt.Sprintf("/api/v1/customer/%s/transactions?page=%d&per_page=%d",
		code, pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
