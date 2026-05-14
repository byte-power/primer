package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/byte-power/primer"
)

func main() {
	apiKey := os.Getenv("PRIMER_API_KEY")
	if apiKey == "" {
		log.Fatal("please set PRIMER_API_KEY environment variable")
	}

	customerID := os.Getenv("PRIMER_CUSTOMER_ID")
	if customerID == "" {
		log.Fatal("please set PRIMER_CUSTOMER_ID environment variable")
	}

	client := primer.NewClient(apiKey, nil)

	// ===== 1. 保存支付方式（Vault）=====
	paymentMethodToken := os.Getenv("PRIMER_PAYMENT_METHOD_TOKEN")
	if paymentMethodToken != "" {
		fmt.Println("=== Vault Payment Method ===")

		saved, err := client.VaultPaymentMethod(paymentMethodToken, &primer.VaultPaymentMethodRequest{
			CustomerID: customerID,
		})
		if err != nil {
			log.Fatalf("VaultPaymentMethod failed: %s (status=%d, errorId=%s)",
				err.Message, err.StatusCode, err.ErrorID)
		}

		fmt.Printf("Saved Token:  %s\n", saved.Token)
		fmt.Printf("Token Type:   %s\n", saved.TokenType)
		fmt.Printf("Method Type:  %s\n", saved.PaymentInstrumentType)
		fmt.Printf("Customer ID:  %s\n", saved.CustomerID)
		fmt.Printf("Is Default:   %v\n", saved.Default)
	}

	// ===== 2. 列表查询已保存的支付方式 =====
	fmt.Println("\n=== List Payment Methods ===")

	methods, err := client.ListPaymentMethods(customerID)
	if err != nil {
		log.Fatalf("ListPaymentMethods failed: %s (status=%d, errorId=%s)",
			err.Message, err.StatusCode, err.ErrorID)
	}

	fmt.Printf("Found %d saved payment methods\n", len(methods.Data))
	for i, m := range methods.Data {
		fmt.Printf("\n  [%d] Token: %s\n", i+1, m.Token)
		fmt.Printf("      Type:      %s\n", m.PaymentMethodType)
		fmt.Printf("      Default:   %v\n", m.Default)
		fmt.Printf("      Created:   %s\n", m.CreatedAt)

		if m.Description != "" {
			fmt.Printf("      Desc:      %s\n", m.Description)
		}

		if len(m.PaymentMethodData) > 0 && m.PaymentMethodType == "PAYMENT_CARD" {
			var card primer.PaymentCardData
			if err := json.Unmarshal(m.PaymentMethodData, &card); err == nil {
				fmt.Printf("      Card:      %s ****%s (%s/%s) network=%s\n",
					card.CardholderName, card.Last4Digits,
					card.ExpirationMonth, card.ExpirationYear,
					card.Network)
			}
		}
	}

	// ===== 3. 设置默认支付方式 =====
	if len(methods.Data) > 0 {
		token := methods.Data[0].Token

		fmt.Println("\n=== Set Default Payment Method ===")

		updated, err := client.SetDefaultPaymentMethod(token)
		if err != nil {
			log.Fatalf("SetDefaultPaymentMethod failed: %s", err.Message)
		}

		fmt.Printf("Token: %s → default: %v\n", updated.Token, updated.Default)
	}

	// ===== 4. 删除支付方式 =====
	// 取消注释以下代码可实际执行删除
	/*
		if len(methods.Data) > 0 {
			token := methods.Data[len(methods.Data)-1].Token

			fmt.Println("\n=== Delete Payment Method ===")

			deleted, err := client.DeletePaymentMethod(token)
			if err != nil {
				log.Fatalf("DeletePaymentMethod failed: %s", err.Message)
			}

			fmt.Printf("Deleted token: %s (deletedAt: %v)\n", deleted.Token, deleted.DeletedAt)
		}
	*/
}
