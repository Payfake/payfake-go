package main

import (
	"context"
	"fmt"
	"log"

	payfake "github.com/payfake/payfake-go"
)

func main() {
	ctx := context.Background()

	// STEP 1: Create auth client (no secret key needed for auth endpoints)

	authClient := payfake.New(payfake.Config{
		SecretKey: "", // Not required for auth endpoints
		BaseURL:   "http://localhost:8080",
	})

	// STEP 2: Register or Login to get auth token

	var token string

	// Try to register first
	regResp, err := authClient.Auth.Register(ctx, payfake.RegisterInput{
		BusinessName: "G_KANAD",
		Email:        "gkanad@acme.com",
		Password:     "secret123",
	})
	if err != nil {
		// If email already taken, log in instead
		if payfake.IsCode(err, payfake.CodeEmailTaken) {
			loginResp, err := authClient.Auth.Login(ctx, payfake.LoginInput{
				Email:    "gkanad@acme.com",
				Password: "secret123",
			})
			if err != nil {
				log.Fatalf("login failed: %v", err)
			}
			token = loginResp.Token
			fmt.Printf("Logged in as: %s\n", loginResp.Merchant.Email)
		} else {
			log.Fatalf("registration failed: %v", err)
		}
	} else {
		token = regResp.Token
		fmt.Printf("Registered: %s\n", regResp.Merchant.ID)
	}

	// STEP 3: Get actual API keys using auth token

	keys, err := authClient.Auth.GetKeys(ctx, token)
	if err != nil {
		log.Fatalf("get keys failed: %v", err)
	}

	fmt.Println("\nAPI Keys retrieved:")
	fmt.Printf("Public Key: %s\n", keys.PublicKey)
	fmt.Printf("Secret Key: %s...\n", keys.SecretKey[:15])

	// STEP 4: Create authenticated client with real secret key

	client := payfake.New(payfake.Config{
		SecretKey: keys.SecretKey,
		BaseURL:   "http://localhost:8080",
	})

	// STEP 5: Initialize a transaction

	tx, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
		Email:       "archergorden@gmail.com",
		Amount:      10000, // GHS 100.00 — amounts are in the smallest unit (pesewas)
		Currency:    "GHS",
		CallbackURL: "http://localhost:5173",
	})
	if err != nil {
		log.Fatalf("initialize failed: %v", err)
	}

	fmt.Println("\nTransaction initialized")
	fmt.Println("Reference:         ", tx.Reference)
	fmt.Println("Access code:       ", tx.AccessCode)
	fmt.Println("Authorization URL: ", tx.AuthorizationURL)

	if err != nil {
		log.Fatalf("initialize failed: %v", err)
	}
	fmt.Println("Reference:", tx.Reference)
	fmt.Println("Access code:", tx.AccessCode)

	// Step 1 — initiate card charge
	step1, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
		AccessCode: tx.AccessCode,
		CardNumber: "5061000000000000", // Verve — local card → send_pin
		CardExpiry: "12/26",
		CVV:        "123",
		Email:      "customer@example.com",
	})
	if err != nil {
		log.Fatalf("charge failed: %v", err)
	}
	fmt.Println("Step 1 status:", step1.Status) // "send_pin"

	// Step 2 — submit PIN
	step2, err := client.Charge.SubmitPIN(ctx, payfake.SubmitPINInput{
		Reference: tx.Reference,
		PIN:       "1234",
	})
	if err != nil {
		log.Fatalf("submit PIN failed: %v", err)
	}
	fmt.Println("Step 2 status:", step2.Status) // "send_otp"

	// Step 3 — get OTP from logs
	otpLogs, err := client.Control.GetOTPLogs(ctx, token, tx.Reference)
	if err != nil {
		log.Fatalf("get OTP logs failed: %v", err)
	}
	if len(otpLogs) == 0 {
		log.Fatal("no OTP logs found")
	}
	otp := otpLogs[0].OTPCode
	fmt.Println("OTP from logs:", otp)

	// Step 4 — submit OTP
	step3, err := client.Charge.SubmitOTP(ctx, payfake.SubmitOTPInput{
		Reference: tx.Reference,
		OTP:       otp,
	})
	if err != nil {
		log.Fatalf("submit OTP failed: %v", err)
	}
	fmt.Println("Step 3 status:", step3.Status) // "success"

	// Verify
	verified, err := client.Transaction.Verify(ctx, tx.Reference)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}
	fmt.Println("Verified:", verified.Status)

	// STEP 6: Charge a card

	// charge, err := client.Charge.Card(ctx, payfake.ChargeCardInput{
	// 	AccessCode: tx.AccessCode,
	// 	CardNumber: "4111111111111111",
	// 	CardExpiry: "12/26",
	// 	CVV:        "123",
	// 	Email:      "customer@example.com",
	// })
	// if err != nil {
	// 	if payfake.IsCode(err, payfake.CodeChargeFailed) {
	// 		fmt.Printf("\nCharge failed: %v\n", err)
	// 	} else {
	// 		log.Fatalf("charge failed: %v", err)
	// 	}
	// } else {
	// 	fmt.Println("\nCharge status:", charge.Transaction.Status)
	// }

	// // STEP 7: Verify transaction

	// verified, err := client.Transaction.Verify(ctx, tx.Reference)
	// if err != nil {
	// 	log.Fatalf("verify failed: %v", err)
	// }
	// fmt.Println("Verified status:", verified.Status)

	// STEP 8: Mobile Money flow

	// tx2, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
	// 	Email:  "momo@example.com",
	// 	Amount: 5000,
	// })
	// if err != nil {
	// 	log.Fatalf("momo initialize failed: %v", err)
	// }

	fmt.Println("Transaction:\n", *tx)

	// 	// Check the correct type name for mobile money
	// 	momo, err := client.Charge.MobileMoney(ctx, payfake.ChargeMomoInput{
	// 		AccessCode: tx2.AccessCode,
	// 		Phone:      "+233241234567",
	// 		Provider:   "mtn",
	// 		Email:      "momo@example.com",
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("momo charge failed: %v", err)
	// 	}
	// 	fmt.Println("\nMoMo status:", momo.Transaction.Status)

	// 	// STEP 9: Control panel operations (using auth token, not secret key)

	// 	failureRate := 0.5
	// 	delayMs := 2000
	// 	scenario, err := authClient.Control.UpdateScenario(ctx, token, payfake.UpdateScenarioInput{
	// 		FailureRate: &failureRate,
	// 		DelayMS:     &delayMs, // Note: DelayMS not DelayMs
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("scenario update failed: %v", err)
	// 	}
	// 	fmt.Printf("\nScenario updated, failure rate: %.2f\n", scenario.FailureRate)

	// 	// Force a specific transaction to fail
	// 	tx3, err := client.Transaction.Initialize(ctx, payfake.InitializeInput{
	// 		Email:  "force@example.com",
	// 		Amount: 2000,
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("force transaction initialize failed: %v", err)
	// 	}

	// 	status := "failed"
	// 	errorCode := "CHARGE_INSUFFICIENT_FUNDS"
	// 	forced, err := authClient.Control.ForceTransaction(ctx, token, tx3.Reference, payfake.ForceTransactionInput{
	// 		Status:    status,    // Not a pointer
	// 		ErrorCode: errorCode, // Not a pointer
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("force transaction failed: %v", err)
	// 	}
	// 	fmt.Println("Forced status:", forced.Status)

	// 	// Reset scenario
	// 	_, err = authClient.Control.ResetScenario(ctx, token)
	// 	if err != nil {
	// 		log.Fatalf("scenario reset failed: %v", err)
	// 	}
	// 	fmt.Println("Scenario reset")

	// 	// STEP 10: Get recent logs

	//	logs, err := authClient.Control.GetLogs(ctx, token, payfake.ListOptions{
	//		Page:    1,
	//		PerPage: 5,
	//	})
	//
	//	if err != nil {
	//		if payfake.IsCode(err, "LOGS_EMPTY") {
	//			fmt.Println("\nNo logs found yet (expected for new merchant)")
	//		} else {
	//			log.Fatalf("get logs failed: %v", err)
	//		}
	//	} else {
	//
	//		fmt.Printf("\nRecent requests: %d\n", len(logs))
	//		for _, log := range logs {
	//			fmt.Printf("  %s %s -> %d\n", log.Method, log.Path, log.StatusCode)
	//		}
	//	}
}
