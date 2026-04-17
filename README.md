# Primer Go SDK

[Primer](https://primer.io) API v2.4 的 Go SDK，覆盖 Client Session、Payments、Payment Methods 全部 API 及 Webhook 解析。

## 安装

```bash
go get github.com/byte-power/primer
```

## 快速开始

### 创建客户端

```go
import "github.com/byte-power/primer"

// Sandbox 环境（默认）
client := primer.NewClient("your-api-key", logger)

// Production 环境
client := primer.NewProductionClient("your-api-key", logger)

// 自定义 Base URL
client := primer.NewClientWithBaseURL("your-api-key", "https://custom.url", logger)

// 自定义 HTTP Client
client := primer.NewClientWithHTTPClient("your-api-key", primer.SandboxBaseURL, httpClient, logger)
```

## Client Session API

管理客户端会话，生成用于初始化 Universal Checkout 的 client token。

### 创建会话

```go
amount := int64(5000)
session, err := client.CreateClientSession(&primer.CreateClientSessionRequest{
    OrderID:      "order-123",
    CurrencyCode: "EUR",
    Amount:       &amount,
    Customer: &primer.CustomerDetails{
        EmailAddress: "john@example.com",
        FirstName:    "John",
        LastName:     "Doe",
    },
})
if err != nil {
    log.Fatalf("create session failed: %s", err.Message)
}
fmt.Printf("Client Token: %s (expires: %s)\n", session.ClientToken, session.ClientTokenExpirationDate)
```

### 更新会话

```go
newAmount := int64(6000)
updated, err := client.UpdateClientSession(&primer.UpdateClientSessionRequest{
    ClientToken: session.ClientToken,
    Amount:      &newAmount,
})
```

### 检索会话

```go
session, err := client.GetClientSession("client-token-xxx")
```

## Payments API

完整的支付生命周期管理：创建、授权、捕获、退款、取消。

### 创建支付

```go
payment, err := client.CreatePayment(&primer.CreatePaymentRequest{
    PaymentMethodToken: "pm-token-xxx",
    OrderID:            "order-123",
    CurrencyCode:       "EUR",
    Amount:             &amount,
})
```

使用幂等键防止重复支付：

```go
payment, err := client.CreatePaymentWithIdempotencyKey(req, "unique-idempotency-key")
```

### 获取支付详情

```go
payment, err := client.GetPayment("payment-id")
```

### 捕获支付

全额捕获（不传 body）：

```go
payment, err := client.CapturePayment("payment-id", nil)
```

部分捕获：

```go
captureAmount := int64(3000)
payment, err := client.CapturePayment("payment-id", &primer.CapturePaymentRequest{
    Amount: &captureAmount,
})
```

### 取消支付

```go
payment, err := client.CancelPayment("payment-id", &primer.CancelPaymentRequest{
    Reason: "customer requested cancellation",
})
```

### 退款

全额退款：

```go
payment, err := client.RefundPayment("payment-id", nil)
```

部分退款：

```go
refundAmount := int64(1000)
payment, err := client.RefundPayment("payment-id", &primer.RefundPaymentRequest{
    Amount:  &refundAmount,
    OrderID: "order-123",
    Reason:  "item returned",
})
```

### 恢复支付

当支付被 Workflow 暂停时（如等待 3DS 验证），恢复支付流程：

```go
payment, err := client.ResumePayment("payment-id", &primer.ResumePaymentRequest{
    ResumeToken: "resume-token-xxx",
})
```

### 手动授权支付

```go
payment, err := client.AuthorizePayment("payment-id", &primer.AuthorizePaymentRequest{
    Processor: primer.AuthorizeProcessor{
        ProcessorMerchantID: "merchant-id-123",
        Name:                "STRIPE",
    },
})
```

### 调整授权金额

仅在 `authorizationType` 为 `ESTIMATED` 时可用：

```go
payment, err := client.AdjustAuthorization("payment-id", &primer.AdjustAuthorizationRequest{
    Amount: 7500,
})
```

### 查询支付列表

```go
limit := int64(20)
result, err := client.ListPayments(&primer.PaymentListParams{
    Status:       []string{"AUTHORIZED", "SETTLED"},
    CurrencyCode: []string{"EUR"},
    FromDate:     "2025-01-01T00:00:00Z",
    Limit:        &limit,
})

for _, p := range result.Data {
    fmt.Printf("[%s] %s %d %s\n", p.Status, p.ID, p.Amount, p.CurrencyCode)
}

// 翻页
if result.NextCursor != "" {
    nextPage, _ := client.ListPayments(&primer.PaymentListParams{
        Cursor: result.NextCursor,
    })
    _ = nextPage
}
```

## Payment Methods API

管理客户已保存的支付方式。

### 保存支付方式

将一次性 token 保存为可重复使用的支付方式：

```go
saved, err := client.VaultPaymentMethod("single-use-token", &primer.VaultPaymentMethodRequest{
    CustomerID: "customer-123",
})
```

### 列表查询已保存的支付方式

```go
methods, err := client.ListPaymentMethods("customer-123")

for _, m := range methods.Data {
    fmt.Printf("[%s] %s (default: %v)\n", m.PaymentMethodType, m.Token, m.Default)
}
```

### 设置默认支付方式

```go
updated, err := client.SetDefaultPaymentMethod("payment-method-token")
```

### 删除已保存的支付方式

```go
deleted, err := client.DeletePaymentMethod("payment-method-token")
```

## Webhook 解析

SDK 支持解析所有 Primer Webhook 事件，并内置 HMAC-SHA256 签名验证。

### 按事件类型解析

```go
// PAYMENT.STATUS
webhook, err := primer.ParsePaymentStatusWebhookFromRequest(r, signingSecret)

// PAYMENT.REFUND
webhook, err := primer.ParsePaymentRefundWebhookFromRequest(r, signingSecret)

// DISPUTE.OPENED
webhook, err := primer.ParseDisputeOpenWebhookFromRequest(r, signingSecret)

// DISPUTE.STATUS
webhook, err := primer.ParseDisputeStatusWebhookFromRequest(r, signingSecret)
```

### 通用路由模式

当一个端点需要处理多种 webhook 事件时，使用通用解析器先提取 `eventType`，再按类型路由：

```go
eventType, body, err := primer.ParseWebhookFromRequest(r, signingSecret)
if err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}

switch eventType {
case "PAYMENT.STATUS":
    webhook, _ := primer.ParsePaymentStatusWebhook(body)
    handlePaymentStatus(webhook)
case "PAYMENT.REFUND":
    webhook, _ := primer.ParsePaymentRefundWebhook(body)
    handlePaymentRefund(webhook)
case "DISPUTE.OPENED":
    webhook, _ := primer.ParseDisputeOpenWebhook(body)
    handleDisputeOpen(webhook)
case "DISPUTE.STATUS":
    webhook, _ := primer.ParseDisputeStatusWebhook(body)
    handleDisputeStatus(webhook)
}
```

### 从 JSON 字节解析（不验证签名）

```go
webhook, err := primer.ParsePaymentStatusWebhook(jsonBytes)
```

## Debug 日志

SDK 在每次 API 请求和响应时都会输出 debug 日志。只需传入一个实现 `primer.Logger` 接口的实例即可：

```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```

示例：基于标准库 `log` 的实现：

```go
type StdLogger struct{}

func (l *StdLogger) Debug(msg string, fields ...primer.Field) {
    log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *StdLogger) Info(msg string, fields ...primer.Field) {
    log.Printf("[INFO]  %s %v", msg, fields)
}

func (l *StdLogger) Error(msg string, fields ...primer.Field) {
    log.Printf("[ERROR] %s %v", msg, fields)
}
```

传入 `nil` 时使用内置的 `NopLogger`（不输出任何日志）。

日志事件说明：

| 事件名 | 级别 | 说明 |
|--------|------|------|
| `primer_request` | Debug | API 请求发起，包含 method、url、body |
| `primer_response` | Debug | API 响应返回，包含 status_code 和 body |
| `primer_request_error` | Error | HTTP 请求执行失败 |
| `primer_read_response_error` | Error | 读取响应体失败 |
| `primer_api_error` | Error | API 返回业务错误，包含 error_id 和 description |

## 错误处理

SDK 返回 `*primer.Error`，包含以下字段：

```go
type Error struct {
    Message    string // 错误描述
    StatusCode int    // HTTP 状态码（API 错误时）
    ErrorID    string // Primer API 错误 ID（API 错误时）
    Err        error  // 底层错误（可用 errors.Unwrap 获取）
}
```

示例：

```go
payment, err := client.GetPayment("invalid-id")
if err != nil {
    fmt.Printf("Error: %s\n", err.Message)
    fmt.Printf("HTTP Status: %d\n", err.StatusCode)
    fmt.Printf("Error ID: %s\n", err.ErrorID)

    // 根据 HTTP 状态码处理
    switch err.StatusCode {
    case 400:
        // 请求参数错误
    case 401:
        // API Key 无效
    case 404:
        // 资源不存在
    case 409:
        // 状态冲突（如重复捕获）
    case 422:
        // 验证错误
    }
}
```

## 示例

完整的可运行示例位于 [`examples/`](./examples/) 目录：

| 示例 | 说明 |
|------|------|
| [`examples/client_session/`](./examples/client_session/) | 创建、更新、检索 Client Session |
| [`examples/payment/`](./examples/payment/) | 完整支付生命周期：创建 → 捕获 → 退款 → 查询 |
| [`examples/payment_method/`](./examples/payment_method/) | 保存、查询、设置默认、删除支付方式 |
| [`examples/webhook/`](./examples/webhook/) | 启动 HTTP 服务器，接收并路由所有类型的 Webhook 事件 |

### 运行示例

```bash
cd examples

# Client Session：创建 + 更新 + 检索
PRIMER_API_KEY=your-key go run ./client_session/

# 支付生命周期（需要有效的 payment method token）
PRIMER_API_KEY=your-key PRIMER_PAYMENT_METHOD_TOKEN=pm-xxx go run ./payment/

# 支付方式管理
PRIMER_API_KEY=your-key PRIMER_CUSTOMER_ID=cust-123 PRIMER_PAYMENT_METHOD_TOKEN=pm-xxx go run ./payment_method/

# Webhook 服务器（监听 :8080）
PRIMER_WEBHOOK_SECRET=your-secret go run ./webhook/
```

## API 参考

本 SDK 基于 [Primer API v2.4](https://primer.io/docs/api-reference/get-started/overview) 构建。

### Client Session API

| 方法 | HTTP | 端点 | 说明 |
|------|------|------|------|
| `CreateClientSession` | POST | `/client-session` | [创建客户端会话](https://primer.io/docs/api-reference/v2.4/api-reference/client-session-api/create-a-client-session) |
| `UpdateClientSession` | PATCH | `/client-session` | [更新客户端会话](https://primer.io/docs/api-reference/v2.4/api-reference/client-session-api/update-client-session) |
| `GetClientSession` | GET | `/client-session` | [检索客户端会话](https://primer.io/docs/api-reference/v2.4/api-reference/client-session-api/retrieve-a-client-session) |

### Payments API

| 方法 | HTTP | 端点 | 说明 |
|------|------|------|------|
| `CreatePayment` | POST | `/payments` | [创建支付](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/create-a-payment) |
| `GetPayment` | GET | `/payments/{id}` | [获取支付详情](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/get-a-payment) |
| `CapturePayment` | POST | `/payments/{id}/capture` | [捕获支付](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/capture-a-payment) |
| `CancelPayment` | POST | `/payments/{id}/cancel` | [取消支付](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/cancel-a-payment) |
| `RefundPayment` | POST | `/payments/{id}/refund` | [退款](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/refund-a-payment) |
| `ResumePayment` | POST | `/payments/{id}/resume` | [恢复支付](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/resume-a-payment) |
| `AuthorizePayment` | POST | `/payments/{id}/authorize` | [授权支付](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/authorize-a-payment) |
| `AdjustAuthorization` | POST | `/payments/{id}/adjust-authorization` | [调整授权金额](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/adjust-authorized-amount) |
| `ListPayments` | GET | `/payments` | [查询支付列表](https://primer.io/docs/api-reference/v2.4/api-reference/payments-api/search-&-list-payments) |

### Payment Methods API

| 方法 | HTTP | 端点 | 说明 |
|------|------|------|------|
| `VaultPaymentMethod` | POST | `/payment-instruments/{token}/vault` | [保存支付方式](https://primer.io/docs/api-reference/v2.4/api-reference/payment-methods-api/save-a-payment-method-token) |
| `ListPaymentMethods` | GET | `/payment-instruments` | [列表查询已保存的支付方式](https://primer.io/docs/api-reference/v2.4/api-reference/payment-methods-api/list-saved-payment-methods) |
| `DeletePaymentMethod` | DELETE | `/payment-instruments/{token}` | [删除已保存的支付方式](https://primer.io/docs/api-reference/v2.4/api-reference/payment-methods-api/delete-a-saved-payment-method) |
| `SetDefaultPaymentMethod` | POST | `/payment-instruments/{token}/default` | [设置默认支付方式](https://primer.io/docs/api-reference/v2.4/api-reference/payment-methods-api/update-the-default-saved-payment-method) |

### Webhook 解析

| 函数 | 事件类型 | 说明 |
|------|----------|------|
| `ParsePaymentStatusWebhook(FromRequest)` | `PAYMENT.STATUS` | [支付状态更新](https://primer.io/docs/api-reference/endpoints/v2.4/payment-webhooks/payment-status-update) |
| `ParsePaymentRefundWebhook(FromRequest)` | `PAYMENT.REFUND` | [退款完成](https://primer.io/docs/api-reference/endpoints/v2.4/payment-webhooks/payment-refund) |
| `ParseDisputeOpenWebhook(FromRequest)` | `DISPUTE.OPENED` | [争议开启](https://primer.io/docs/api-reference/endpoints/v2.4/dispute-&-chargebacks-webhooks/dispute-open) |
| `ParseDisputeStatusWebhook(FromRequest)` | `DISPUTE.STATUS` | [争议状态更新](https://primer.io/docs/api-reference/endpoints/v2.4/dispute-&-chargebacks-webhooks/dispute-status) |
| `ParseWebhookFromRequest` | 通用 | 验签 + 提取 eventType，用于路由分发 |

## License

MIT
