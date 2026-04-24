package payfake

import "time"

// Shared

// envelope is the Paystack-compatible response wrapper.
// status is a boolean, true on success, false on any failure.
// The SDK unwraps this internally, callers only ever see the typed
// return value or an SDKError.
type envelope struct {
	Status  bool                        `json:"status"`
	Message string                      `json:"message"`
	Data    any                         `json:"data"`
	Errors  map[string][]ValidationRule `json:"errors"`
}

// ValidationRule is a single rule violation inside the errors map.
// Matches Paystack's { "rule": "required", "message": "..." } shape.
type ValidationRule struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// PaginationMeta is the Paystack-compatible pagination metadata
// returned on all list endpoints.
type PaginationMeta struct {
	Total     int64 `json:"total"`
	Skipped   int   `json:"skipped"`
	PerPage   int   `json:"per_page"`
	Page      int   `json:"page"`
	PageCount int   `json:"pageCount"`
}

// ListOptions controls pagination on list endpoints.
type ListOptions struct {
	Page    int
	PerPage int
}

// Auth

type RegisterInput struct {
	BusinessName string `json:"business_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Merchant    MerchantSummary `json:"merchant"`
	AccessToken string          `json:"access_token"`
}

type MerchantSummary struct {
	ID           string `json:"id"`
	BusinessName string `json:"business_name"`
	Email        string `json:"email"`
	PublicKey    string `json:"public_key"`
}

type KeysResponse struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
}

// Transaction

type InitializeInput struct {
	Email       string         `json:"email"`
	Amount      int64          `json:"amount"`
	Currency    string         `json:"currency,omitempty"`
	Reference   string         `json:"reference,omitempty"`
	CallbackURL string         `json:"callback_url,omitempty"`
	Channels    []string       `json:"channels,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// InitializeResponse is the data returned by POST /transaction/initialize.
// Matches Paystack exactly:
//
//	{ "authorization_url": "...", "access_code": "...", "reference": "..." }
type InitializeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

// Transaction is a full transaction record as returned by verify/list/fetch.
// Matches Paystack's transaction object shape including authorization and customer.
type Transaction struct {
	ID              string          `json:"id"`
	Domain          string          `json:"domain"`
	Status          string          `json:"status"`
	Reference       string          `json:"reference"`
	ReceiptNumber   any             `json:"receipt_number"`
	Amount          int64           `json:"amount"`
	Message         any             `json:"message"`
	GatewayResponse string          `json:"gateway_response"`
	PaidAt          *time.Time      `json:"paid_at"`
	CreatedAt       time.Time       `json:"created_at"`
	Channel         string          `json:"channel"`
	Currency        string          `json:"currency"`
	IPAddress       any             `json:"ip_address"`
	Fees            int64           `json:"fees"`
	FeesSplit       any             `json:"fees_split"`
	FeesBreakdown   any             `json:"fees_breakdown"`
	Metadata        map[string]any  `json:"metadata"`
	Log             any             `json:"log"`
	Customer        CustomerSummary `json:"customer"`
	Authorization   *Authorization  `json:"authorization"`
}

// Authorization is the payment instrument object embedded in transactions.
// Store this for recurring charges, the authorization_code charges the card again.
type Authorization struct {
	AuthorizationCode string `json:"authorization_code"`
	Bin               string `json:"bin"`
	Last4             string `json:"last4"`
	ExpMonth          string `json:"exp_month"`
	ExpYear           string `json:"exp_year"`
	Channel           string `json:"channel"`
	CardType          string `json:"card_type"`
	Bank              string `json:"bank"`
	CountryCode       string `json:"country_code"`
	Brand             string `json:"brand"`
	Reusable          bool   `json:"reusable"`
	Signature         string `json:"signature"`
	AccountName       any    `json:"account_name"`
	// MoMo fields
	MobileMoneynumber string `json:"mobile_money_number,omitempty"`
	MobileMoneyName   string `json:"mobile_money_name,omitempty"`
}

// TransactionList is returned by GET /transaction.
type TransactionList struct {
	Data []Transaction  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Charge

// CardDetails is the nested card object for POST /charge.
// Matches Paystack's card shape: { number, cvv, expiry_month, expiry_year }
type CardDetails struct {
	Number      string `json:"number"`
	CVV         string `json:"cvv"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
}

// MomoDetails is the nested mobile_money object for POST /charge.
type MomoDetails struct {
	Phone    string `json:"phone"`
	Provider string `json:"provider"`
}

// BankDetails is the nested bank object for POST /charge.
type BankDetails struct {
	Code          string `json:"code"`
	AccountNumber string `json:"account_number"`
}

// ChargeCardInput is the body for a card charge.
// Channel is inferred from the presence of the card object.
type ChargeCardInput struct {
	Email     string       `json:"email"`
	Amount    int64        `json:"amount,omitempty"`
	Reference string       `json:"reference,omitempty"`
	Card      *CardDetails `json:"card"`
}

// ChargeMomoInput is the body for a mobile money charge.
type ChargeMomoInput struct {
	Email       string       `json:"email"`
	Amount      int64        `json:"amount,omitempty"`
	Reference   string       `json:"reference,omitempty"`
	MobileMoney *MomoDetails `json:"mobile_money"`
}

// ChargeBankInput is the body for a bank transfer charge.
type ChargeBankInput struct {
	Email     string       `json:"email"`
	Amount    int64        `json:"amount,omitempty"`
	Reference string       `json:"reference,omitempty"`
	Bank      *BankDetails `json:"bank"`
	Birthday  string       `json:"birthday,omitempty"`
}

// ChargeFlowResponse is returned by every charge step endpoint.
//
// Read Status to decide what the checkout page renders next:
//
//	"send_pin"      → show PIN input, call SubmitPIN
//	"send_otp"      → show OTP input, call SubmitOTP
//	"send_birthday" → show DOB input, call SubmitBirthday
//	"send_address"  → show address form, call SubmitAddress
//	"open_url"      → open URL field in checkout app 3DS page
//	"pay_offline"   → show approve on phone screen, poll PublicVerify
//	"success"       → payment complete, webhook fired
//	"failed"        → payment declined
type ChargeFlowResponse struct {
	Status      string `json:"status"`
	Reference   string `json:"reference"`
	DisplayText string `json:"display_text"`
	// URL is populated when Status is "open_url" (3DS).
	// Navigate the checkout app to this URL.
	URL             string         `json:"url"`
	Amount          int64          `json:"amount"`
	Currency        string         `json:"currency"`
	Channel         string         `json:"channel"`
	GatewayResponse string         `json:"gateway_response"`
	Authorization   *Authorization `json:"authorization"`
}

type SubmitPINInput struct {
	Reference string `json:"reference"`
	PIN       string `json:"pin"`
}

type SubmitOTPInput struct {
	Reference string `json:"reference"`
	OTP       string `json:"otp"`
}

type SubmitBirthdayInput struct {
	Reference string `json:"reference"`
	// Birthday format: YYYY-MM-DD
	Birthday string `json:"birthday"`
}

type SubmitAddressInput struct {
	Reference string `json:"reference"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	Country   string `json:"country"`
}

type ResendOTPInput struct {
	Reference string `json:"reference"`
}

// OTPLog is a generated OTP stored for developer inspection.
// OTPs are never returned in API responses, read them here during testing.
type OTPLog struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	Reference  string `json:"reference"`
	Channel    string `json:"channel"`
	// OTPCode is the 6-digit OTP, use this in SubmitOTP during testing.
	OTPCode   string    `json:"otp_code"`
	Step      string    `json:"step"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// PublicTransactionResponse is returned by GET /api/v1/public/transaction/:access_code.
// Used by the checkout page on mount.
type PublicTransactionResponse struct {
	Amount      int64          `json:"amount"`
	Currency    string         `json:"currency"`
	Status      string         `json:"status"`
	Reference   string         `json:"reference"`
	CallbackURL string         `json:"callback_url"`
	AccessCode  string         `json:"access_code"`
	Merchant    PublicMerchant `json:"merchant"`
	Customer    PublicCustomer `json:"customer"`
	Charge      *PublicCharge  `json:"charge"`
}

type PublicMerchant struct {
	BusinessName string `json:"business_name"`
	PublicKey    string `json:"public_key"`
}

type PublicCustomer struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// PublicCharge is the charge state embedded in the public transaction response.
// Check FlowStatus during MoMo polling to know where in the flow things are.
type PublicCharge struct {
	FlowStatus string `json:"flow_status"`
	Status     string `json:"status"`
	ErrorCode  string `json:"error_code"`
	Channel    string `json:"channel"`
}

// PublicVerifyResponse is returned by GET /api/v1/public/transaction/verify/:reference.
// Used for MoMo polling, check Status and Charge.FlowStatus each tick.
type PublicVerifyResponse struct {
	Status    string        `json:"status"`
	Reference string        `json:"reference"`
	Amount    int64         `json:"amount"`
	Currency  string        `json:"currency"`
	PaidAt    *time.Time    `json:"paid_at"`
	Charge    *PublicCharge `json:"charge"`
}

// Customer

type CreateCustomerInput struct {
	Email     string         `json:"email"`
	FirstName string         `json:"first_name,omitempty"`
	LastName  string         `json:"last_name,omitempty"`
	Phone     string         `json:"phone,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type UpdateCustomerInput struct {
	FirstName *string        `json:"first_name,omitempty"`
	LastName  *string        `json:"last_name,omitempty"`
	Phone     *string        `json:"phone,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Customer is the full customer object matching Paystack's shape.
type Customer struct {
	ID           string         `json:"id"`
	CustomerCode string         `json:"customer_code"`
	Email        string         `json:"email"`
	FirstName    string         `json:"first_name"`
	LastName     string         `json:"last_name"`
	Phone        string         `json:"phone"`
	Metadata     map[string]any `json:"metadata"`
	Domain       string         `json:"domain"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	Integration  int            `json:"integration"`
}

// CustomerSummary is the condensed customer object embedded in transactions.
type CustomerSummary struct {
	ID           string `json:"id"`
	CustomerCode string `json:"customer_code"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
}

// CustomerList is returned by GET /customer.
type CustomerList struct {
	Data []Customer     `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Merchant

type MerchantProfile struct {
	ID           string    `json:"id"`
	BusinessName string    `json:"business_name"`
	Email        string    `json:"email"`
	PublicKey    string    `json:"public_key"`
	WebhookURL   string    `json:"webhook_url"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateProfileInput struct {
	BusinessName string `json:"business_name,omitempty"`
	WebhookURL   string `json:"webhook_url,omitempty"`
}

type WebhookConfig struct {
	WebhookURL string `json:"webhook_url"`
	IsSet      bool   `json:"is_set"`
}

type WebhookTestResult struct {
	WebhookURL string         `json:"webhook_url"`
	Success    bool           `json:"success"`
	StatusCode int            `json:"status_code"`
	Payload    map[string]any `json:"payload"`
}

// Control

type ScenarioConfig struct {
	ID          string  `json:"id"`
	MerchantID  string  `json:"merchant_id"`
	FailureRate float64 `json:"failure_rate"`
	DelayMS     int     `json:"delay_ms"`
	ForceStatus string  `json:"force_status"`
	ErrorCode   string  `json:"error_code"`
}

type UpdateScenarioInput struct {
	FailureRate *float64 `json:"failure_rate,omitempty"`
	DelayMS     *int     `json:"delay_ms,omitempty"`
	ForceStatus *string  `json:"force_status,omitempty"`
	ErrorCode   *string  `json:"error_code,omitempty"`
}

type ForceTransactionInput struct {
	Status    string `json:"status"`
	ErrorCode string `json:"error_code,omitempty"`
}

type WebhookEvent struct {
	ID            string         `json:"id"`
	Event         string         `json:"event"`
	TransactionID string         `json:"transaction_id"`
	Delivered     bool           `json:"delivered"`
	Attempts      int            `json:"attempts"`
	LastAttemptAt *time.Time     `json:"last_attempt_at"`
	Payload       map[string]any `json:"payload"`
	CreatedAt     time.Time      `json:"created_at"`
}

type WebhookAttempt struct {
	ID           string    `json:"id"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body"`
	Succeeded    bool      `json:"succeeded"`
	AttemptedAt  time.Time `json:"attempted_at"`
}

type WebhookList struct {
	Data []WebhookEvent `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type RequestLog struct {
	ID           string    `json:"id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	RequestBody  string    `json:"request_body"`
	ResponseBody string    `json:"response_body"`
	DurationMS   int64     `json:"duration_ms"`
	RequestID    string    `json:"request_id"`
	LoggedAt     time.Time `json:"logged_at"`
}

type LogList struct {
	Data []RequestLog   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type OTPLogList struct {
	Data []OTPLog       `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type MerchantStats struct {
	Transactions struct {
		Total       int64   `json:"total"`
		Successful  int64   `json:"successful"`
		Failed      int64   `json:"failed"`
		Pending     int64   `json:"pending"`
		Abandoned   int64   `json:"abandoned"`
		SuccessRate float64 `json:"success_rate"`
	} `json:"transactions"`
	Volume struct {
		TotalAmount int64 `json:"total_amount"`
	} `json:"volume"`
	Customers struct {
		Total int64 `json:"total"`
	} `json:"customers"`
	Webhooks struct {
		Total     int64 `json:"total"`
		Delivered int64 `json:"delivered"`
		Failed    int64 `json:"failed"`
	} `json:"webhooks"`
	DailyActivity []DailyActivity `json:"daily_activity"`
}

type DailyActivity struct {
	Date   string `json:"date"`
	Count  int64  `json:"count"`
	Volume int64  `json:"volume"`
}
