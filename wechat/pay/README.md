# WeChat Pay

`wechat/pay` wraps WeChat Pay API v3 ordinary merchant mode with business-oriented request and response types.

```go
wxpay, err := pay.New(pay.Config{
	AppID:                      "wx...",
	MchID:                      "1900000001",
	MchCertificateSerialNumber: "cert-serial",
	MchPrivateKeyPath:          "/path/to/apiclient_key.pem",
	MchAPIv3Key:                "32-byte-api-v3-key",
})
```

## App Payment

```go
resp, err := wxpay.AppPrepay(ctx, pay.PrepayRequest{
	Description: "membership",
	OutTradeNo: "order-123",
	NotifyURL:  "https://example.com/wxpay/notify",
	Total:      100,
})
```

## Web Payment

PC Web QR-code payment:

```go
resp, err := wxpay.NativePrepay(ctx, pay.PrepayRequest{
	Description: "membership",
	OutTradeNo: "order-123",
	NotifyURL:  "https://example.com/wxpay/notify",
	Total:      100,
})
```

Mobile Web H5 payment:

```go
resp, err := wxpay.H5Prepay(ctx, pay.H5PrepayRequest{
	PrepayRequest: pay.PrepayRequest{
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
