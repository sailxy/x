package pay

import (
	"context"
	"errors"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	sdkpayments "github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	sdkapp "github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
)

type orderService interface {
	QueryOrderByOutTradeNo(context.Context, string) (*sdkpayments.Transaction, error)
	CloseOrderByOutTradeNo(context.Context, string) error
}

type sdkOrderService struct {
	app   sdkapp.AppApiService
	mchID string
}

func newSDKOrderService(client *core.Client, mchID string) orderService {
	return &sdkOrderService{
		app:   sdkapp.AppApiService{Client: client},
		mchID: mchID,
	}
}

func (s *sdkOrderService) QueryOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*sdkpayments.Transaction, error) {
	resp, _, err := s.app.QueryOrderByOutTradeNo(ctx, sdkapp.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(outTradeNo),
		Mchid:      core.String(s.mchID),
	})
	return resp, err
}

func (s *sdkOrderService) CloseOrderByOutTradeNo(ctx context.Context, outTradeNo string) error {
	_, err := s.app.CloseOrder(ctx, sdkapp.CloseOrderRequest{
		OutTradeNo: core.String(outTradeNo),
		Mchid:      core.String(s.mchID),
	})
	return err
}

func (p *Pay) QueryOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*Order, error) {
	if strings.TrimSpace(outTradeNo) == "" {
		return nil, errors.New("out trade no is required")
	}
	tx, err := p.order.QueryOrderByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, err
	}
	return mapOrder(tx), nil
}

func (p *Pay) CloseOrderByOutTradeNo(ctx context.Context, outTradeNo string) error {
	if strings.TrimSpace(outTradeNo) == "" {
		return errors.New("out trade no is required")
	}
	return p.order.CloseOrderByOutTradeNo(ctx, outTradeNo)
}

func mapOrder(tx *sdkpayments.Transaction) *Order {
	if tx == nil {
		return &Order{}
	}

	order := &Order{
		TransactionID:  stringValue(tx.TransactionId),
		OutTradeNo:     stringValue(tx.OutTradeNo),
		TradeState:     stringValue(tx.TradeState),
		TradeStateDesc: stringValue(tx.TradeStateDesc),
		TradeType:      stringValue(tx.TradeType),
		SuccessTime:    stringValue(tx.SuccessTime),
	}
	if tx.Amount != nil {
		order.Amount = OrderAmount{
			Total:         int64Value(tx.Amount.Total),
			PayerTotal:    int64Value(tx.Amount.PayerTotal),
			Currency:      stringValue(tx.Amount.Currency),
			PayerCurrency: stringValue(tx.Amount.PayerCurrency),
		}
	}
	if tx.Payer != nil {
		order.Payer = Payer{OpenID: stringValue(tx.Payer.Openid)}
	}
	return order
}

func int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
