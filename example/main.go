package main

import (
	"context"
	"fmt"
	"log"
	"time"

	payfake "github.com/payfake/payfake-go"
)

func main() {
	// Setup
	// Point at the hosted instance or your local server.
	// To use your own server: BaseURL: "http://localhost:8080"
	client := payfake.New(payfake.Config{
		SecretKey: "sk_test_your_key_here",
		BaseURL:   "http://localhost:8080",
	})

	ctx := context.Background()

	// Register (first time only)
	authResp, err := client.Auth.Register(ctx, payfake.RegisterInput{
		BusinessName: "Acme Store",
		Email:        "dev@acme.com",
		Password:     "secret123",
	})
	if err != nil {
		// Check for specific error codes
		if payfake.IsCode(err, payfake.CodeEmailTaken) {
			log.Println("Email already registered — logging in instead")
		} else {
			log.Fatalf("register failed: %v", err)
		}
	} else {
		fmt.Println("Registered:", authResp.Merchant.ID)
	}

	// Login
	loginResp, err := client.Auth.Login(ctx, payfake.LoginInput{
		Email:    "dev@acme.com",
		Password: "secret123",
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	token := loginResp.AccessToken
	fmt.Println("Token:", token[:20]+"...")

	// Get keys — the secret key goes in your app's .env
	keys, err := client.Auth.GetKeys(ctx, token)
	if err != nil {
		log.Fatalf("get keys failed: %v", err)
	}
	fmt.Println("Secret key:", keys.SecretKey[:20]+"...")

	// Configure the client with the secret key
	client = payfake.New(payfake.Config{
		SecretKey: keys.SecretKey,
		BaseURL:   "http://localhost:8080", //https://api.payfake.co
	})

	// Initialize a transaction
	tx, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
		Email:    "customer@example.com",
		Amount:   10000, // GHS 100.00 — amounts are in the smallest unit (pesewas)
		Currency: "GHS",
		// CallbackURL: "http://localhost:5173",
	})
	if err != nil {
		log.Fatalf("initialize failed: %v", err)
	}
	fmt.Println("Reference:         ", tx.Reference)
	fmt.Println("Authorization URL: ", tx.AuthorizationURL)
	fmt.Println("Access code:       ", tx.AccessCode)

	//  Full local Verve card flow
	fmt.Println("\n── Card flow (local Verve) ──")

	// // Step 1: Initiate
	// // 5061xxxxxx = local Ghana Verve card → send_pin
	// // 4111xxxxxx = international Visa card → open_url (3DS)
	step1, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
		Email:     "customer@example.com",
		Reference: tx.Reference,
		Card: &payfake.CardDetails{
			Number:      "5061000000000000",
			CVV:         "123",
			ExpiryMonth: "12",
			ExpiryYear:  "2026",
		},
	})
	if err != nil {
		log.Fatalf("charge card failed: %v", err)
	}
	fmt.Println("Step 1 status:", step1.Status) // "send_pin"

	// Step 2: Submit PIN (any 4-digit PIN is accepted)
	step2, err := client.Charge.SubmitPIN(ctx, payfake.SubmitPINInput{
		Reference: tx.Reference,
		PIN:       "1234",
	})
	if err != nil {
		log.Fatalf("submit PIN failed: %v", err)
	}
	fmt.Println("Step 2 status:", step2.Status) // "send_otp"

	// Step 3: Get OTP from logs — no real phone needed during testing
	otpLogs, err := client.Control.GetOTPLogs(ctx, token, tx.Reference, payfake.ListOptions{})
	if err != nil {
		log.Fatalf("get OTP logs failed: %v", err)
	}
	if len(otpLogs) == 0 {
		log.Fatal("no OTP logs found")
	}
	otpCode := otpLogs[0].OTPCode
	fmt.Println("OTP from logs:    ", otpCode)

	// Step 4: Submit OTP
	step3, err := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
		Reference: tx.Reference,
		OTP:       otpCode,
	})
	if err != nil {
		// Handle specific charge failures
		if payfake.IsCode(err, payfake.CodeChargeInvalidOTP) {
			log.Fatal("OTP expired or invalid — call ResendOTP")
		}
		log.Fatalf("submit OTP failed: %v", err)
	}
	fmt.Println("Step 3 status:", step3.Status) // "success"

	// Verify
	verified, err := client.Transaction.Verify(ctx, tx.Reference)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}
	fmt.Println("\nVerified status:         ", verified.Status)
	fmt.Println("Gateway response:        ", verified.GatewayResponse)
	fmt.Println("Authorization code:      ", verified.Authorization.AuthorizationCode)

	//  MoMo flow
	fmt.Println("\n── MoMo flow ──")

	tx2, _ := client.Transaction.Initialize(ctx, payfake.InitializeInput{
		Email:  "momo@example.com",
		Amount: 5000,
	})

	momo1, err := client.Charge.MobileMoney(ctx, payfake.ChargeMomoInput{
		Email:     "momo@example.com",
		Reference: tx2.Reference,
		MobileMoney: &payfake.MomoDetails{
			Phone:    "+233241234567",
			Provider: "mtn",
		},
	})
	if err != nil {
		log.Fatalf("momo charge failed: %v", err)
	}
	fmt.Println("MoMo step 1:", momo1.Status) // "send_otp"

	// Get OTP
	momoOTPLogs, _ := client.Control.GetOTPLogs(ctx, token, tx2.Reference, payfake.ListOptions{})
	momo2, _ := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
		Reference: tx2.Reference,
		OTP:       momoOTPLogs[0].OTPCode,
	})
	fmt.Println("MoMo step 2:", momo2.Status) // "pay_offline"

	// Poll until resolved
	fmt.Println("Polling for MoMo resolution...")
	for i := 0; i < 10; i++ {
		result, err := client.Transaction.PublicVerify(ctx, tx2.Reference, tx2.AccessCode)
		if err != nil {
			break
		}
		fmt.Printf("  poll %d: status=%s flow=%s\n", i+1, result.Status,
			func() string {
				if result.Charge != nil {
					return result.Charge.FlowStatus
				}
				return "–"
			}())
		if result.Status == "success" || result.Status == "failed" {
			fmt.Println("MoMo resolved:", result.Status)
			break
		}
		time.Sleep(1 * time.Second)
	}

	//  Scenario testing
	fmt.Println("\n── Scenario testing ──")

	// Force all charges to fail with insufficient funds
	status := "failed"
	code := "CHARGE_INSUFFICIENT_FUNDS"
	scenario, err := client.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
		ForceStatus: &status,
		ErrorCode:   &code,
	})
	if err != nil {
		log.Fatalf("update scenario failed: %v", err)
	}
	fmt.Printf("Scenario set: force_status=%s error_code=%s\n",
		scenario.ForceStatus, scenario.ErrorCode)

	// This charge will now fail
	tx3, _ := client.Transaction.Initialize(ctx, payfake.InitializeInput{
		Email:  "fail@example.com",
		Amount: 10000,
	})

	_, err = client.Charge.Card(ctx, payfake.ChargeCardInput{
		Email:     "fail@example.com",
		Reference: tx3.Reference,
		Card: &payfake.CardDetails{
			Number:      "5061000000000000",
			CVV:         "123",
			ExpiryMonth: "12",
			ExpiryYear:  "2026",
		},
	})

	if err != nil {
		fmt.Println("Charge failed as expected:", err)
		if payfake.IsCode(err, payfake.CodeInsufficientFunds) {
			fmt.Println("Correctly identified as insufficient funds")
		}
	}

	// // Reset when done
	_, err = client.Control.ResetScenario(ctx, token)
	if err != nil {
		log.Fatalf("reset scenario failed: %v", err)
	}
	fmt.Println("Scenario reset — charges will succeed again")

	// //  Stats
	stats, err := client.Control.GetStats(ctx, token)
	if err != nil {
		log.Fatalf("get stats failed: %v", err)
	}
	fmt.Printf("\nStats:\n  total=%d successful=%d failed=%d success_rate=%.1f%%\n",
		stats.Transactions.Total,
		stats.Transactions.Successful,
		stats.Transactions.Failed,
		stats.Transactions.SuccessRate,
	)
}
