package pay

import (
	"context"
	"testing"

	sdk "github.com/smartwalle/alipay/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryOrderMapsResponse(t *testing.T) {
	svc := &fakeOrderService{
		queryRsp: &sdk.TradeQueryRsp{
			TradeNo:      "202606142200141234",
			OutTradeNo:   "order-123",
			TradeStatus:  sdk.TradeStatusSuccess,
			TotalAmount:  "9.90",
			BuyerLogonId: "buyer@example.com",
			BuyerUserId:  "2088123412341234",
			BuyerOpenId:  "buyer-open-id",
			SendPayDate:  "2026-06-14 12:00:00",
		},
	}
	p := testPayWithOrderService(svc)

	got, err := p.QueryOrder(context.Background(), "order-123")

	require.NoError(t, err)
	assert.Equal(t, "order-123", svc.queryOutTradeNo)
	assert.Equal(t, "202606142200141234", got.TradeNo)
	assert.Equal(t, "order-123", got.OutTradeNo)
	assert.Equal(t, "TRADE_SUCCESS", got.TradeStatus)
	assert.Equal(t, "9.90", got.TotalAmount)
	assert.Equal(t, "buyer@example.com", got.BuyerLogonID)
	assert.Equal(t, "2088123412341234", got.BuyerUserID)
	assert.Equal(t, "buyer-open-id", got.BuyerOpenID)
	assert.Equal(t, "2026-06-14 12:00:00", got.SendPayDate)
}

func TestCloseOrderUsesOutTradeNo(t *testing.T) {
	svc := &fakeOrderService{}
	p := testPayWithOrderService(svc)

	err := p.CloseOrder(context.Background(), "order-123")

	require.NoError(t, err)
	assert.Equal(t, "order-123", svc.closeOutTradeNo)
}

func TestOrderRequiresOutTradeNo(t *testing.T) {
	p := testPayWithOrderService(&fakeOrderService{})

	got, err := p.QueryOrder(context.Background(), "")
	assert.Nil(t, got)
	assert.Error(t, err)

	err = p.CloseOrder(context.Background(), " ")
	assert.Error(t, err)
}

func testPayWithOrderService(svc orderService) *Pay {
	return &Pay{order: svc}
}

type fakeOrderService struct {
	queryOutTradeNo string
	queryRsp        *sdk.TradeQueryRsp
	closeOutTradeNo string
}

func (f *fakeOrderService) QueryOrder(_ context.Context, outTradeNo string) (*sdk.TradeQueryRsp, error) {
	f.queryOutTradeNo = outTradeNo
	return f.queryRsp, nil
}

func (f *fakeOrderService) CloseOrder(_ context.Context, outTradeNo string) error {
	f.closeOutTradeNo = outTradeNo
	return nil
}
