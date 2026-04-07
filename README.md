# payfake-go

Official Go SDK for [Payfake API](https://github.com/payfake/payfake-api) — a self-hostable African payment simulator that mirrors the Paystack API exactly. Test every payment scenario without touching real money.

## Installation

```bash
go get github.com/payfake/payfake-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    payfake "github.com/payfake/payfake-go"
)

func main() {
    client := payfake.New(payfake.Config{
        SecretKey: "sk_test_xxx",
        BaseURL:   "http://localhost:8080", // your Payfake server
    })

    ctx := context.Background()

    // Initialize a transaction
    tx, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
        Email:    "customer@example.com",
        Amount:   10000, // GHS 100.00 — amounts in smallest unit (pesewas)
        Currency: "GHS",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Access code:", tx.AccessCode)
    fmt.Println("Auth URL:", tx.AuthorizationURL)

    // Charge a card
    charge, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
        AccessCode: tx.AccessCode,
        CardNumber: "4111111111111111",
        CardExpiry: "12/26",
        CardCVV:    "123",
        Email:      "customer@example.com",
    })
    if err != nil {
        if payfake.IsCode(err, payfake.CodeChargeFailed) {
            fmt.Println("Charge failed — check scenario config")
            return
        }
        log.Fatal(err)
    }

    fmt.Println("Status:", charge.Transaction.Status)

    // Verify the transaction
    verified, err := client.Transaction.Verify(ctx, tx.Reference)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Verified:", verified.Status)
}
```

## Namespaces

| Namespace | Access | Description |
|-----------|--------|-------------|
| `client.Auth` | Public + JWT | Register, login, key management |
| `client.Transaction` | Secret key | Initialize, verify, list, refund |
| `client.Charge` | Secret key | Card, mobile money, bank transfer |
| `client.Customer` | Secret key | Create, list, fetch, update |
| `client.Control` | JWT | Scenarios, webhooks, logs, force outcomes |

## Error Handling

Every failed API call returns an `*SDKError`. Use `IsCode()` for programmatic handling:

```go
_, err := client.Transaction.Initialize(ctx, input)
if err != nil {
    if payfake.IsCode(err, payfake.CodeReferenceTaken) {
        // duplicate reference, verify the existing transaction instead
    }
    if payfake.IsCode(err, payfake.CodeInvalidAmount) {
        // amount is zero or negative
    }
    // fallback
    log.Fatal(err)
}
```

Available error code constants:

```go
payfake.CodeEmailTaken
payfake.CodeInvalidCredentials
payfake.CodeUnauthorized
payfake.CodeTokenExpired
payfake.CodeTransactionNotFound
payfake.CodeReferenceTaken
payfake.CodeInvalidAmount
payfake.CodeChargeFailed
payfake.CodeChargePending
payfake.CodeCustomerNotFound
payfake.CodeCustomerEmailTaken
payfake.CodeValidationError
payfake.CodeInternalError
```

## Scenario Control

Use the control namespace to configure simulation behavior:

```go
// Login first to get a JWT
login, _ := client.Auth.Login(ctx, payfake.LoginInput{
    Email:    "dev@acme.com",
    Password: "secret123",
})
token := login.Token

// Set 30% failure rate with 1 second delay
failureRate := 0.3
delayMs := 1000
client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
    FailureRate: &failureRate,
    DelayMS:     &delayMs,
})

// Force a specific transaction to fail
client.Control.ForceTransaction(ctx, token, reference, payfake.ForceTransactionInput{
    Status:    "failed",
    ErrorCode: "CHARGE_INSUFFICIENT_FUNDS",
})

// Reset everything back to defaults
client.Control.ResetScenario(ctx, token)
```

## Mobile Money

MoMo charges are async, they always return `pending` immediately.
The final outcome arrives via webhook after the simulated delay:

```go
charge, err := client.Charge.MobileMoney(ctx, payfake.ChargeMomoInput{
    AccessCode: tx.AccessCode,
    Phone:      "+233241234567",
    Provider:   "mtn", // mtn | vodafone | airteltigo
    Email:      "customer@example.com",
})

// charge.Transaction.Status is always "pending" here
// implement a webhook handler for the final outcome
```

## Requirements

- Go 1.21+
- A running [Payfake](https://github.com/payfake/payfake-api) server

## License

MIT
