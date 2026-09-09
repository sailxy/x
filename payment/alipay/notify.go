package pay

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	sdk "github.com/smartwalle/alipay/v3"
)

func (p *Pay) ParsePaymentNotify(r *http.Request) (*PaymentNotify, error) {
	if r == nil {
		return nil, errors.New("alipay pay: notify request is required")
	}
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	notification, err := p.notify.ParsePaymentNotify(r.Context(), r.Form)
	if err != nil {
		return nil, err
	}
	return mapPaymentNotify(notification), nil
}

type sdkNotifyParser struct {
	client *sdk.Client
}

func newSDKNotifyParser(client *sdk.Client) notifyParser {
	return &sdkNotifyParser{client: client}
}

func (p *sdkNotifyParser) ParsePaymentNotify(ctx context.Context, values url.Values) (*sdk.Notification, error) {
	return p.client.DecodeNotification(ctx, values)
}

func mapPaymentNotify(notification *sdk.Notification) *PaymentNotify {
	if notification == nil {
		return nil
	}
	return &PaymentNotify{
		AppID:          notification.AppId,
		NotifyID:       notification.NotifyId,
		NotifyType:     notification.NotifyType,
		NotifyTime:     notification.NotifyTime,
		TradeNo:        notification.TradeNo,
		OutTradeNo:     notification.OutTradeNo,
		TradeStatus:    string(notification.TradeStatus),
		TotalAmount:    notification.TotalAmount,
		ReceiptAmount:  notification.ReceiptAmount,
		BuyerID:        notification.BuyerId,
		BuyerLogonID:   notification.BuyerLogonId,
		BuyerOpenID:    notification.BuyerOpenId,
		SellerID:       notification.SellerId,
		Subject:        notification.Subject,
		Body:           notification.Body,
		GmtCreate:      notification.GmtCreate,
		GmtPayment:     notification.GmtPayment,
		PassbackParams: notification.PassbackParams,
	}
}
