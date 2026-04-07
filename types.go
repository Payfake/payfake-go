package payfake

import "time"

// Shared

// Response is the base envelope every Payfake API response follows.
// The SDK unwraps this internally, callers never see the envelope,
// they just get the typed Data or a clean error.
type apiResponse struct {
	Status   string          `json:"status"`
	Message  string          `json:"message"`
	Data     any             `json:"data"`
	Errors   []APIErrorField `json:"errors"`
	Metadata struct {
		Timestamp string `json:"timestamp"`
		RequestID string `json:"request_id"`
	} `json:"metadata"`
	Code string `json:"code"`
}

// APIErrorField is a single field-level error returned by the API.
type APIErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Auth

type RegisterInput struct {
	BusinessName string `json:"business_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type RegisterResponse struct {
	Merchant MerchantData `json:"merchant"`
	Token    string       `json:"token"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Merchant MerchantData `json:"merchant"`
	Token    string       `json:"token"`
}

type MerchantData struct {
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

type InitializeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

type Transaction struct {
	ID          string         `json:"id"`
	Reference   string         `json:"reference"`
	Amount      int64          `json:"amount"`
	Currency    string         `json:"currency"`
	Status      string         `json:"status"`
	Channel     string         `json:"channel"`
	Fees        int64          `json:"fees"`
	AccessCode  string         `json:"access_code"`
	CallbackURL string         `json:"callback_url"`
	PaidAt      *time.Time     `json:"paid_at"`
	CreatedAt   time.Time      `json:"created_at"`
	Customer    Customer       `json:"customer"`
	Metadata    map[string]any `json:"metadata"`
}

type TransactionList struct {
	Transactions []Transaction  `json:"transactions"`
	Meta         PaginationMeta `json:"meta"`
}

// Charge

type ChargeCardInput struct {
	AccessCode string `json:"access_code,omitempty"`
	Reference  string `json:"reference,omitempty"`
	CardNumber string `json:"card_number"`
	CardExpiry string `json:"card_expiry"`
	CVV        string `json:"cvv"`
	Email      string `json:"email"`
}

type ChargeMomoInput struct {
	AccessCode string `json:"access_code,omitempty"`
	Reference  string `json:"reference,omitempty"`
	Phone      string `json:"phone"`
	Provider   string `json:"provider"`
	Email      string `json:"email"`
}

type ChargeBankInput struct {
	AccessCode    string `json:"access_code,omitempty"`
	Reference     string `json:"reference,omitempty"`
	BankCode      string `json:"bank_code"`
	AccountNumber string `json:"account_number"`
	Email         string `json:"email"`
}

type ChargeResponse struct {
	Transaction Transaction `json:"transaction"`
	Charge      ChargeData  `json:"charge"`
}

type ChargeData struct {
	ID           string `json:"id"`
	Channel      string `json:"channel"`
	Status       string `json:"status"`
	CardBrand    string `json:"card_brand"`
	CardLast4    string `json:"card_last4"`
	MomoPhone    string `json:"momo_phone"`
	MomoProvider string `json:"momo_provider"`
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

type Customer struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Phone     string         `json:"phone"`
	Code      string         `json:"customer_code"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

type CustomerList struct {
	Customers []Customer     `json:"customers"`
	Meta      PaginationMeta `json:"meta"`
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
	ID            string     `json:"id"`
	Event         string     `json:"event"`
	TransactionID string     `json:"transaction_id"`
	Delivered     bool       `json:"delivered"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type WebhookAttempt struct {
	ID           string    `json:"id"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body"`
	Succeeded    bool      `json:"succeeded"`
	AttemptedAt  time.Time `json:"attempted_at"`
}

type RequestLog struct {
	ID           string    `json:"id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	RequestBody  string    `json:"request_body"`
	ResponseBody string    `json:"response_body"`
	Duration     int64     `json:"duration_ms"`
	RequestID    string    `json:"request_id"`
	LoggedAt     time.Time `json:"logged_at"`
}

// Shared

type PaginationMeta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int64 `json:"pages"`
}

type ListOptions struct {
	Page    int
	PerPage int
}
