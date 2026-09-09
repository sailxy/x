package pay

import (
	"context"
	"errors"
	"net/http"

	sdknotify "github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	sdkpayments "github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

type notifyParser interface {
	ParsePaymentNotify(context.Context, *http.Request) (*sdknotify.Request, *sdkpayments.Transaction, error)
}

type sdkNotifyParser struct {
	handler *sdknotify.Handler
}

func (p *sdkNotifyParser) ParsePaymentNotify(
	ctx context.Context,
	request *http.Request,
) (*sdknotify.Request, *sdkpayments.Transaction, error) {
	transaction := new(sdkpayments.Transaction)
	notifyReq, err := p.handler.ParseNotifyRequest(ctx, request, transaction)
	if err != nil {
		return nil, nil, err
	}
	return notifyReq, transaction, nil
}

func (p *Pay) ParsePaymentNotify(ctx context.Context, request *http.Request) (*PaymentNotify, error) {
	if request == nil {
		return nil, errors.New("request is required")
	}

	notifyReq, transaction, err := p.notifyParser.ParsePaymentNotify(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapPaymentNotify(notifyReq, transaction), nil
}

func mapPaymentNotify(req *sdknotify.Request, tx *sdkpayments.Transaction) *PaymentNotify {
	if req == nil {
		return &PaymentNotify{Order: *mapOrder(tx)}
	}
	return &PaymentNotify{
		ID:           req.ID,
		EventType:    req.EventType,
		ResourceType: req.ResourceType,
		Summary:      req.Summary,
		Order:        *mapOrder(tx),
	}
}
