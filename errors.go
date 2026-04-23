package payfake

import (
	"fmt"
	"strings"
)

// SDKError is returned by every SDK method on failure.
// Switch on Code for programmatic error handling, never parse Message.
type SDKError struct {
	// Code is the Payfake response code from X-Payfake-Code header.
	// e.g. AUTH_EMAIL_TAKEN, CHARGE_INSUFFICIENT_FUNDS
	Code string
	// Message is the human-readable error from the API.
	Message string
	// Fields contains field-level validation errors when Code is VALIDATION_ERROR.
	Fields []ErrorField
	// HTTPStatus is the HTTP status code of the failed response.
	HTTPStatus int
}

func (e *SDKError) Error() string {
	if len(e.Fields) > 0 {
		parts := make([]string, 0, len(e.Fields))
		for _, f := range e.Fields {
			parts = append(parts, fmt.Sprintf("%s: %s", f.Field, f.Message))
		}
		return fmt.Sprintf("payfake [%s] %s — %s", e.Code, e.Message, strings.Join(parts, ", "))
	}
	return fmt.Sprintf("payfake [%s] %s", e.Code, e.Message)
}

// ErrorField is a single field-level validation error.
type ErrorField struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// IsCode reports whether err is an SDKError with the given code.
// Use this instead of string-matching on Error().
//
//	_, err := client.Charge.Card(ctx, input)
//	if payfake.IsCode(err, payfake.CodeInsufficientFunds) {
//	    // handle specific failure
//	}
func IsCode(err error, code string) bool {
	if err == nil {
		return false
	}
	sdkErr, ok := err.(*SDKError)
	if !ok {
		return false
	}
	return sdkErr.Code == code
}

// Response code constants, use these instead of raw strings.
const (
	// Auth
	CodeEmailTaken         = "AUTH_EMAIL_TAKEN"
	CodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeUnauthorized       = "AUTH_UNAUTHORIZED"
	CodeTokenExpired       = "AUTH_TOKEN_EXPIRED"
	CodeTokenInvalid       = "AUTH_TOKEN_INVALID"

	// Transaction
	CodeTransactionNotFound    = "TRANSACTION_NOT_FOUND"
	CodeReferenceTaken         = "TRANSACTION_REFERENCE_TAKEN"
	CodeInvalidAmount          = "TRANSACTION_INVALID_AMOUNT"
	CodeTransactionAlreadyDone = "TRANSACTION_ALREADY_VERIFIED"

	// Charge
	CodeChargeFailed       = "CHARGE_FAILED"
	CodeChargeSuccessful   = "CHARGE_SUCCESSFUL"
	CodeChargeSendPIN      = "CHARGE_SEND_PIN"
	CodeChargeSendOTP      = "CHARGE_SEND_OTP"
	CodeChargeSendBirthday = "CHARGE_SEND_BIRTHDAY"
	CodeChargeSendAddress  = "CHARGE_SEND_ADDRESS"
	CodeChargeOpenURL      = "CHARGE_OPEN_URL"
	CodeChargePayOffline   = "CHARGE_PAY_OFFLINE"
	CodeChargeInvalidOTP   = "CHARGE_INVALID_OTP"
	CodeInsufficientFunds  = "CHARGE_INSUFFICIENT_FUNDS"
	CodeDoNotHonor         = "CHARGE_DO_NOT_HONOR"
	CodeMomoTimeout        = "CHARGE_MOMO_TIMEOUT"

	// Customer
	CodeCustomerNotFound   = "CUSTOMER_NOT_FOUND"
	CodeCustomerEmailTaken = "CUSTOMER_EMAIL_TAKEN"

	// Generic
	CodeValidationError = "VALIDATION_ERROR"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeRateLimited     = "RATE_LIMIT_EXCEEDED"
)
