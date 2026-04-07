package main

import (
	"context"
	"fmt"
	"log"

	payfake "github.com/payfake/payfake-go"
)

func main() {
	// Initialize the client with your secret key.
	// Point BaseURL at wherever your Payfake server is running.
	client := payfake.New(payfake.Config{
		SecretKey: "sk_test_your_key_here",
		BaseURL:   "http://localhost:8080",
	})

	ctx := context.Background()

	//  Initialize a transaction
	tx, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
		Email:    "customer@example.com",
		Amount:   10000, // GHS 100.00 — amounts are in the smallest unit (pesewas)
		Currency: "GHS",
	})
	if err != nil {
		log.Fatalf("initialize failed: %v", err)
	}

	fmt.Println("Authorization URL:", tx.AuthorizationURL)
	fmt.Println("Access Code:", tx.AccessCode)
	fmt.Println("Reference:", tx.Reference)

	//  Charge a card
	charge, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
		AccessCode: tx.AccessCode,
		CardNumber: "4111111111111111",
		CardExpiry: "12/26",
		CVV:        "123",
		Email:      "customer@example.com",
	})
	if err != nil {
		// Check for specific error codes to handle them gracefully.
		if payfake.IsCode(err, payfake.CodeChargeFailed) {
			fmt.Println("Charge failed — check your scenario config")
			return
		}
		log.Fatalf("charge failed: %v", err)
	}

	fmt.Println("Charge status:", charge.Transaction.Status)

	// Verify the transaction
	verified, err := client.Transaction.Verify(ctx, tx.Reference)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	fmt.Println("Transaction status:", verified.Status)

	//  Simulate a failure scenario
	// First login to get a JWT for control operations.
	loginResp, err := client.Auth.Login(ctx, payfake.LoginInput{
		Email:    "dev@acme.com",
		Password: "secret123",
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	token := loginResp.Token

	// Set failure rate to 100% — every charge will fail.
	failureRate := 1.0
	_, err = client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
		FailureRate: &failureRate,
	})
	if err != nil {
		log.Fatalf("scenario update failed: %v", err)
	}

	fmt.Println("Scenario updated — all charges will now fail")

	// Reset back to defaults when done testing.
	_, err = client.Control.ResetScenario(ctx, token)
	if err != nil {
		log.Fatalf("scenario reset failed: %v", err)
	}

	fmt.Println("Scenario reset — all charges will succeed again")
}
