package pay

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdknotify "github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	sdkpayments "github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

func TestParsePaymentNotifyMapsRequestAndTransaction(t *testing.T) {
	parser := &fakeNotifyParser{
		req: &sdknotify.Request{
			ID:           "notify-123",
			EventType:    "TRANSACTION.SUCCESS",
			ResourceType: "encrypt-resource",
			Summary:      "支付成功",
		},
		transaction: &sdkpayments.Transaction{
			TransactionId: strPtr("wx-trade-123"),
			OutTradeNo:    strPtr("order-123"),
			TradeState:    strPtr("SUCCESS"),
			Amount:        &sdkpayments.TransactionAmount{Total: int64Ptr(100)},
			Payer:         &sdkpayments.TransactionPayer{Openid: strPtr("openid")},
		},
	}
	p := &Pay{notifyParser: parser}
	request, err := http.NewRequest(http.MethodPost, "/notify", strings.NewReader("{}"))
	require.NoError(t, err)

	got, err := p.ParsePaymentNotify(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, "notify-123", got.ID)
	assert.Equal(t, "TRANSACTION.SUCCESS", got.EventType)
	assert.Equal(t, "支付成功", got.Summary)
	assert.Equal(t, "wx-trade-123", got.Order.TransactionID)
	assert.Equal(t, "order-123", got.Order.OutTradeNo)
	assert.Equal(t, "SUCCESS", got.Order.TradeState)
	assert.Equal(t, int64(100), got.Order.Amount.Total)
	assert.Equal(t, "openid", got.Order.Payer.OpenID)
}

func TestParsePaymentNotifyReturnsParserError(t *testing.T) {
	p := &Pay{notifyParser: &fakeNotifyParser{err: errors.New("invalid notification")}}
	request, err := http.NewRequest(http.MethodPost, "/notify", strings.NewReader("{}"))
	require.NoError(t, err)

	got, err := p.ParsePaymentNotify(context.Background(), request)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestParsePaymentNotifyRejectsNilRequest(t *testing.T) {
	p := &Pay{notifyParser: &fakeNotifyParser{}}

	got, err := p.ParsePaymentNotify(context.Background(), nil)

	assert.Nil(t, got)
	assert.Error(t, err)
}

type fakeNotifyParser struct {
	req         *sdknotify.Request
	transaction *sdkpayments.Transaction
	err         error
}

func (f *fakeNotifyParser) ParsePaymentNotify(context.Context, *http.Request) (*sdknotify.Request, *sdkpayments.Transaction, error) {
	return f.req, f.transaction, f.err
}
