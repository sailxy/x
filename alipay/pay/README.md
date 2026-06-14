# Alipay Pay

`alipay/pay` wraps Alipay ordinary merchant payment APIs with business-oriented request and response types.

```go
alipayPay, err := pay.New(pay.Config{
	AppID:              "2021000000000000",
	PrivateKeyPath:     "/path/to/app_private_key.pem",
	AlipayPublicKeyPEM: "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----",
	NotifyURL:          "https://example.com/alipay/notify",
	ReturnURL:          "https://example.com/pay/return",
	Sandbox:            true,
})
```

## App Payment

```go
resp, err := alipayPay.AppPay(ctx, pay.PayRequest{
	Subject:     "membership",
	OutTradeNo:  "order-123",
	TotalAmount: "9.90",
})
```

## Web Payment

PC Web payment:

```go
resp, err := alipayPay.PagePay(ctx, pay.PagePayRequest{
	PayRequest: pay.PayRequest{
		Subject:     "membership",
		OutTradeNo:  "order-123",
		TotalAmount: "9.90",
	},
})
```

Mobile Web WAP payment:

```go
resp, err := alipayPay.WapPay(ctx, pay.WapPayRequest{
	PayRequest: pay.PayRequest{
		Subject:     "membership",
		OutTradeNo:  "order-123",
		TotalAmount: "9.90",
	},
	QuitURL: "https://example.com/pay/cancel",
})
```

## Orders and Notifications

```go
order, err := alipayPay.QueryOrder(ctx, "order-123")
err = alipayPay.CloseOrder(ctx, "order-123")

notify, err := alipayPay.ParsePaymentNotify(request)
```
