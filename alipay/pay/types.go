package pay

import (
	"context"
	"net/url"

	sdk "github.com/smartwalle/alipay/v3"
)

type paymentService interface {
	AppPay(context.Context, sdkAppPayRequest) (string, error)
	PagePay(context.Context, sdkPagePayRequest) (string, error)
	WapPay(context.Context, sdkWapPayRequest) (string, error)
}

type orderService interface {
	QueryOrder(context.Context, string) (*sdk.TradeQueryRsp, error)
	CloseOrder(context.Context, string) error
}

type notifyParser interface {
	ParsePaymentNotify(context.Context, url.Values) (*sdk.Notification, error)
}

type PayRequest struct {
	Subject     string
	OutTradeNo  string
	TotalAmount string
	ProductCode string
	Body        string
	NotifyURL   string
}

type PagePayRequest struct {
	PayRequest
	ReturnURL string
}

type WapPayRequest struct {
	PayRequest
	ReturnURL string
	QuitURL   string
}

type AppPayResponse struct {
	OrderString string
}

type PagePayResponse struct {
	URL string
}

type WapPayResponse struct {
	URL string
}

type Order struct {
	TradeNo      string
	OutTradeNo   string
	TradeStatus  string
	TotalAmount  string
	BuyerLogonID string
	BuyerUserID  string
	BuyerOpenID  string
	SendPayDate  string
}

type PaymentNotify struct {
	AppID          string
	NotifyID       string
	NotifyType     string
	NotifyTime     string
	TradeNo        string
	OutTradeNo     string
	TradeStatus    string
	TotalAmount    string
	ReceiptAmount  string
	BuyerID        string
	BuyerLogonID   string
	BuyerOpenID    string
	SellerID       string
	Subject        string
	Body           string
	GmtCreate      string
	GmtPayment     string
	PassbackParams string
}

type sdkAppPayRequest struct {
	Subject     string
	OutTradeNo  string
	TotalAmount string
	ProductCode string
	Body        string
	NotifyURL   string
}

type sdkPagePayRequest struct {
	sdkAppPayRequest
	ReturnURL string
}

type sdkWapPayRequest struct {
	sdkAppPayRequest
	ReturnURL string
	QuitURL   string
}
