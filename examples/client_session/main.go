package main

import (
	"fmt"
	"log"
	"os"

	"github.com/byte-power/primer"
)

type StdLogger struct{}

func (l *StdLogger) Debug(msg string, fields ...primer.Field) {
	log.Printf("[DEBUG] %s %v", msg, formatFields(fields))
}

func (l *StdLogger) Info(msg string, fields ...primer.Field) {
	log.Printf("[INFO]  %s %v", msg, formatFields(fields))
}

func (l *StdLogger) Error(msg string, fields ...primer.Field) {
	log.Printf("[ERROR] %s %v", msg, formatFields(fields))
}

func formatFields(fields []primer.Field) map[string]interface{} {
	m := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}

func main() {
	apiKey := os.Getenv("PRIMER_API_KEY")
	if apiKey == "" {
		log.Fatal("please set PRIMER_API_KEY environment variable")
	}

	client := primer.NewClient(apiKey, &StdLogger{})

	// ===== 1. 创建客户端会话 =====
	fmt.Println("=== Create Client Session ===")

	amount := int64(5000)
	created, err := client.CreateClientSession(&primer.CreateClientSessionRequest{
		OrderID:      "order-demo-001",
		CurrencyCode: "EUR",
		Amount:       &amount,
		Customer: &primer.CustomerDetails{
			EmailAddress: "john@example.com",
			FirstName:    "John",
			LastName:     "Doe",
			BillingAddress: &primer.Address{
				AddressLine1: "123 Main St",
				City:         "Berlin",
				CountryCode:  "DE",
				PostalCode:   "10115",
			},
		},
		Order: &primer.OrderDetails{
			LineItems: []primer.OrderLineItem{
				{
					ItemID:      "item-1",
					Description: "T-Shirt",
					Amount:      3000,
					Quantity:    1,
				},
				{
					ItemID:      "item-2",
					Description: "Shipping",
					Amount:      2000,
					Quantity:    1,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateClientSession failed: %s (status=%d, errorId=%s)",
			err.Message, err.StatusCode, err.ErrorID)
	}

	fmt.Printf("Client Token:  %s\n", created.ClientToken)
	fmt.Printf("Expires:       %s\n", created.ClientTokenExpirationDate)
	fmt.Printf("Order ID:      %s\n", created.OrderID)
	fmt.Printf("Amount:        %d %s\n", created.Amount, created.CurrencyCode)

	if len(created.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range created.Warnings {
			fmt.Printf("  [%s] %s: %s\n", w.Type, w.Code, w.Message)
		}
	}

	// ===== 2. 更新客户端会话 =====
	fmt.Println("\n=== Update Client Session ===")

	newAmount := int64(6500)
	updated, err := client.UpdateClientSession(&primer.UpdateClientSessionRequest{
		ClientToken: created.ClientToken,
		Amount:      &newAmount,
		Metadata: map[string]interface{}{
			"source": "go-sdk-example",
		},
	})
	if err != nil {
		log.Fatalf("UpdateClientSession failed: %s (status=%d, errorId=%s)",
			err.Message, err.StatusCode, err.ErrorID)
	}

	fmt.Printf("Updated Amount: %d %s\n", updated.Amount, updated.CurrencyCode)

	// ===== 3. 检索客户端会话 =====
	fmt.Println("\n=== Retrieve Client Session ===")

	session, err := client.GetClientSession(created.ClientToken)
	if err != nil {
		log.Fatalf("GetClientSession failed: %s (status=%d, errorId=%s)",
			err.Message, err.StatusCode, err.ErrorID)
	}

	fmt.Printf("Customer ID:   %s\n", session.CustomerID)
	fmt.Printf("Order ID:      %s\n", session.OrderID)
	fmt.Printf("Currency:      %s\n", session.CurrencyCode)
	fmt.Printf("Amount:        %d\n", session.Amount)

	if session.Customer != nil {
		fmt.Printf("Email:         %s\n", session.Customer.EmailAddress)
		fmt.Printf("Name:          %s %s\n", session.Customer.FirstName, session.Customer.LastName)
	}

	if session.Order != nil {
		fmt.Printf("Line Items:    %d\n", len(session.Order.LineItems))
		for i, item := range session.Order.LineItems {
			fmt.Printf("  [%d] %s - %d (qty: %d)\n", i+1, item.Description, item.Amount, item.Quantity)
		}
	}

	if session.PaymentMethod != nil {
		fmt.Printf("Vault on Success: %v\n", session.PaymentMethod.VaultOnSuccess)
		fmt.Printf("Authorization:    %s\n", session.PaymentMethod.AuthorizationType)
	}
}
