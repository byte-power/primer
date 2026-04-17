package main

import (
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

	paymentMethodToken := os.Getenv("PRIMER_PAYMENT_METHOD_TOKEN")
	if paymentMethodToken == "" {
		log.Fatal("please set PRIMER_PAYMENT_METHOD_TOKEN environment variable")
	}

	client := primer.NewClient(apiKey, nil)

	// ===== 1. 创建支付 =====
	fmt.Println("=== Create Payment ===")

	amount := int64(5000)
	payment, err := client.CreatePayment(&primer.CreatePaymentRequest{
		PaymentMethodToken: paymentMethodToken,
		OrderID:            "order-demo-001",
		CurrencyCode:       "EUR",
		Amount:             &amount,
	})
	if err != nil {
		log.Fatalf("CreatePayment failed: %s (status=%d, errorId=%s)",
			err.Message, err.StatusCode, err.ErrorID)
	}

	fmt.Printf("Payment ID:  %s\n", payment.ID)
	fmt.Printf("Status:      %s\n", payment.Status)
	fmt.Printf("Amount:      %d %s\n", payment.Amount, payment.CurrencyCode)
	fmt.Printf("Order ID:    %s\n", payment.OrderID)

	if payment.RequiredAction != nil {
		fmt.Printf("Required Action: %s - %s\n", payment.RequiredAction.Name, payment.RequiredAction.Description)
		fmt.Println("Payment requires further action (e.g. 3DS). Use ResumePayment after completing it.")
		return
	}

	// ===== 2. 获取支付详情 =====
	fmt.Println("\n=== Get Payment ===")

	fetched, err := client.GetPayment(payment.ID)
	if err != nil {
		log.Fatalf("GetPayment failed: %s", err.Message)
	}

	fmt.Printf("Payment ID:  %s\n", fetched.ID)
	fmt.Printf("Status:      %s\n", fetched.Status)
	fmt.Printf("Transactions: %d\n", len(fetched.Transactions))
	for i, tx := range fetched.Transactions {
		fmt.Printf("  [%d] %s via %s → %s (%d %s)\n",
			i+1, tx.TransactionType, tx.ProcessorName, tx.ProcessorStatus,
			tx.Amount, tx.CurrencyCode)
	}

	// ===== 3. 捕获支付（仅在 AUTHORIZED 状态时）=====
	if fetched.Status == primer.PaymentStatusAuthorized {
		fmt.Println("\n=== Capture Payment ===")

		captured, err := client.CapturePayment(payment.ID, nil)
		if err != nil {
			log.Fatalf("CapturePayment failed: %s", err.Message)
		}

		fmt.Printf("Status after capture: %s\n", captured.Status)

		// ===== 4. 退款（仅在 SETTLED/SETTLING 状态时）=====
		if captured.Status == primer.PaymentStatusSettled || captured.Status == primer.PaymentStatusSettling {
			fmt.Println("\n=== Refund Payment ===")

			refundAmount := int64(2000)
			refunded, err := client.RefundPayment(payment.ID, &primer.RefundPaymentRequest{
				Amount: &refundAmount,
				Reason: "partial refund demo",
			})
			if err != nil {
				log.Fatalf("RefundPayment failed: %s", err.Message)
			}

			fmt.Printf("Status after refund: %s\n", refunded.Status)
			if refunded.Processor != nil {
				fmt.Printf("Amount refunded:     %d\n", refunded.Processor.AmountRefunded)
			}
		}
	}

	// ===== 5. 查询支付列表 =====
	fmt.Println("\n=== List Payments ===")

	limit := int64(5)
	result, err := client.ListPayments(&primer.PaymentListParams{
		Status: []string{"AUTHORIZED", "SETTLED", "SETTLING"},
		Limit:  &limit,
	})
	if err != nil {
		log.Fatalf("ListPayments failed: %s", err.Message)
	}

	fmt.Printf("Found %d payments\n", len(result.Data))
	for _, p := range result.Data {
		fmt.Printf("  [%s] %s — %d %s (order: %s)\n",
			p.Status, p.ID, p.Amount, p.CurrencyCode, p.OrderID)
	}

	if result.NextCursor != "" {
		fmt.Printf("Next cursor: %s\n", result.NextCursor)
	}
}
