package pay

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkpayments "github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	sdkapp "github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	sdkh5 "github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	sdknative "github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

func TestQueryOrderByOutTradeNoMapsTransaction(t *testing.T) {
	svc := &fakeOrderService{
		transaction: &sdkpayments.Transaction{
			TransactionId: strPtr("wx-trade-123"),
			OutTradeNo:    strPtr("order-123"),
			TradeState:    strPtr("SUCCESS"),
			TradeType:     strPtr("NATIVE"),
			SuccessTime:   strPtr("2026-06-14T10:00:00+08:00"),
			Amount:        &sdkpayments.TransactionAmount{Total: int64Ptr(100), PayerTotal: int64Ptr(100), Currency: strPtr("CNY")},
			Payer:         &sdkpayments.TransactionPayer{Openid: strPtr("openid")},
		},
	}
	p := testPayWithOrderService(svc)

	got, err := p.QueryOrderByOutTradeNo(context.Background(), "order-123")

	require.NoError(t, err)
	assert.Equal(t, "order-123", svc.queryOutTradeNo)
	assert.Equal(t, "wx-trade-123", got.TransactionID)
	assert.Equal(t, "order-123", got.OutTradeNo)
	assert.Equal(t, "SUCCESS", got.TradeState)
	assert.Equal(t, "NATIVE", got.TradeType)
	assert.Equal(t, "2026-06-14T10:00:00+08:00", got.SuccessTime)
	assert.Equal(t, int64(100), got.Amount.Total)
	assert.Equal(t, int64(100), got.Amount.PayerTotal)
	assert.Equal(t, "CNY", got.Amount.Currency)
	assert.Equal(t, "openid", got.Payer.OpenID)
}

func TestQueryOrderByOutTradeNoValidation(t *testing.T) {
	p := testPayWithOrderService(&fakeOrderService{})

	got, err := p.QueryOrderByOutTradeNo(context.Background(), "")

	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestCloseOrderByOutTradeNo(t *testing.T) {
	svc := &fakeOrderService{}
	p := testPayWithOrderService(svc)

	err := p.CloseOrderByOutTradeNo(context.Background(), "order-123")

	require.NoError(t, err)
	assert.Equal(t, "order-123", svc.closedOutTradeNo)
}

func TestCloseOrderByOutTradeNoValidation(t *testing.T) {
	p := testPayWithOrderService(&fakeOrderService{})

	err := p.CloseOrderByOutTradeNo(context.Background(), "")

	assert.Error(t, err)
}

func testPayWithOrderService(svc orderService) *Pay {
	return &Pay{appID: "wx-app", mchID: "mch-123", order: svc}
}

type fakeOrderService struct {
	queryOutTradeNo  string
	closedOutTradeNo string
	transaction      *sdkpayments.Transaction
}

func (f *fakeOrderService) QueryOrderByOutTradeNo(_ context.Context, outTradeNo string) (*sdkpayments.Transaction, error) {
	f.queryOutTradeNo = outTradeNo
	return f.transaction, nil
}

func (f *fakeOrderService) CloseOrderByOutTradeNo(_ context.Context, outTradeNo string) error {
	f.closedOutTradeNo = outTradeNo
	return nil
}

func (f *fakeOrderService) AppPrepay(_ context.Context, req sdkapp.PrepayRequest) (*AppPrepayResponse, error) {
	return nil, nil
}

func (f *fakeOrderService) NativePrepay(_ context.Context, req sdknative.PrepayRequest) (string, error) {
	return "", nil
}

func (f *fakeOrderService) H5Prepay(_ context.Context, req sdkh5.PrepayRequest) (string, error) {
	return "", nil
}

func strPtr(v string) *string {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
