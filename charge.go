package payfake

import (
	"context"
	"net/http"
)

type ChargeService struct {
	client *Client
}

// Card charges a card directly.
// Returns the transaction and charge on success.
// Returns an SDKError with code CHARGE_FAILED and a specific
// error code in Fields on simulated failure.
func (s *ChargeService) Card(ctx context.Context, input ChargeCardInput) (*ChargeResponse, error) {
	var out ChargeResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/card", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MobileMoney initiates a mobile money charge.
// Returns immediately with status "pending", the final outcome
// arrives via webhook after the simulated approval window.
// Always implement webhook handling for MoMo, never assume
// a pending response means success.
func (s *ChargeService) MobileMoney(ctx context.Context, input ChargeMomoInput) (*ChargeResponse, error) {
	var out ChargeResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/mobile_money", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Bank initiates a bank transfer charge.
func (s *ChargeService) Bank(ctx context.Context, input ChargeBankInput) (*ChargeResponse, error) {
	var out ChargeResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/bank", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a charge by transaction reference.
func (s *ChargeService) Fetch(ctx context.Context, reference string) (*ChargeData, error) {
	var out ChargeData
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/charge/"+reference, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
