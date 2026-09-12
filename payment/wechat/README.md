# 微信支付

`payment/wechat` 封装微信支付 API v3 普通商户的一次性收款 API，并提供业务请求和响应类型。

迁移：将 `github.com/sailxy/x/wechat/pay` 改为 `github.com/sailxy/x/payment/wechat`。本包不提供跨平台统一支付接口。

```go
wxpay, err := wechat.New(wechat.Config{
	AppID:                      "wx...",
	MchID:                      "1900000001",
	MchCertificateSerialNumber: "cert-serial",
	MchPrivateKeyPath:          "/path/to/apiclient_key.pem",
	MchAPIv3Key:                "32-byte-api-v3-key",
})
```

## App Payment

```go
resp, err := wxpay.AppPrepay(ctx, wechat.PrepayRequest{
	Description: "membership",
	OutTradeNo: "order-123",
	NotifyURL:  "https://example.com/wxpay/notify",
	Total:      100,
})
```

## Web Payment

PC Web QR-code payment:

```go
resp, err := wxpay.NativePrepay(ctx, wechat.PrepayRequest{
	Description: "membership",
	OutTradeNo: "order-123",
	NotifyURL:  "https://example.com/wxpay/notify",
	Total:      100,
})
```

Mobile Web H5 payment:

```go
resp, err := wxpay.H5Prepay(ctx, wechat.H5PrepayRequest{
	PrepayRequest: wechat.PrepayRequest{
		Description: "membership",
		OutTradeNo: "order-123",
		NotifyURL:  "https://example.com/wxpay/notify",
		Total:      100,
	},
	PayerClientIP: "203.0.113.1",
	SceneType:     "Wap",
	AppName:       "Example",
	AppURL:        "https://example.com",
})
```

## Orders and Notifications

```go
order, err := wxpay.QueryOrderByOutTradeNo(ctx, "order-123")
err = wxpay.CloseOrderByOutTradeNo(ctx, "order-123")

notify, err := wxpay.ParsePaymentNotify(ctx, request)
```
