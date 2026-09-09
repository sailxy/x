package pay

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppPayMapsRequestAndReturnsOrderString(t *testing.T) {
	svc := &fakePaymentService{appOrderString: "signed-order"}
	p := testPayWithPaymentService(svc)

	got, err := p.AppPay(context.Background(), samplePayRequest())

	require.NoError(t, err)
	assert.Equal(t, "signed-order", got.OrderString)
	assert.Equal(t, "membership", svc.appReq.Subject)
	assert.Equal(t, "order-123", svc.appReq.OutTradeNo)
	assert.Equal(t, "9.90", svc.appReq.TotalAmount)
	assert.Equal(t, "QUICK_MSECURITY_PAY", svc.appReq.ProductCode)
	assert.Equal(t, "https://example.com/alipay/notify", svc.appReq.NotifyURL)
}

func TestPagePayMapsRequestAndReturnsURL(t *testing.T) {
	svc := &fakePaymentService{pageURL: "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay"}
	p := testPayWithPaymentService(svc)

	got, err := p.PagePay(context.Background(), samplePagePayRequest())

	require.NoError(t, err)
	assert.Equal(t, "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay", got.URL)
	assert.Equal(t, "FAST_INSTANT_TRADE_PAY", svc.pageReq.ProductCode)
	assert.Equal(t, "https://example.com/pay/return", svc.pageReq.ReturnURL)
}

func TestWapPayMapsRequestAndReturnsURL(t *testing.T) {
	svc := &fakePaymentService{wapURL: "https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay"}
	p := testPayWithPaymentService(svc)

	got, err := p.WapPay(context.Background(), sampleWapPayRequest())

	require.NoError(t, err)
	assert.Equal(t, "https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay", got.URL)
	assert.Equal(t, "QUICK_WAP_WAY", svc.wapReq.ProductCode)
	assert.Equal(t, "https://example.com/pay/return", svc.wapReq.ReturnURL)
	assert.Equal(t, "https://example.com/pay/cancel", svc.wapReq.QuitURL)
}

func TestPayRequestValidation(t *testing.T) {
	p := testPayWithPaymentService(&fakePaymentService{})

	tests := []struct {
		name string
		req  PayRequest
	}{
		{name: "missing subject", req: samplePayRequest(func(r *PayRequest) { r.Subject = "" })},
		{name: "missing out trade no", req: samplePayRequest(func(r *PayRequest) { r.OutTradeNo = "" })},
		{name: "missing amount", req: samplePayRequest(func(r *PayRequest) { r.TotalAmount = "" })},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.AppPay(context.Background(), tt.req)

			assert.Nil(t, got)
			assert.Error(t, err)
		})
	}
}

func TestPageAndWapUseDefaultURLs(t *testing.T) {
	svc := &fakePaymentService{pageURL: "page", wapURL: "wap"}
	p := testPayWithPaymentService(svc)

	_, err := p.PagePay(context.Background(), PagePayRequest{PayRequest: samplePayRequest()})
	require.NoError(t, err)
	_, err = p.WapPay(context.Background(), WapPayRequest{PayRequest: samplePayRequest()})
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/pay/return", svc.pageReq.ReturnURL)
	assert.Equal(t, "https://example.com/pay/return", svc.wapReq.ReturnURL)
}

func samplePayRequest(mutators ...func(*PayRequest)) PayRequest {
	req := PayRequest{
		Subject:     "membership",
		OutTradeNo:  "order-123",
		TotalAmount: "9.90",
		ProductCode: "QUICK_MSECURITY_PAY",
		Body:        "monthly membership",
		NotifyURL:   "https://example.com/alipay/notify",
	}
	for _, mutate := range mutators {
		mutate(&req)
	}
	return req
}

func samplePagePayRequest(mutators ...func(*PagePayRequest)) PagePayRequest {
	req := PagePayRequest{
		PayRequest: samplePayRequest(func(r *PayRequest) {
			r.ProductCode = "FAST_INSTANT_TRADE_PAY"
		}),
		ReturnURL: "https://example.com/pay/return",
	}
	for _, mutate := range mutators {
		mutate(&req)
	}
	return req
}

func sampleWapPayRequest(mutators ...func(*WapPayRequest)) WapPayRequest {
	req := WapPayRequest{
		PayRequest: samplePayRequest(func(r *PayRequest) {
			r.ProductCode = "QUICK_WAP_WAY"
		}),
		ReturnURL: "https://example.com/pay/return",
		QuitURL:   "https://example.com/pay/cancel",
	}
	for _, mutate := range mutators {
		mutate(&req)
	}
	return req
}

func testPayWithPaymentService(svc paymentService) *Pay {
	return &Pay{
		appID:     "app-123",
		notifyURL: "https://example.com/alipay/notify",
		returnURL: "https://example.com/pay/return",
		payment:   svc,
	}
}

type fakePaymentService struct {
	appReq         sdkAppPayRequest
	appOrderString string
	pageReq        sdkPagePayRequest
	pageURL        string
	wapReq         sdkWapPayRequest
	wapURL         string
}

func (f *fakePaymentService) AppPay(_ context.Context, req sdkAppPayRequest) (string, error) {
	f.appReq = req
	return f.appOrderString, nil
}

func (f *fakePaymentService) PagePay(_ context.Context, req sdkPagePayRequest) (string, error) {
	f.pageReq = req
	return f.pageURL, nil
}

func (f *fakePaymentService) WapPay(_ context.Context, req sdkWapPayRequest) (string, error) {
	f.wapReq = req
	return f.wapURL, nil
}
