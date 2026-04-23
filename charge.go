package payfake

import (
	"context"
	"fmt"
	"net/http"
)

// ChargeNamespace wraps /charge endpoints.
// These match https://api.paystack.co/charge exactly.
// All methods call POST /charge, channel is detected from the body object.
// Auth: Bearer sk_test_xxx
type ChargeNamespace struct {
	client *Client
}

// Card initiates a card charge via POST /charge.
//
// Local Ghana cards (Verve: 5061, 5062, 5063, 6500, 6501):
//
//	returns Status "send_pin" → call SubmitPIN
//
// International cards (Visa 4xxx, Mastercard 5xxx):
//
//	returns Status "open_url" + URL → checkout navigates to URL
func (n *ChargeNamespace) Card(ctx context.Context, input ChargeCardInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MobileMoney initiates a MoMo charge via POST /charge.
// Returns Status "send_otp" → call SubmitOTP.
// After OTP returns "pay_offline" → poll Transaction.PublicVerify.
// Final outcome arrives via webhook.
func (n *ChargeNamespace) MobileMoney(ctx context.Context, input ChargeMomoInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Bank initiates a bank transfer charge via POST /charge.
// Returns Status "send_birthday" → call SubmitBirthday.
func (n *ChargeNamespace) Bank(ctx context.Context, input ChargeBankInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitPIN submits the card PIN after Status "send_pin".
// Returns Status "send_otp", OTP sent to registered phone.
// Read the OTP from Control.GetOTPLogs during testing (no real phone needed).
func (n *ChargeNamespace) SubmitPIN(ctx context.Context, input SubmitPINInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge/submit_pin", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitOTP submits the OTP after Status "send_otp".
//
// Card/bank: returns "success" or "failed".
// MoMo: returns "pay_offline", poll Transaction.PublicVerify for final outcome.
func (n *ChargeNamespace) SubmitOTP(ctx context.Context, input SubmitOTPInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge/submit_otp", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitBirthday submits the date of birth after Status "send_birthday".
// Returns Status "send_otp" on success.
func (n *ChargeNamespace) SubmitBirthday(ctx context.Context, input SubmitBirthdayInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge/submit_birthday", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitAddress submits billing address after Status "send_address".
// Returns "success" or "failed".
func (n *ChargeNamespace) SubmitAddress(ctx context.Context, input SubmitAddressInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge/submit_address", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResendOTP requests a fresh OTP when the customer hasn't received one.
// Invalidates the previous OTP. Returns Status "send_otp".
// Read the new OTP from Control.GetOTPLogs.
func (n *ChargeNamespace) ResendOTP(ctx context.Context, input ResendOTPInput) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodPost, "/charge/resend_otp", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves the current state of a charge by transaction reference.
func (n *ChargeNamespace) Fetch(ctx context.Context, reference string) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	if err := n.client.do(ctx, http.MethodGet, "/charge/"+reference, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Simulate3DS completes the simulated 3DS verification.
// Called by the checkout app after the customer confirms on the 3DS page.
// Returns "success" or "failed" based on the scenario config.
func (n *ChargeNamespace) Simulate3DS(ctx context.Context, reference string) (*ChargeFlowResponse, error) {
	var out ChargeFlowResponse
	path := fmt.Sprintf("/api/v1/public/simulate/3ds/%s", reference)
	if err := n.client.doPublic(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
