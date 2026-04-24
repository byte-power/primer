package primer

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ===== Payments API =====

// CreatePayment 创建支付
// POST /payments
func (c *Client) CreatePayment(req *CreatePaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost("/payments", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Deprecated: Use CreatePayment with WithIdempotencyKey option instead.
//
//	resp, err := client.CreatePayment(req, primer.WithIdempotencyKey("key"))
func (c *Client) CreatePaymentWithIdempotencyKey(req *CreatePaymentRequest, idempotencyKey string) (*PaymentResponse, *Error) {
	return c.CreatePayment(req, WithIdempotencyKey(idempotencyKey))
}

// GetPayment 获取支付详情
// GET /payments/{id}
func (c *Client) GetPayment(paymentID string) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doGet(fmt.Sprintf("/payments/%s", paymentID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CapturePayment 捕获支付
// POST /payments/{id}/capture
func (c *Client) CapturePayment(paymentID string, req *CapturePaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/capture", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelPayment 取消支付
// POST /payments/{id}/cancel
func (c *Client) CancelPayment(paymentID string, req *CancelPaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/cancel", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RefundPayment 退款
// POST /payments/{id}/refund
func (c *Client) RefundPayment(paymentID string, req *RefundPaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/refund", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResumePayment 恢复支付
// POST /payments/{id}/resume
func (c *Client) ResumePayment(paymentID string, req *ResumePaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/resume", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AuthorizePayment 授权支付
// POST /payments/{id}/authorize
func (c *Client) AuthorizePayment(paymentID string, req *AuthorizePaymentRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/authorize", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AdjustAuthorization 调整授权金额
// POST /payments/{id}/adjust-authorization
func (c *Client) AdjustAuthorization(paymentID string, req *AdjustAuthorizationRequest, opts ...RequestOption) (*PaymentResponse, *Error) {
	var resp PaymentResponse
	if err := c.doPost(fmt.Sprintf("/payments/%s/adjust-authorization", paymentID), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPayments 列表查询支付
// GET /payments
func (c *Client) ListPayments(params *PaymentListParams) (*PaymentListResponse, *Error) {
	var queryParams url.Values
	if params != nil {
		queryParams = params.toValues()
	}

	var resp PaymentListResponse
	if err := c.doGet("/payments", queryParams, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *PaymentListParams) toValues() url.Values {
	v := url.Values{}
	addSlice := func(key string, vals []string) {
		if len(vals) > 0 {
			v.Set(key, strings.Join(vals, ","))
		}
	}

	addSlice("status", p.Status)
	addSlice("payment_method_type", p.PaymentMethodType)
	addSlice("processor", p.Processor)
	addSlice("currency_code", p.CurrencyCode)
	addSlice("customer_id", p.CustomerID)
	addSlice("merchant_id", p.MerchantID)
	addSlice("customer_email_address", p.CustomerEmailAddress)
	addSlice("last_4_digits", p.Last4Digits)
	addSlice("paypal_email", p.PaypalEmail)
	addSlice("klarna_email", p.KlarnaEmail)

	if p.FromDate != "" {
		v.Set("from_date", p.FromDate)
	}
	if p.ToDate != "" {
		v.Set("to_date", p.ToDate)
	}
	if p.OrderID != "" {
		v.Set("order_id", p.OrderID)
	}
	if p.MinAmount != nil {
		v.Set("min_amount", strconv.FormatInt(*p.MinAmount, 10))
	}
	if p.MaxAmount != nil {
		v.Set("max_amount", strconv.FormatInt(*p.MaxAmount, 10))
	}
	if p.Limit != nil {
		v.Set("limit", strconv.FormatInt(*p.Limit, 10))
	}
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	}
	return v
}
