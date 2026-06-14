package pay

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	sdk "github.com/smartwalle/alipay/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePaymentNotifyMapsVerifiedNotification(t *testing.T) {
	parser := &fakeNotifyParser{
		notification: &sdk.Notification{
			AppId:          "app-123",
			NotifyId:       "notify-123",
			NotifyType:     sdk.NotifyTypeTradeStatusSync,
			NotifyTime:     "2026-06-14 12:01:00",
			TradeNo:        "202606142200141234",
			OutTradeNo:     "order-123",
			TradeStatus:    sdk.TradeStatusSuccess,
			TotalAmount:    "9.90",
			ReceiptAmount:  "9.90",
			BuyerId:        "2088123412341234",
			BuyerLogonId:   "buyer@example.com",
			BuyerOpenId:    "buyer-open-id",
			SellerId:       "2088999911112222",
			Subject:        "membership",
			Body:           "monthly membership",
			GmtCreate:      "2026-06-14 12:00:00",
			GmtPayment:     "2026-06-14 12:00:05",
			PassbackParams: "tenant=alpha",
		},
	}
	p := testPayWithNotifyParser(parser)
	req := newNotifyRequest(url.Values{
		"out_trade_no": {"order-123"},
		"trade_status": {"TRADE_SUCCESS"},
	})

	got, err := p.ParsePaymentNotify(req)

	require.NoError(t, err)
	assert.Equal(t, "order-123", parser.values.Get("out_trade_no"))
	assert.Equal(t, "app-123", got.AppID)
	assert.Equal(t, "notify-123", got.NotifyID)
	assert.Equal(t, "trade_status_sync", got.NotifyType)
	assert.Equal(t, "2026-06-14 12:01:00", got.NotifyTime)
	assert.Equal(t, "202606142200141234", got.TradeNo)
	assert.Equal(t, "order-123", got.OutTradeNo)
	assert.Equal(t, "TRADE_SUCCESS", got.TradeStatus)
	assert.Equal(t, "9.90", got.TotalAmount)
	assert.Equal(t, "9.90", got.ReceiptAmount)
	assert.Equal(t, "2088123412341234", got.BuyerID)
	assert.Equal(t, "buyer@example.com", got.BuyerLogonID)
	assert.Equal(t, "buyer-open-id", got.BuyerOpenID)
	assert.Equal(t, "2088999911112222", got.SellerID)
	assert.Equal(t, "membership", got.Subject)
	assert.Equal(t, "monthly membership", got.Body)
	assert.Equal(t, "2026-06-14 12:00:00", got.GmtCreate)
	assert.Equal(t, "2026-06-14 12:00:05", got.GmtPayment)
	assert.Equal(t, "tenant=alpha", got.PassbackParams)
}

func TestParsePaymentNotifyRequiresRequest(t *testing.T) {
	p := testPayWithNotifyParser(&fakeNotifyParser{})

	got, err := p.ParsePaymentNotify(nil)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func newNotifyRequest(values url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/alipay/notify", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func testPayWithNotifyParser(parser notifyParser) *Pay {
	return &Pay{notify: parser}
}

type fakeNotifyParser struct {
	values       url.Values
	notification *sdk.Notification
}

func (f *fakeNotifyParser) ParsePaymentNotify(_ context.Context, values url.Values) (*sdk.Notification, error) {
	f.values = values
	return f.notification, nil
}
