package payfake

import (
	"context"
	"fmt"
	"net/http"
)

// CustomerNamespace wraps /customer endpoints.
// These match https://api.paystack.co/customer exactly.
// Auth: Bearer sk_test_xxx
type CustomerNamespace struct {
	client *Client
}

// Create creates a new customer.
func (n *CustomerNamespace) Create(ctx context.Context, input CreateCustomerInput) (*Customer, error) {
	var out Customer
	if err := n.client.do(ctx, http.MethodPost, "/customer", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of customers.
func (n *CustomerNamespace) List(ctx context.Context, opts ListOptions) (*CustomerList, error) {
	path := fmt.Sprintf("/customer?page=%d&perPage=%d",
		pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	var out CustomerList
	if err := n.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a customer by their code (CUS_xxxxxxxx).
func (n *CustomerNamespace) Fetch(ctx context.Context, code string) (*Customer, error) {
	var out Customer
	if err := n.client.do(ctx, http.MethodGet, "/customer/"+code, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update partially updates a customer. Only non-nil pointer fields are updated.
func (n *CustomerNamespace) Update(ctx context.Context, code string, input UpdateCustomerInput) (*Customer, error) {
	var out Customer
	if err := n.client.do(ctx, http.MethodPut, "/customer/"+code, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Transactions returns paginated transactions for a customer.
func (n *CustomerNamespace) Transactions(ctx context.Context, code string, opts ListOptions) (*TransactionList, error) {
	path := fmt.Sprintf("/customer/%s/transactions?page=%d&perPage=%d",
		code, pageOrDefault(opts.Page), perPageOrDefault(opts.PerPage))
	var out TransactionList
	if err := n.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
