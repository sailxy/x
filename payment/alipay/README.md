# 支付宝支付

`payment/alipay` 封装支付宝普通商户的一次性收款 API，并提供业务请求和响应类型。

迁移：将 `github.com/sailxy/x/alipay/pay` 改为 `github.com/sailxy/x/payment/alipay`。本包不提供跨平台统一支付接口。

```go
alipayPay, err := alipay.New(alipay.Config{
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
resp, err := alipayPay.AppPay(ctx, alipay.PayRequest{
	Subject:     "membership",
	OutTradeNo:  "order-123",
	TotalAmount: "9.90",
})
```

## Web Payment

PC Web payment:

```go
resp, err := alipayPay.PagePay(ctx, alipay.PagePayRequest{
	PayRequest: alipay.PayRequest{
		Subject:     "membership",
		OutTradeNo:  "order-123",
		TotalAmount: "9.90",
	},
})
```

Mobile Web WAP payment:

```go
resp, err := alipayPay.WapPay(ctx, alipay.WapPayRequest{
	PayRequest: alipay.PayRequest{
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
