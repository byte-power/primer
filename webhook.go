package primer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ===== Webhook 签名验证 =====

// VerifyWebhookSignature 验证 webhook 请求的签名
// 依次检查 X-Signature-Primary 和 X-Signature-Secondary
func VerifyWebhookSignature(body []byte, r *http.Request, signingSecret string) error {
	if signingSecret == "" {
		return nil
	}
	primarySig := r.Header.Get("X-Signature-Primary")
	if primarySig == "" {
		return fmt.Errorf("missing X-Signature-Primary header")
	}
	if verifySignature(body, signingSecret, primarySig) {
		return nil
	}
	secondarySig := r.Header.Get("X-Signature-Secondary")
	if secondarySig != "" && verifySignature(body, signingSecret, secondarySig) {
		return nil
	}
	return fmt.Errorf("webhook signature verification failed")
}

// verifySignature 使用 HMAC-SHA256 验证 webhook 签名
func verifySignature(payload []byte, secret, expectedSig string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computed := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(expectedSig))
}

// readAndVerify 从 HTTP 请求中读取 body 并验证签名
func readAndVerify(r *http.Request, signingSecret string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	if err := VerifyWebhookSignature(body, r, signingSecret); err != nil {
		return nil, err
	}
	return body, nil
}

// ===== PAYMENT.STATUS Webhook =====

// ParsePaymentStatusWebhook 从原始 JSON 字节解析 PAYMENT.STATUS Webhook 载荷
func ParsePaymentStatusWebhook(data []byte) (*PaymentStatusWebhook, error) {
	var webhook PaymentStatusWebhook
	if err := json.Unmarshal(data, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &webhook, nil
}

// ParsePaymentStatusWebhookFromRequest 从 HTTP 请求中解析 PAYMENT.STATUS Webhook 载荷
// 如果提供了 signingSecret，将验证签名
func ParsePaymentStatusWebhookFromRequest(r *http.Request, signingSecret string) (*PaymentStatusWebhook, error) {
	body, err := readAndVerify(r, signingSecret)
	if err != nil {
		return nil, err
	}
	return ParsePaymentStatusWebhook(body)
}

// ===== PAYMENT.REFUND Webhook =====

// ParsePaymentRefundWebhook 从原始 JSON 字节解析 PAYMENT.REFUND Webhook 载荷
func ParsePaymentRefundWebhook(data []byte) (*PaymentRefundWebhook, error) {
	var webhook PaymentRefundWebhook
	if err := json.Unmarshal(data, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &webhook, nil
}

// ParsePaymentRefundWebhookFromRequest 从 HTTP 请求中解析 PAYMENT.REFUND Webhook 载荷
// 如果提供了 signingSecret，将验证签名
func ParsePaymentRefundWebhookFromRequest(r *http.Request, signingSecret string) (*PaymentRefundWebhook, error) {
	body, err := readAndVerify(r, signingSecret)
	if err != nil {
		return nil, err
	}
	return ParsePaymentRefundWebhook(body)
}

// ===== DISPUTE.OPENED Webhook =====

// ParseDisputeOpenWebhook 从原始 JSON 字节解析 DISPUTE.OPENED Webhook 载荷
func ParseDisputeOpenWebhook(data []byte) (*DisputeOpenWebhook, error) {
	var webhook DisputeOpenWebhook
	if err := json.Unmarshal(data, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &webhook, nil
}

// ParseDisputeOpenWebhookFromRequest 从 HTTP 请求中解析 DISPUTE.OPENED Webhook 载荷
// 如果提供了 signingSecret，将验证签名
func ParseDisputeOpenWebhookFromRequest(r *http.Request, signingSecret string) (*DisputeOpenWebhook, error) {
	body, err := readAndVerify(r, signingSecret)
	if err != nil {
		return nil, err
	}
	return ParseDisputeOpenWebhook(body)
}

// ===== DISPUTE.STATUS Webhook =====

// ParseDisputeStatusWebhook 从原始 JSON 字节解析 DISPUTE.STATUS Webhook 载荷
func ParseDisputeStatusWebhook(data []byte) (*DisputeStatusWebhook, error) {
	var webhook DisputeStatusWebhook
	if err := json.Unmarshal(data, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &webhook, nil
}

// ParseDisputeStatusWebhookFromRequest 从 HTTP 请求中解析 DISPUTE.STATUS Webhook 载荷
// 如果提供了 signingSecret，将验证签名
func ParseDisputeStatusWebhookFromRequest(r *http.Request, signingSecret string) (*DisputeStatusWebhook, error) {
	body, err := readAndVerify(r, signingSecret)
	if err != nil {
		return nil, err
	}
	return ParseDisputeStatusWebhook(body)
}

// ===== 通用 Webhook 解析 =====

// WebhookEvent 通用 webhook 事件，用于先提取 eventType 再按类型解析
type WebhookEvent struct {
	EventType string          `json:"eventType"`
	RawData   json.RawMessage `json:"-"`
}

// ParseWebhookEventType 从原始 JSON 字节中提取 eventType
func ParseWebhookEventType(data []byte) (string, error) {
	var event struct {
		EventType string `json:"eventType"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return "", fmt.Errorf("failed to parse webhook event type: %w", err)
	}
	return event.EventType, nil
}

// ParseWebhookFromRequest 从 HTTP 请求中读取并验证签名，返回原始 body 和 eventType
// 调用方可根据 eventType 选择对应的 Parse 函数进行解析
func ParseWebhookFromRequest(r *http.Request, signingSecret string) (eventType string, body []byte, err error) {
	body, err = readAndVerify(r, signingSecret)
	if err != nil {
		return "", nil, err
	}
	eventType, err = ParseWebhookEventType(body)
	if err != nil {
		return "", nil, err
	}
	return eventType, body, nil
}
