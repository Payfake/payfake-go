package payfake

import (
	"context"
	"fmt"
	"net/http"
)

type ChargeService struct {
	client *Client
}

// Card initiates a card charge.
// Returns ChargeFlowResponse with Status "send_pin" for local Verve cards
// or "open_url" for international Visa/Mastercard.
// For "open_url" open ThreeDSURL in the checkout app's 3DS page.
func (s *ChargeService) Card(ctx context.Context, input ChargeCardInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/card", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MobileMoney initiates a mobile money charge.
// Always returns Status "send_otp", customer must verify phone with OTP.
// After OTP, returns "pay_offline" while waiting for USSD approval.
// Final outcome arrives via webhook, poll transaction status.
func (s *ChargeService) MobileMoney(ctx context.Context, input ChargeMomoInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/mobile_money", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Bank initiates a bank transfer charge.
// Returns Status "send_birthday", customer must enter date of birth first.
func (s *ChargeService) Bank(ctx context.Context, input ChargeBankInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/bank", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitPIN submits the card PIN after a "send_pin" response.
// Returns "send_otp" on success, OTP sent to registered phone.
// Read the OTP from GET /control/otp-logs?reference=xxx during testing.
func (s *ChargeService) SubmitPIN(ctx context.Context, input SubmitPINInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/submit_pin", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitOTP submits the OTP after a "send_otp" response.
// For card charges: returns "success" or "failed".
// For MoMo charges: returns "pay_offline", then poll for final outcome.
// For bank charges: returns "success" or "failed".
func (s *ChargeService) SubmitOTP(ctx context.Context, input SubmitOTPInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/submit_otp", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitBirthday submits the date of birth after a "send_birthday" response.
// Returns "send_otp" on success.
func (s *ChargeService) SubmitBirthday(ctx context.Context, input SubmitBirthdayInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/submit_birthday", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitAddress submits the billing address after a "send_address" response.
// Returns "success" or "failed".
func (s *ChargeService) SubmitAddress(ctx context.Context, input SubmitAddressInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/submit_address", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResendOTP requests a new OTP when the customer hasn't received one.
// Invalidates the previous OTP and generates a fresh one.
// Returns "send_otp", read the new OTP from /control/otp-logs.
func (s *ChargeService) ResendOTP(ctx context.Context, input ResendOTPInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge/resend_otp", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Simulate3DS completes the simulated 3DS verification.
// Called after the customer confirms on the checkout app's 3DS page.
// Returns "success" or "failed" based on the scenario config.
func (s *ChargeService) Simulate3DS(ctx context.Context, reference string) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	path := fmt.Sprintf("/api/v1/public/simulate/3ds/%s", reference)
	if err := s.client.do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves the current state of a charge by transaction reference.
func (s *ChargeService) Fetch(ctx context.Context, reference string) (*ChargeData, error) {
	var out ChargeData
	if err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/charge/%s", reference), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
