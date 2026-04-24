package primer

import (
	"fmt"
	"net/url"
)

// ===== Payment Methods API =====

// VaultPaymentMethod 保存支付方式
// POST /payment-instruments/{paymentMethodToken}/vault
func (c *Client) VaultPaymentMethod(paymentMethodToken string, req *VaultPaymentMethodRequest, opts ...RequestOption) (*PaymentMethodTokenResponse, *Error) {
	var resp PaymentMethodTokenResponse
	if err := c.doPost(fmt.Sprintf("/payment-instruments/%s/vault", paymentMethodToken), req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPaymentMethods 列表查询已保存的支付方式
// GET /payment-instruments?customer_id=xxx
func (c *Client) ListPaymentMethods(customerID string) (*PaymentMethodTokenListResponse, *Error) {
	params := url.Values{}
	params.Set("customer_id", customerID)

	var resp PaymentMethodTokenListResponse
	if err := c.doGet("/payment-instruments", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeletePaymentMethod 删除已保存的支付方式
// DELETE /payment-instruments/{paymentMethodToken}
func (c *Client) DeletePaymentMethod(paymentMethodToken string, opts ...RequestOption) (*PaymentMethodTokenResponse, *Error) {
	var resp PaymentMethodTokenResponse
	if err := c.doDelete(fmt.Sprintf("/payment-instruments/%s", paymentMethodToken), &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetDefaultPaymentMethod 设置默认支付方式
// POST /payment-instruments/{paymentMethodToken}/default
func (c *Client) SetDefaultPaymentMethod(paymentMethodToken string, opts ...RequestOption) (*PaymentMethodTokenResponse, *Error) {
	var resp PaymentMethodTokenResponse
	if err := c.doPost(fmt.Sprintf("/payment-instruments/%s/default", paymentMethodToken), nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
