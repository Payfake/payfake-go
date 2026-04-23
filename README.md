# payfake-go

Official Go SDK for [Payfake](https://payfake.co) — a Paystack-compatible payment
simulator for African developers.

Build and test your entire Paystack integration without touching real money, a real
phone, or Paystack's sandbox. Change one environment variable to go live.

```bash
go get github.com/payfake/payfake-go
```

---

## Quick Start

```go
import payfake "github.com/payfake/payfake-go"

client := payfake.New(payfake.Config{
    SecretKey: "sk_test_xxx",                   // from api.payfake.co or your local server
    BaseURL:   "https://api.payfake.co",        // or "http://localhost:8080" for self-hosted
})
```

That's it. The same code works against real Paystack — change `BaseURL` and `SecretKey`:

```bash
# Development
PAYSTACK_BASE_URL=https://api.payfake.co
PAYSTACK_SECRET_KEY=sk_test_xxx

# Production
PAYSTACK_BASE_URL=https://api.paystack.co
PAYSTACK_SECRET_KEY=sk_live_xxx
```

---

## Authentication

Three auth schemes across the SDK:

| Namespace | Auth | Used for |
|-----------|------|----------|
| `Transaction`, `Charge`, `Customer` | `sk_test_xxx` secret key | Paystack-compatible server-side calls |
| `Auth`, `Merchant`, `Control` | JWT token (from `Auth.Login`) | Payfake dashboard and control panel |
| `Transaction.PublicFetch`, `Transaction.PublicVerify`, `Charge.Simulate3DS` | None | Checkout page browser calls |

---

## Initialize and Charge — Full Local Card Flow

```go
ctx := context.Background()

// Step 1 — Initialize
tx, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
    Email:    "customer@example.com",
    Amount:   10000,  // GHS 100.00 — amounts in pesewas
    Currency: "GHS",
})
// tx.AuthorizationURL → redirect customer here (or use in checkout app)
// tx.AccessCode → pass to charge

// Step 2 — Charge (local Verve card: 5061xxxxxxxxxxxxxxxx)
step1, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
    Email:      "customer@example.com",
    AccessCode: tx.AccessCode,
    Card: &payfake.CardDetails{
        Number:      "5061000000000000",
        CVV:         "123",
        ExpiryMonth: "12",
        ExpiryYear:  "2026",
    },
})
// step1.Status == "send_pin"

// Step 3 — Submit PIN
step2, err := client.Charge.SubmitPIN(ctx, payfake.SubmitPINInput{
    Reference: tx.Reference,
    PIN:       "1234",
})
// step2.Status == "send_otp"

// Step 4 — Get OTP from logs (no real phone needed)
otpLogs, _ := client.Control.GetOTPLogs(ctx, token, tx.Reference, payfake.ListOptions{})
otp := otpLogs[0].OTPCode

// Step 5 — Submit OTP
step3, err := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
    Reference: tx.Reference,
    OTP:       otp,
})
// step3.Status == "success"

// Step 6 — Verify (always verify before delivering value)
verified, err := client.Transaction.Verify(ctx, tx.Reference)
// verified.Status == "success"
// verified.GatewayResponse == "Approved"
// verified.Authorization.AuthorizationCode → store for recurring charges
```

---

## Charge Flows

### International Card (Visa/Mastercard — 3DS)

```go
step1, _ := client.Charge.Card(ctx, payfake.ChargeCardInput{
    Email:      "customer@example.com",
    AccessCode: tx.AccessCode,
    Card: &payfake.CardDetails{
        Number:      "4111111111111111",  // Visa → 3DS flow
        CVV:         "123",
        ExpiryMonth: "12",
        ExpiryYear:  "2026",
    },
})
// step1.Status == "open_url"
// step1.URL → checkout app navigates here for 3DS

// After customer confirms on 3DS page:
result, _ := client.Charge.Simulate3DS(ctx, tx.Reference)
// result.Status == "success" or "failed"
```

### Mobile Money

```go
step1, _ := client.Charge.MobileMoney(ctx, payfake.ChargeMomoInput{
    Email:      "customer@example.com",
    AccessCode: tx.AccessCode,
    MobileMoney: &payfake.MomoDetails{
        Phone:    "+233241234567",
        Provider: "mtn",  // mtn | vodafone | airteltigo
    },
})
// step1.Status == "send_otp"

// Get OTP (delivered via SMS in production, read from logs in testing)
otpLogs, _ := client.Control.GetOTPLogs(ctx, token, tx.Reference, payfake.ListOptions{})
step2, _ := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
    Reference: tx.Reference,
    OTP:       otpLogs[0].OTPCode,
})
// step2.Status == "pay_offline" — customer must approve USSD prompt

// Poll until resolved (webhook fires in production)
for {
    result, _ := client.Transaction.PublicVerify(ctx, tx.Reference)
    if result.Status == "success" || result.Status == "failed" {
        break
    }
    time.Sleep(3 * time.Second)
}
```

### Bank Transfer

```go
step1, _ := client.Charge.Bank(ctx, payfake.ChargeBankInput{
    Email:      "customer@example.com",
    AccessCode: tx.AccessCode,
    Bank: &payfake.BankDetails{
        Code:          "GCB",
        AccountNumber: "1234567890",
    },
})
// step1.Status == "send_birthday"

step2, _ := client.Charge.SubmitBirthday(ctx, payfake.SubmitBirthdayInput{
    Reference: tx.Reference,
    Birthday:  "1990-01-15",
})
// step2.Status == "send_otp"

// Get OTP and submit
otpLogs, _ := client.Control.GetOTPLogs(ctx, token, tx.Reference, payfake.ListOptions{})
step3, _ := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
    Reference: tx.Reference,
    OTP:       otpLogs[0].OTPCode,
})
// step3.Status == "success"
```

---

## Charge Flow Status Reference

| Status | Meaning | Next Call |
|--------|---------|-----------|
| `send_pin` | Enter card PIN | `Charge.SubmitPIN` |
| `send_otp` | Enter OTP | `Charge.SubmitOTP` |
| `send_birthday` | Enter date of birth | `Charge.SubmitBirthday` |
| `send_address` | Enter billing address | `Charge.SubmitAddress` |
| `open_url` | Complete 3DS — open `URL` field | Navigate checkout to `URL` |
| `pay_offline` | Approve USSD prompt | Poll `Transaction.PublicVerify` |
| `success` | Payment complete | Webhook fired, call `Verify` |
| `failed` | Payment declined | Read `GatewayResponse` |

---

## Scenario Testing

```go
// Login to get JWT for control operations
loginResp, _ := client.Auth.Login(ctx, payfake.LoginInput{
    Email:    "dev@acme.com",
    Password: "secret123",
})
token := loginResp.AccessToken

// Force all charges to fail
status := "failed"
code   := "CHARGE_INSUFFICIENT_FUNDS"
client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
    ForceStatus: &status,
    ErrorCode:   &code,
})

// 30% random failure rate with 2 second delay
rate  := 0.3
delay := 2000
client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
    FailureRate: &rate,
    DelayMS:     &delay,
})

// Simulate MoMo timeout
status = "failed"
code   = "CHARGE_MOMO_TIMEOUT"
client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
    ForceStatus: &status,
    ErrorCode:   &code,
})

// Reset when done — always reset after tests
client.Control.ResetScenario(ctx, token)
```

### Available Error Codes

| Code | Channel | Meaning |
|------|---------|---------|
| `CHARGE_INSUFFICIENT_FUNDS` | Card | Not enough balance |
| `CHARGE_DO_NOT_HONOR` | Card | Bank declined — most common in Ghana |
| `CHARGE_INVALID_PIN` | Card | Wrong PIN |
| `CHARGE_INVALID_OTP` | Card/MoMo/Bank | Wrong or expired OTP |
| `CHARGE_CARD_EXPIRED` | Card | Card past expiry |
| `CHARGE_NOT_PERMITTED` | Card | Online payments disabled |
| `CHARGE_LIMIT_EXCEEDED` | Card | Daily limit reached |
| `CHARGE_MOMO_TIMEOUT` | MoMo | Customer ignored prompt |
| `CHARGE_MOMO_INVALID_NUMBER` | MoMo | Number not on network |
| `CHARGE_MOMO_PROVIDER_UNAVAILABLE` | MoMo | Network down |
| `CHARGE_BANK_INVALID_ACCOUNT` | Bank | Account doesn't exist |
| `CHARGE_BANK_TRANSFER_FAILED` | Bank | Bank rejected transfer |

---

## Error Handling

```go
_, err := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
    Reference: ref,
    OTP:       "000000",
})
if err != nil {
    switch {
    case payfake.IsCode(err, payfake.CodeChargeInvalidOTP):
        // OTP wrong or expired, resend
        client.Charge.ResendOTP(ctx, payfake.ResendOTPInput{Reference: ref})
    case payfake.IsCode(err, payfake.CodeInsufficientFunds):
        // Funds issue
    default:
        log.Printf("unexpected error: %v", err)
    }
}

// Access the full error details
if sdkErr, ok := err.(*payfake.SDKError); ok {
    fmt.Println("Code:       ", sdkErr.Code)
    fmt.Println("Message:    ", sdkErr.Message)
    fmt.Println("HTTP status:", sdkErr.HTTPStatus)
    for _, f := range sdkErr.Fields {
        fmt.Printf("  %s (%s): %s\n", f.Field, f.Rule, f.Message)
    }
}
```

---

## Customers

```go
// Create
customer, _ := client.Customer.Create(ctx, payfake.CreateCustomerInput{
    Email:     "kofi@example.com",
    FirstName: "Kofi",
    LastName:  "Mensah",
    Phone:     "+233241234567",
})

// Fetch by code
customer, _ = client.Customer.Fetch(ctx, customer.CustomerCode)

// Update (pointer fields — nil means don't change)
firstName := "Kwame"
client.Customer.Update(ctx, customer.CustomerCode, payfake.UpdateCustomerInput{
    FirstName: &firstName,
})

// Transactions
txList, _ := client.Customer.Transactions(ctx, customer.CustomerCode, payfake.ListOptions{Page: 1, PerPage: 20})
```

---

## Webhook Verification

Payfake signs every webhook with HMAC-SHA512 identical to Paystack.
Your existing verification code works unchanged.

```go
import (
    "crypto/hmac"
    "crypto/sha512"
    "encoding/hex"
    "net/http"
)

func verifyWebhook(r *http.Request, secretKey string) bool {
    signature := r.Header.Get("X-Paystack-Signature")
    // Read raw body — never the parsed JSON
    body, _ := io.ReadAll(r.Body)
    mac := hmac.New(sha512.New, []byte(secretKey))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

---

## API Reference

Full API documentation at [docs.payfake.co](https://docs.payfake.co).

Payfake is hosted at `https://api.payfake.co` — no account setup or Docker required to try it.

---

## License

MIT
