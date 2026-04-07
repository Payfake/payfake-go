package payfake

import (
	"fmt"
	"strings"
)

// SDKError is returned by every SDK method on failure.
// It carries the API response code, message, and field-level errors
// so callers can handle specific failure cases programmatically
// without parsing error strings.
type SDKError struct {
	// Code is the Payfake response code, e.g. AUTH_EMAIL_TAKEN.
	// Switch on this for programmatic error handling.
	Code string
	// Message is the human-readable error from the API.
	Message string
	// Fields contains field-level validation errors if any.
	// Empty for non-validation errors.
	Fields []APIErrorField
	// HTTPStatus is the HTTP status code of the failed response.
	HTTPStatus int
}

func (e *SDKError) Error() string {
	if len(e.Fields) > 0 {
		fieldMessages := make([]string, 0, len(e.Fields))
		for _, f := range e.Fields {
			fieldMessages = append(fieldMessages, fmt.Sprintf("%s: %s", f.Field, f.Message))
		}
		return fmt.Sprintf("payfake [%s] %s — %s", e.Code, e.Message, strings.Join(fieldMessages, ", "))
	}
	return fmt.Sprintf("payfake [%s] %s", e.Code, e.Message)
}

// IsCode checks if an error matches a specific Payfake response code.
// Use this instead of string comparison on the error message.
//
// Example:
//
//	_, err := client.Auth.Register(ctx, input)
//	if payfake.IsCode(err, "AUTH_EMAIL_TAKEN") {
//	    // handle duplicate email
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

// Common code constants so callers don't use raw strings.
const (
	CodeEmailTaken            = "AUTH_EMAIL_TAKEN"
	CodeInvalidCredentials    = "AUTH_INVALID_CREDENTIALS"
	CodeUnauthorized          = "AUTH_UNAUTHORIZED"
	CodeTokenExpired          = "AUTH_TOKEN_EXPIRED"
	CodeTransactionNotFound   = "TRANSACTION_NOT_FOUND"
	CodeReferenceTaken        = "TRANSACTION_REFERENCE_TAKEN"
	CodeInvalidAmount         = "TRANSACTION_INVALID_AMOUNT"
	CodeChargeFailed          = "CHARGE_FAILED"
	CodeChargePending         = "CHARGE_PENDING"
	CodeCustomerNotFound      = "CUSTOMER_NOT_FOUND"
	CodeCustomerEmailTaken    = "CUSTOMER_EMAIL_TAKEN"
	CodeScenarioInvalidConfig = "SCENARIO_INVALID_CONFIG"
	CodeValidationError       = "VALIDATION_ERROR"
	CodeInternalError         = "INTERNAL_ERROR"
)
