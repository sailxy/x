package pay

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkapp "github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	sdkh5 "github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	sdknative "github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

func TestAppPrepayMapsRequestAndResponse(t *testing.T) {
	svc := &fakePaymentService{
		appResp: &AppPrepayResponse{
			PrepayID:  "prepay-app",
			PartnerID: "mch-123",
			TimeStamp: "1710000000",
			NonceStr:  "nonce",
			Package:   "Sign=WXPay",
			Sign:      "signed",
		},
	}
	p := testPayWithPaymentService(svc)

	got, err := p.AppPrepay(context.Background(), samplePrepayRequest())

	require.NoError(t, err)
	assert.Equal(t, "prepay-app", got.PrepayID)
	assert.Equal(t, "mch-123", got.PartnerID)
	require.NotNil(t, svc.appReq)
	assert.Equal(t, "wx-app", *svc.appReq.Appid)
	assert.Equal(t, "mch-123", *svc.appReq.Mchid)
	assert.Equal(t, "membership", *svc.appReq.Description)
	assert.Equal(t, "order-123", *svc.appReq.OutTradeNo)
	assert.Equal(t, "https://example.com/wxpay/notify", *svc.appReq.NotifyUrl)
	assert.Equal(t, int64(100), *svc.appReq.Amount.Total)
}

func TestNativePrepayReturnsCodeURL(t *testing.T) {
	svc := &fakePaymentService{nativeCodeURL: "weixin://wxpay/bizpayurl?pr=abc"}
	p := testPayWithPaymentService(svc)

	got, err := p.NativePrepay(context.Background(), samplePrepayRequest())

	require.NoError(t, err)
	assert.Equal(t, "weixin://wxpay/bizpayurl?pr=abc", got.CodeURL)
	require.NotNil(t, svc.nativeReq)
	assert.Equal(t, "wx-app", *svc.nativeReq.Appid)
	assert.Equal(t, "mch-123", *svc.nativeReq.Mchid)
	assert.Equal(t, int64(100), *svc.nativeReq.Amount.Total)
}

func TestH5PrepayReturnsH5URLAndMapsScene(t *testing.T) {
	svc := &fakePaymentService{h5URL: "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=abc"}
	p := testPayWithPaymentService(svc)

	got, err := p.H5Prepay(context.Background(), sampleH5PrepayRequest())

	require.NoError(t, err)
	assert.Equal(t, "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=abc", got.H5URL)
	require.NotNil(t, svc.h5Req)
	assert.Equal(t, "wx-app", *svc.h5Req.Appid)
	assert.Equal(t, "mch-123", *svc.h5Req.Mchid)
	assert.Equal(t, "Wap", *svc.h5Req.SceneInfo.H5Info.Type)
	assert.Equal(t, "Example", *svc.h5Req.SceneInfo.H5Info.AppName)
	assert.Equal(t, "https://example.com", *svc.h5Req.SceneInfo.H5Info.AppUrl)
}

func TestPrepayValidation(t *testing.T) {
	p := testPayWithPaymentService(&fakePaymentService{})

	tests := []struct {
		name string
		req  PrepayRequest
	}{
		{name: "missing description", req: samplePrepayRequest(func(r *PrepayRequest) { r.Description = "" })},
		{name: "missing out trade no", req: samplePrepayRequest(func(r *PrepayRequest) { r.OutTradeNo = "" })},
		{name: "missing notify url", req: samplePrepayRequest(func(r *PrepayRequest) { r.NotifyURL = "" })},
		{name: "missing amount", req: samplePrepayRequest(func(r *PrepayRequest) { r.Total = 0 })},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.NativePrepay(context.Background(), tt.req)

			assert.Nil(t, got)
			assert.Error(t, err)
		})
	}
}

func TestH5PrepayValidationRequiresScene(t *testing.T) {
	p := testPayWithPaymentService(&fakePaymentService{})
	req := sampleH5PrepayRequest(func(r *H5PrepayRequest) {
		r.SceneType = ""
	})

	got, err := p.H5Prepay(context.Background(), req)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func samplePrepayRequest(mutators ...func(*PrepayRequest)) PrepayRequest {
	req := PrepayRequest{
		Description: "membership",
		OutTradeNo:  "order-123",
		NotifyURL:   "https://example.com/wxpay/notify",
		Total:       100,
		Currency:    "CNY",
		Attach:      "attach-data",
		GoodsTag:    "tag",
	}
	for _, mutate := range mutators {
		mutate(&req)
	}
	return req
}

func sampleH5PrepayRequest(mutators ...func(*H5PrepayRequest)) H5PrepayRequest {
	req := H5PrepayRequest{
		PrepayRequest: samplePrepayRequest(),
		PayerClientIP: "203.0.113.1",
		SceneType:     "Wap",
		AppName:       "Example",
		AppURL:        "https://example.com",
	}
	for _, mutate := range mutators {
		mutate(&req)
	}
	return req
}

func testPayWithPaymentService(svc paymentService) *Pay {
	return &Pay{appID: "wx-app", mchID: "mch-123", payment: svc}
}

type fakePaymentService struct {
	appReq        *sdkapp.PrepayRequest
	appResp       *AppPrepayResponse
	nativeReq     *sdknative.PrepayRequest
	nativeCodeURL string
	h5Req         *sdkh5.PrepayRequest
	h5URL         string
}

func (f *fakePaymentService) AppPrepay(_ context.Context, req sdkapp.PrepayRequest) (*AppPrepayResponse, error) {
	f.appReq = &req
	return f.appResp, nil
}

func (f *fakePaymentService) NativePrepay(_ context.Context, req sdknative.PrepayRequest) (string, error) {
	f.nativeReq = &req
	return f.nativeCodeURL, nil
}

func (f *fakePaymentService) H5Prepay(_ context.Context, req sdkh5.PrepayRequest) (string, error) {
	f.h5Req = &req
	return f.h5URL, nil
}
