package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/byte-power/primer"
)

func main() {
	signingSecret := os.Getenv("PRIMER_WEBHOOK_SECRET")

	// 单一端点处理所有类型的 webhook，使用通用路由模式
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		eventType, body, err := primer.ParseWebhookFromRequest(r, signingSecret)
		if err != nil {
			log.Printf("webhook parse error: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		log.Printf("received webhook: eventType=%s", eventType)

		switch eventType {
		case "PAYMENT.STATUS":
			handlePaymentStatus(body)
		case "PAYMENT.REFUND":
			handlePaymentRefund(body)
		case "DISPUTE.OPENED":
			handleDisputeOpen(body)
		case "DISPUTE.STATUS":
			handleDisputeStatus(body)
		default:
			log.Printf("unhandled event type: %s", eventType)
		}

		w.WriteHeader(http.StatusOK)
	})

	// 也可以按事件类型分配到不同端点
	http.HandleFunc("/webhook/payment-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		webhook, err := primer.ParsePaymentStatusWebhookFromRequest(r, signingSecret)
		if err != nil {
			log.Printf("failed to parse webhook: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		handlePaymentStatusWebhook(webhook)
		w.WriteHeader(http.StatusOK)
	})

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	fmt.Printf("webhook server listening on %s\n", addr)
	fmt.Println("  POST /webhook               — unified endpoint (all event types)")
	fmt.Println("  POST /webhook/payment-status — dedicated payment status endpoint")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handlePaymentStatus(body []byte) {
	webhook, err := primer.ParsePaymentStatusWebhook(body)
	if err != nil {
		log.Printf("failed to parse PAYMENT.STATUS: %v", err)
		return
	}
	handlePaymentStatusWebhook(webhook)
}

func handlePaymentStatusWebhook(webhook *primer.PaymentStatusWebhook) {
	payment := webhook.Payment
	if payment == nil {
		log.Printf("PAYMENT.STATUS: no payment data")
		return
	}

	log.Printf("PAYMENT.STATUS: id=%s status=%s amount=%d %s order=%s",
		payment.ID, payment.Status, payment.Amount, payment.CurrencyCode, payment.OrderID)

	switch payment.Status {
	case primer.PaymentStatusPending:
		log.Printf("  → payment %s is pending", payment.ID)
	case primer.PaymentStatusAuthorized:
		log.Printf("  → payment %s authorized, ready to capture", payment.ID)
	case primer.PaymentStatusSettling:
		log.Printf("  → payment %s is settling", payment.ID)
	case primer.PaymentStatusSettled:
		log.Printf("  → payment %s settled, order %s fulfilled", payment.ID, payment.OrderID)
	case primer.PaymentStatusDeclined:
		if payment.StatusReason != nil {
			log.Printf("  → payment %s declined: type=%s code=%s message=%s",
				payment.ID, payment.StatusReason.Type, payment.StatusReason.Code, payment.StatusReason.Message)
		}
	case primer.PaymentStatusFailed:
		log.Printf("  → payment %s failed", payment.ID)
	case primer.PaymentStatusCancelled:
		log.Printf("  → payment %s cancelled", payment.ID)
	}

	printPaymentMethodInfo(payment)
	printTransactions(payment)
}

func handlePaymentRefund(body []byte) {
	webhook, err := primer.ParsePaymentRefundWebhook(body)
	if err != nil {
		log.Printf("failed to parse PAYMENT.REFUND: %v", err)
		return
	}

	payment := webhook.Payment
	if payment == nil {
		log.Printf("PAYMENT.REFUND: no payment data")
		return
	}

	log.Printf("PAYMENT.REFUND: id=%s status=%s amount=%d %s",
		payment.ID, payment.Status, payment.Amount, payment.CurrencyCode)

	for _, tx := range payment.Transactions {
		if tx.TransactionType == "REFUND" {
			log.Printf("  refund transaction: processor=%s status=%s amount=%d %s",
				tx.ProcessorName, tx.ProcessorStatus, tx.Amount, tx.CurrencyCode)
		}
	}
}

func handleDisputeOpen(body []byte) {
	webhook, err := primer.ParseDisputeOpenWebhook(body)
	if err != nil {
		log.Printf("failed to parse DISPUTE.OPENED: %v", err)
		return
	}

	log.Printf("DISPUTE.OPENED: paymentId=%s orderId=%s processor=%s disputeId=%s",
		webhook.PaymentID, webhook.OrderID, webhook.ProcessorID, webhook.ProcessorDisputeID)
}

func handleDisputeStatus(body []byte) {
	webhook, err := primer.ParseDisputeStatusWebhook(body)
	if err != nil {
		log.Printf("failed to parse DISPUTE.STATUS: %v", err)
		return
	}

	log.Printf("DISPUTE.STATUS: type=%s status=%s paymentId=%s amount=%d %s",
		webhook.Type, webhook.Status, webhook.PaymentID, webhook.Amount, webhook.Currency)

	if webhook.Reason != "" {
		log.Printf("  reason: %s (code: %s)", webhook.Reason, webhook.ReasonCode)
	}
	if webhook.ChallengeRequiredBy != nil {
		log.Printf("  challenge required by: %s", webhook.ChallengeRequiredBy)
	}
}

func printPaymentMethodInfo(payment *primer.PaymentResponse) {
	pm := payment.PaymentMethod
	if pm == nil {
		return
	}

	log.Printf("  payment method: type=%s vaulted=%v", pm.PaymentMethodType, pm.IsVaulted)

	if pm.PaymentMethodType == "PAYMENT_CARD" && len(pm.PaymentMethodData) > 0 {
		var card primer.PaymentCardData
		if err := json.Unmarshal(pm.PaymentMethodData, &card); err == nil {
			log.Printf("  card: %s ****%s (%s/%s) network=%s",
				card.CardholderName, card.Last4Digits,
				card.ExpirationMonth, card.ExpirationYear, card.Network)
		}
	}

	if pm.ThreeDSecureAuthentication != nil {
		tds := pm.ThreeDSecureAuthentication
		log.Printf("  3DS: response=%s protocol=%s", tds.ResponseCode, tds.ProtocolVersion)
	}
}

func printTransactions(payment *primer.PaymentResponse) {
	for i, tx := range payment.Transactions {
		log.Printf("  tx[%d]: type=%s processor=%s status=%s amount=%d %s",
			i, tx.TransactionType, tx.ProcessorName, tx.ProcessorStatus,
			tx.Amount, tx.CurrencyCode)
	}
}
