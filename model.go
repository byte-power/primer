package primer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Time 兼容 Primer API 日期格式的时间类型
// Primer API 返回的日期可能不含时区后缀（如 "2021-03-24T14:56:56.869248"），
// 而 Go 标准库 time.Time 的 JSON 反序列化要求 RFC3339 格式。
// 此类型自动处理多种格式。
type Time struct {
	time.Time
}

var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05.999999",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func (t *Time) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		return nil
	}
	for _, layout := range timeFormats {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("unable to parse time %q", s)
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", t.Time.Format(time.RFC3339Nano))), nil
}

// ===== Client Session API =====

// CreateClientSessionRequest 创建客户端会话请求
type CreateClientSessionRequest struct {
	OrderID       string                 `json:"orderId,omitempty"`
	CurrencyCode  string                 `json:"currencyCode,omitempty"`
	Amount        *int64                 `json:"amount,omitempty"`
	Order         *OrderDetails          `json:"order,omitempty"`
	CustomerID    string                 `json:"customerId,omitempty"`
	Customer      *CustomerDetails       `json:"customer,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	PaymentMethod *PaymentMethodRequest  `json:"paymentMethod,omitempty"`
}

// PaymentMethodRequest 支付方式请求选项
type PaymentMethodRequest struct {
	VaultOnSuccess             *bool                  `json:"vaultOnSuccess,omitempty"`
	VaultOn3DS                 *bool                  `json:"vaultOn3DS,omitempty"`
	VaultOnAgreement           *bool                  `json:"vaultOnAgreement,omitempty"`
	Descriptor                 string                 `json:"descriptor,omitempty"`
	PaymentType                string                 `json:"paymentType,omitempty"`
	OrderedAllowedCardNetworks []string               `json:"orderedAllowedCardNetworks,omitempty"`
	Options                    map[string]interface{} `json:"options,omitempty"`
	AuthorizationType          string                 `json:"authorizationType,omitempty"`
	FirstPaymentReason         string                 `json:"firstPaymentReason,omitempty"`
}

// ClientSessionWithTokenResponse 创建客户端会话响应（含 clientToken）
type ClientSessionWithTokenResponse struct {
	ClientToken               string                 `json:"clientToken"`
	ClientTokenExpirationDate Time                   `json:"clientTokenExpirationDate"`
	CustomerID                string                 `json:"customerId,omitempty"`
	OrderID                   string                 `json:"orderId,omitempty"`
	CurrencyCode              string                 `json:"currencyCode,omitempty"`
	Amount                    int64                  `json:"amount,omitempty"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty"`
	Customer                  *CustomerDetails       `json:"customer,omitempty"`
	Order                     *OrderDetails          `json:"order,omitempty"`
	PaymentMethod             *PaymentMethodOptions  `json:"paymentMethod"`
	Warnings                  []ClientSessionWarning `json:"warnings,omitempty"`
}

// ClientSessionWarning 客户端会话警告
type ClientSessionWarning struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UpdateClientSessionRequest 更新客户端会话请求
type UpdateClientSessionRequest struct {
	ClientToken   string                 `json:"clientToken"`
	CustomerID    string                 `json:"customerId,omitempty"`
	OrderID       string                 `json:"orderId,omitempty"`
	CurrencyCode  string                 `json:"currencyCode,omitempty"`
	Amount        *int64                 `json:"amount,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Customer      *CustomerDetails       `json:"customer,omitempty"`
	Order         *OrderDetails          `json:"order,omitempty"`
	PaymentMethod *PaymentMethodRequest  `json:"paymentMethod,omitempty"`
}

// ClientSessionResponse 客户端会话响应（retrieve / update）
type ClientSessionResponse struct {
	CustomerID    string                 `json:"customerId,omitempty"`
	OrderID       string                 `json:"orderId,omitempty"`
	CurrencyCode  string                 `json:"currencyCode,omitempty"`
	Amount        int64                  `json:"amount,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Customer      *CustomerDetails       `json:"customer,omitempty"`
	Order         *OrderDetails          `json:"order,omitempty"`
	PaymentMethod *PaymentMethodOptions  `json:"paymentMethod"`
}

// CustomerDetails 客户详情
type CustomerDetails struct {
	EmailAddress       string   `json:"emailAddress,omitempty"`
	MobileNumber       string   `json:"mobileNumber,omitempty"`
	FirstName          string   `json:"firstName,omitempty"`
	LastName           string   `json:"lastName,omitempty"`
	BillingAddress     *Address `json:"billingAddress,omitempty"`
	ShippingAddress    *Address `json:"shippingAddress,omitempty"`
	TaxID              string   `json:"taxId,omitempty"`
	NationalDocumentID string   `json:"nationalDocumentId,omitempty"`
}

// Address 地址信息
type Address struct {
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	AddressLine1 string `json:"addressLine1,omitempty"`
	AddressLine2 string `json:"addressLine2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
}

// OrderDetails 订单详情
type OrderDetails struct {
	LineItems           []OrderLineItem `json:"lineItems,omitempty"`
	CountryCode         string          `json:"countryCode,omitempty"`
	RetailerCountryCode string          `json:"retailerCountryCode,omitempty"`
	Fees                []OrderFee      `json:"fees,omitempty"`
	Shipping            *OrderShipping  `json:"shipping,omitempty"`
}

// OrderLineItem 订单行项目
type OrderLineItem struct {
	ItemID         string       `json:"itemId,omitempty"`
	Name           string       `json:"name,omitempty"`
	Description    string       `json:"description,omitempty"`
	Amount         int64        `json:"amount"`
	Quantity       int64        `json:"quantity,omitempty"`
	DiscountAmount int64        `json:"discountAmount,omitempty"`
	TaxAmount      int64        `json:"taxAmount,omitempty"`
	TaxCode        string       `json:"taxCode,omitempty"`
	ProductType    string       `json:"productType,omitempty"`
	ProductData    *ProductData `json:"productData,omitempty"`
}

// ProductData 产品数据
type ProductData struct {
	SKU                    string  `json:"sku,omitempty"`
	Brand                  string  `json:"brand,omitempty"`
	Color                  string  `json:"color,omitempty"`
	GlobalTradeItemNumber  string  `json:"globalTradeItemNumber,omitempty"`
	ManufacturerPartNumber string  `json:"manufacturerPartNumber,omitempty"`
	Weight                 float64 `json:"weight,omitempty"`
	WeightUnit             string  `json:"weightUnit,omitempty"`
	PageURL                string  `json:"pageUrl,omitempty"`
}

// OrderFee 订单费用
type OrderFee struct {
	Amount      int64  `json:"amount"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// OrderShipping 订单运费
type OrderShipping struct {
	Amount            int64  `json:"amount,omitempty"`
	MethodID          string `json:"methodId,omitempty"`
	MethodName        string `json:"methodName,omitempty"`
	MethodDescription string `json:"methodDescription,omitempty"`
}

// PaymentMethodOptions 支付方式选项（Client Session 响应）
type PaymentMethodOptions struct {
	VaultOnSuccess             bool                   `json:"vaultOnSuccess"`
	VaultOn3DS                 bool                   `json:"vaultOn3DS"`
	VaultOnAgreement           bool                   `json:"vaultOnAgreement"`
	Descriptor                 string                 `json:"descriptor,omitempty"`
	PaymentType                string                 `json:"paymentType,omitempty"`
	OrderedAllowedCardNetworks []string               `json:"orderedAllowedCardNetworks"`
	Options                    map[string]interface{} `json:"options,omitempty"`
	AuthorizationType          string                 `json:"authorizationType,omitempty"`
	FirstPaymentReason         string                 `json:"firstPaymentReason,omitempty"`
}

// ===== Payments API =====

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	PaymentMethodToken string                 `json:"paymentMethodToken"`
	OrderID            string                 `json:"orderId,omitempty"`
	CurrencyCode       string                 `json:"currencyCode,omitempty"`
	Amount             *int64                 `json:"amount,omitempty"`
	Order              *OrderDetails          `json:"order,omitempty"`
	CustomerID         string                 `json:"customerId,omitempty"`
	Customer           *CustomerDetails       `json:"customer,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	PaymentMethod      *PaymentMethodRequest  `json:"paymentMethod,omitempty"`
}

// PaymentResponse 支付 API 通用响应
type PaymentResponse struct {
	ID             string                 `json:"id"`
	Date           Time                   `json:"date"`
	DateUpdated    Time                   `json:"dateUpdated"`
	Status         PaymentStatus          `json:"status"`
	OrderID        string                 `json:"orderId"`
	CurrencyCode   string                 `json:"currencyCode"`
	Amount         int64                  `json:"amount"`
	CustomerID     string                 `json:"customerId,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Customer       *CustomerDetails       `json:"customer,omitempty"`
	Order          *OrderDetails          `json:"order,omitempty"`
	PaymentMethod  *WebhookPaymentMethod  `json:"paymentMethod"`
	Processor      *Processor             `json:"processor,omitempty"`
	RequiredAction *RequiredAction        `json:"requiredAction,omitempty"`
	StatusReason   *StatusReason          `json:"statusReason,omitempty"`
	Transactions   []Transaction          `json:"transactions"`
	RiskData       *RiskData              `json:"riskData,omitempty"`
}

// RequiredAction 需要执行的操作（如 3DS）
type RequiredAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ClientToken string `json:"clientToken,omitempty"`
}

// CapturePaymentRequest 捕获支付请求
type CapturePaymentRequest struct {
	Amount   *int64                 `json:"amount,omitempty"`
	Final    *bool                  `json:"final,omitempty"`
	Order    *CaptureOrderDetails   `json:"order,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CaptureOrderDetails capture 时可更新的订单信息
type CaptureOrderDetails struct {
	RetailerCountryCode string `json:"retailerCountryCode,omitempty"`
}

// CancelPaymentRequest 取消支付请求
type CancelPaymentRequest struct {
	Reason string `json:"reason,omitempty"`
}

// RefundPaymentRequest 退款请求
type RefundPaymentRequest struct {
	Amount             *int64 `json:"amount,omitempty"`
	OrderID            string `json:"orderId,omitempty"`
	Reason             string `json:"reason,omitempty"`
	TransactionEventID string `json:"transactionEventId,omitempty"`
}

// ResumePaymentRequest 恢复支付请求
type ResumePaymentRequest struct {
	ResumeToken string `json:"resumeToken"`
}

// AuthorizePaymentRequest 授权支付请求
type AuthorizePaymentRequest struct {
	Processor AuthorizeProcessor `json:"processor"`
}

// AuthorizeProcessor 授权处理器信息
type AuthorizeProcessor struct {
	Name                string `json:"name,omitempty"`
	ProcessorMerchantID string `json:"processorMerchantId"`
}

// AdjustAuthorizationRequest 调整授权金额请求
type AdjustAuthorizationRequest struct {
	Amount int64 `json:"amount"`
}

// PaymentListParams 支付列表查询参数
type PaymentListParams struct {
	Status               []string
	PaymentMethodType    []string
	Processor            []string
	CurrencyCode         []string
	FromDate             string
	ToDate               string
	OrderID              string
	MinAmount            *int64
	MaxAmount            *int64
	CustomerID           []string
	MerchantID           []string
	CustomerEmailAddress []string
	Last4Digits          []string
	PaypalEmail          []string
	KlarnaEmail          []string
	Limit                *int64
	Cursor               string
}

// PaymentListResponse 支付列表响应
type PaymentListResponse struct {
	Data       []PaymentSummary `json:"data"`
	NextCursor string           `json:"nextCursor,omitempty"`
	PrevCursor string           `json:"prevCursor,omitempty"`
}

// PaymentSummary 支付摘要
type PaymentSummary struct {
	ID           string                 `json:"id"`
	Date         Time                   `json:"date"`
	DateUpdated  Time                   `json:"dateUpdated"`
	Status       PaymentStatus          `json:"status"`
	OrderID      string                 `json:"orderId"`
	CurrencyCode string                 `json:"currencyCode"`
	Amount       int64                  `json:"amount"`
	Processor    *PaymentSummaryProc    `json:"processor,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentSummaryProc 支付摘要中的处理器信息
type PaymentSummaryProc struct {
	Name                string `json:"name"`
	ProcessorMerchantID string `json:"processorMerchantId,omitempty"`
}

// ===== Payment Methods API =====

// VaultPaymentMethodRequest 保存支付方式请求
type VaultPaymentMethodRequest struct {
	CustomerID string `json:"customerId"`
}

// PaymentMethodTokenResponse 支付方式令牌响应
type PaymentMethodTokenResponse struct {
	CreatedAt         Time            `json:"createdAt"`
	Token             string          `json:"token"`
	TokenType         string          `json:"tokenType"`
	AnalyticsID       string          `json:"analyticsId"`
	PaymentMethodType string          `json:"paymentMethodType"`
	PaymentMethodData json.RawMessage `json:"paymentMethodData"`
	CustomerID        string          `json:"customerId"`
	Default           bool            `json:"default"`
	DeletedAt         *Time           `json:"deletedAt,omitempty"`
	Deleted           bool            `json:"deleted,omitempty"`
	Description       string          `json:"description,omitempty"`
}

// PaymentMethodTokenListResponse 支付方式列表响应
type PaymentMethodTokenListResponse struct {
	Data []PaymentMethodTokenResponse `json:"data"`
}

// ===== Error Responses =====

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}

// ErrorObject API 错误对象
type ErrorObject struct {
	ErrorID            string        `json:"errorId"`
	Description        string        `json:"description"`
	RecoverySuggestion string        `json:"recoverySuggestion,omitempty"`
	DiagnosticsID      string        `json:"diagnosticsId,omitempty"`
	ValidationErrors   []interface{} `json:"validationErrors,omitempty"`
	PaymentID          string        `json:"paymentId,omitempty"`
	PaymentStatus      string        `json:"paymentStatus,omitempty"`
}

// ===== Payment Status Enums =====

// PaymentStatus 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending          PaymentStatus = "PENDING"
	PaymentStatusFailed           PaymentStatus = "FAILED"
	PaymentStatusAuthorized       PaymentStatus = "AUTHORIZED"
	PaymentStatusSettling         PaymentStatus = "SETTLING"
	PaymentStatusPartiallySettled PaymentStatus = "PARTIALLY_SETTLED"
	PaymentStatusSettled          PaymentStatus = "SETTLED"
	PaymentStatusDeclined         PaymentStatus = "DECLINED"
	PaymentStatusCancelled        PaymentStatus = "CANCELLED"
)

// ===== Webhook: Payment Status Update =====

// PaymentStatusWebhook PAYMENT.STATUS 事件载荷
type PaymentStatusWebhook struct {
	EventType          string              `json:"eventType"`
	Date               Time                `json:"date"`
	SignedAt           string              `json:"signedAt"`
	NotificationConfig *NotificationConfig `json:"notificationConfig"`
	Version            string              `json:"version,omitempty"`
	Payment            *PaymentResponse    `json:"payment"`
}

// PaymentRefundWebhook PAYMENT.REFUND 事件载荷
type PaymentRefundWebhook struct {
	EventType          string              `json:"eventType"`
	Date               Time                `json:"date"`
	SignedAt           string              `json:"signedAt"`
	NotificationConfig *NotificationConfig `json:"notificationConfig"`
	Version            string              `json:"version,omitempty"`
	Payment            *PaymentResponse    `json:"payment"`
}

// DisputeOpenWebhook DISPUTE.OPENED 事件载荷
type DisputeOpenWebhook struct {
	EventType          string `json:"eventType"`
	ProcessorID        string `json:"processorId,omitempty"`
	ProcessorDisputeID string `json:"processorDisputeId,omitempty"`
	PaymentID          string `json:"paymentId,omitempty"`
	TransactionID      string `json:"transactionId,omitempty"`
	OrderID            string `json:"orderId,omitempty"`
	PrimerAccountID    string `json:"primerAccountId,omitempty"`
}

// DisputeType 争议类型
type DisputeType string

const (
	DisputeTypeRetrieval      DisputeType = "RETRIEVAL"
	DisputeTypeDispute        DisputeType = "DISPUTE"
	DisputeTypePrearbitration DisputeType = "PREARBITRATION"
)

// DisputeStatus 争议状态
type DisputeStatus string

const (
	DisputeStatusOpen       DisputeStatus = "OPEN"
	DisputeStatusAccepted   DisputeStatus = "ACCEPTED"
	DisputeStatusChallenged DisputeStatus = "CHALLENGED"
	DisputeStatusExpired    DisputeStatus = "EXPIRED"
	DisputeStatusCancelled  DisputeStatus = "CANCELLED"
	DisputeStatusWon        DisputeStatus = "WON"
	DisputeStatusLost       DisputeStatus = "LOST"
)

// DisputeStatusWebhook DISPUTE.STATUS 事件载荷
type DisputeStatusWebhook struct {
	EventType           string                `json:"eventType"`
	Version             string                `json:"version,omitempty"`
	Type                DisputeType           `json:"type"`
	Status              DisputeStatus         `json:"status"`
	PrimerAccountID     string                `json:"primerAccountId,omitempty"`
	TransactionID       string                `json:"transactionId,omitempty"`
	OrderID             string                `json:"orderId,omitempty"`
	PaymentID           string                `json:"paymentId,omitempty"`
	PaymentMethod       *DisputePaymentMethod `json:"paymentMethod,omitempty"`
	Processor           string                `json:"processor,omitempty"`
	ProcessorDisputeID  string                `json:"processorDisputeId,omitempty"`
	ReceivedAt          *Time                 `json:"receivedAt,omitempty"`
	ChallengeRequiredBy *Time                 `json:"challengeRequiredBy,omitempty"`
	Reason              string                `json:"reason,omitempty"`
	ReasonCode          string                `json:"reasonCode,omitempty"`
	ProcessorReason     string                `json:"processorReason,omitempty"`
	Amount              int64                 `json:"amount,omitempty"`
	Currency            string                `json:"currency,omitempty"`
	MerchantID          string                `json:"merchantId,omitempty"`
}

// DisputePaymentMethod 争议中的支付方式
type DisputePaymentMethod struct {
	PaymentMethodType string                    `json:"paymentMethodType,omitempty"`
	PaymentMethodData *DisputePaymentMethodData `json:"paymentMethodData,omitempty"`
}

// DisputePaymentMethodData 争议中的支付方式数据
type DisputePaymentMethodData struct {
	Network string `json:"network,omitempty"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// WebhookPaymentMethod Webhook 中的支付方式信息
type WebhookPaymentMethod struct {
	Descriptor                 string                      `json:"descriptor,omitempty"`
	PaymentType                string                      `json:"paymentType,omitempty"`
	PaymentMethodToken         string                      `json:"paymentMethodToken,omitempty"`
	IsVaulted                  bool                        `json:"isVaulted,omitempty"`
	AnalyticsID                string                      `json:"analyticsId,omitempty"`
	PaymentMethodType          string                      `json:"paymentMethodType,omitempty"`
	PaymentMethodData          json.RawMessage             `json:"paymentMethodData,omitempty"`
	AuthorizationType          string                      `json:"authorizationType,omitempty"`
	ThreeDSecureAuthentication *ThreeDSecureAuthentication `json:"threeDSecureAuthentication,omitempty"`
}

// PaymentCardData 支付卡数据
type PaymentCardData struct {
	First6Digits       string   `json:"first6Digits,omitempty"`
	Last4Digits        string   `json:"last4Digits"`
	ExpirationMonth    string   `json:"expirationMonth"`
	ExpirationYear     string   `json:"expirationYear"`
	CardholderName     string   `json:"cardholderName,omitempty"`
	Network            string   `json:"network,omitempty"`
	IsNetworkTokenized bool     `json:"isNetworkTokenized"`
	BinData            *BinData `json:"binData,omitempty"`
}

// BinData 卡 BIN 数据
type BinData struct {
	Network                    string `json:"network"`
	IssuerCountryCode          string `json:"issuerCountryCode,omitempty"`
	IssuerName                 string `json:"issuerName,omitempty"`
	IssuerCurrencyCode         string `json:"issuerCurrencyCode,omitempty"`
	RegionalRestriction        string `json:"regionalRestriction"`
	AccountNumberType          string `json:"accountNumberType"`
	AccountFundingType         string `json:"accountFundingType"`
	PrepaidReloadableIndicator string `json:"prepaidReloadableIndicator"`
	ProductUsageType           string `json:"productUsageType"`
	ProductCode                string `json:"productCode"`
	ProductName                string `json:"productName"`
}

// ThreeDSecureAuthentication 3DS 认证信息
type ThreeDSecureAuthentication struct {
	ResponseCode    string `json:"responseCode"`
	ReasonCode      string `json:"reasonCode,omitempty"`
	ReasonText      string `json:"reasonText,omitempty"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	ChallengeIssued *bool  `json:"challengeIssued,omitempty"`
}

// Processor 处理器信息
type Processor struct {
	Name                string `json:"name,omitempty"`
	ProcessorMerchantID string `json:"processorMerchantId,omitempty"`
	AmountCaptured      int64  `json:"amountCaptured,omitempty"`
	AmountRefunded      int64  `json:"amountRefunded,omitempty"`
}

// StatusReason 状态原因
type StatusReason struct {
	Type                       string `json:"type"`
	DeclineType                string `json:"declineType,omitempty"`
	Code                       string `json:"code,omitempty"`
	Message                    string `json:"message,omitempty"`
	PaymentMethodResultCode    string `json:"paymentMethodResultCode,omitempty"`
	PaymentMethodResultMessage string `json:"paymentMethodResultMessage,omitempty"`
	PaymentMethodAdviceCode    string `json:"paymentMethodAdviceCode,omitempty"`
	PaymentMethodAdviceMessage string `json:"paymentMethodAdviceMessage,omitempty"`
	AdvisedAction              string `json:"advisedAction,omitempty"`
}

// Transaction 交易概况
type Transaction struct {
	Date                   Time          `json:"date"`
	Amount                 int64         `json:"amount"`
	CurrencyCode           string        `json:"currencyCode"`
	OrderID                string        `json:"orderId,omitempty"`
	TransactionType        string        `json:"transactionType"`
	ProcessorTransactionID string        `json:"processorTransactionId,omitempty"`
	ProcessorName          string        `json:"processorName"`
	ProcessorMerchantID    string        `json:"processorMerchantId"`
	ProcessorStatus        PaymentStatus `json:"processorStatus"`
	ProcessorStatusReason  *StatusReason `json:"processorStatusReason,omitempty"`
	Reason                 string        `json:"reason,omitempty"`
}

// ===== Risk Data =====

// RiskData 风险数据
type RiskData struct {
	FraudChecks *FraudChecks `json:"fraudChecks,omitempty"`
	CVVCheck    *CVVCheck    `json:"cvvCheck,omitempty"`
	AVSCheck    *AVSCheck    `json:"avsCheck,omitempty"`
}

// FraudChecks 欺诈检查
type FraudChecks struct {
	Source                         string `json:"source,omitempty"`
	PreAuthorizationResult         string `json:"preAuthorizationResult,omitempty"`
	PreAuthorizationRecommendation string `json:"preAuthorizationRecommendation,omitempty"`
	PostAuthorizationResult        string `json:"postAuthorizationResult,omitempty"`
}

// CVVCheck CVV 检查
type CVVCheck struct {
	Source string `json:"source,omitempty"`
	Result string `json:"result,omitempty"`
}

// AVSCheck AVS 检查
type AVSCheck struct {
	Source string          `json:"source,omitempty"`
	Result *AVSCheckResult `json:"result,omitempty"`
}

// AVSCheckResult AVS 检查结果
type AVSCheckResult struct {
	StreetAddress string `json:"streetAddress,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
}
