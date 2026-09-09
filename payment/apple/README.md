# Apple 支付

`payment/apple` 用于 iOS、macOS 的 App Store 一次性内购收据校验，并返回最近一笔有效交易，供业务发放权益。

迁移：将 `github.com/sailxy/x/apple/pay` 改为 `github.com/sailxy/x/payment/apple`。本包不提供跨平台统一支付接口，也不包含订阅或 Apple Pay 能力。

```go
applePay := pay.New(pay.Config{
	Endpoint: "https://buy.itunes.apple.com/verifyReceipt",
	BundleID: "com.example.app",
})

order, err := applePay.GetRecentOrder(receipt)
```
