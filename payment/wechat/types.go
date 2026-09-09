package pay

import "time"

const CurrencyCNY = "CNY"

type PrepayRequest struct {
	Description string
	OutTradeNo  string
	NotifyURL   string
	Total       int64
	Currency    string
	Attach      string
	GoodsTag    string
	TimeExpire  *time.Time
}

type H5PrepayRequest struct {
	PrepayRequest
	PayerClientIP string
	SceneType     string
	AppName       string
	AppURL        string
	BundleID      string
	PackageName   string
}

type AppPrepayResponse struct {
	PrepayID  string
	PartnerID string
	TimeStamp string
	NonceStr  string
	Package   string
	Sign      string
}

type NativePrepayResponse struct {
	CodeURL string
}

type H5PrepayResponse struct {
	H5URL string
}

type Order struct {
	TransactionID  string
	OutTradeNo     string
	TradeState     string
	TradeStateDesc string
	TradeType      string
	SuccessTime    string
	Amount         OrderAmount
	Payer          Payer
}

type OrderAmount struct {
	Total         int64
	PayerTotal    int64
	Currency      string
	PayerCurrency string
}

type Payer struct {
	OpenID string
}

type PaymentNotify struct {
	ID           string
	EventType    string
	ResourceType string
	Summary      string
	Order        Order
}
