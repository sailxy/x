package pay

import (
	"context"
	"errors"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	sdkapp "github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	sdkh5 "github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	sdknative "github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

type paymentService interface {
	AppPrepay(context.Context, sdkapp.PrepayRequest) (*AppPrepayResponse, error)
	NativePrepay(context.Context, sdknative.PrepayRequest) (string, error)
	H5Prepay(context.Context, sdkh5.PrepayRequest) (string, error)
}

type sdkPaymentService struct {
	app    sdkapp.AppApiService
	native sdknative.NativeApiService
	h5     sdkh5.H5ApiService
}

func newSDKPaymentService(client *core.Client) paymentService {
	return &sdkPaymentService{
		app:    sdkapp.AppApiService{Client: client},
		native: sdknative.NativeApiService{Client: client},
		h5:     sdkh5.H5ApiService{Client: client},
	}
}

func (s *sdkPaymentService) AppPrepay(ctx context.Context, req sdkapp.PrepayRequest) (*AppPrepayResponse, error) {
	resp, _, err := s.app.PrepayWithRequestPayment(ctx, req)
	if err != nil {
		return nil, err
	}
	return &AppPrepayResponse{
		PrepayID:  stringValue(resp.PrepayId),
		PartnerID: stringValue(resp.PartnerId),
		TimeStamp: stringValue(resp.TimeStamp),
		NonceStr:  stringValue(resp.NonceStr),
		Package:   stringValue(resp.Package),
		Sign:      stringValue(resp.Sign),
	}, nil
}

func (s *sdkPaymentService) NativePrepay(ctx context.Context, req sdknative.PrepayRequest) (string, error) {
	resp, _, err := s.native.Prepay(ctx, req)
	if err != nil {
		return "", err
	}
	return stringValue(resp.CodeUrl), nil
}

func (s *sdkPaymentService) H5Prepay(ctx context.Context, req sdkh5.PrepayRequest) (string, error) {
	resp, _, err := s.h5.Prepay(ctx, req)
	if err != nil {
		return "", err
	}
	return stringValue(resp.H5Url), nil
}

func (p *Pay) AppPrepay(ctx context.Context, req PrepayRequest) (*AppPrepayResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	return p.payment.AppPrepay(ctx, p.appPrepayRequest(req))
}

func (p *Pay) NativePrepay(ctx context.Context, req PrepayRequest) (*NativePrepayResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	codeURL, err := p.payment.NativePrepay(ctx, p.nativePrepayRequest(req))
	if err != nil {
		return nil, err
	}
	return &NativePrepayResponse{CodeURL: codeURL}, nil
}

func (p *Pay) H5Prepay(ctx context.Context, req H5PrepayRequest) (*H5PrepayResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	h5URL, err := p.payment.H5Prepay(ctx, p.h5PrepayRequest(req))
	if err != nil {
		return nil, err
	}
	return &H5PrepayResponse{H5URL: h5URL}, nil
}

func (r PrepayRequest) validate() error {
	switch {
	case strings.TrimSpace(r.Description) == "":
		return errors.New("description is required")
	case strings.TrimSpace(r.OutTradeNo) == "":
		return errors.New("out trade no is required")
	case strings.TrimSpace(r.NotifyURL) == "":
		return errors.New("notify url is required")
	case r.Total <= 0:
		return errors.New("total must be greater than zero")
	default:
		return nil
	}
}

func (r H5PrepayRequest) validate() error {
	if err := r.PrepayRequest.validate(); err != nil {
		return err
	}
	switch {
	case strings.TrimSpace(r.PayerClientIP) == "":
		return errors.New("payer client ip is required")
	case strings.TrimSpace(r.SceneType) == "":
		return errors.New("h5 scene type is required")
	default:
		return nil
	}
}

func (p *Pay) appPrepayRequest(req PrepayRequest) sdkapp.PrepayRequest {
	return sdkapp.PrepayRequest{
		Appid:       core.String(p.appID),
		Mchid:       core.String(p.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		TimeExpire:  req.TimeExpire,
		Attach:      optionalString(req.Attach),
		NotifyUrl:   core.String(req.NotifyURL),
		GoodsTag:    optionalString(req.GoodsTag),
		Amount: &sdkapp.Amount{
			Total:    core.Int64(req.Total),
			Currency: optionalCurrency(req.Currency),
		},
	}
}

func (p *Pay) nativePrepayRequest(req PrepayRequest) sdknative.PrepayRequest {
	return sdknative.PrepayRequest{
		Appid:       core.String(p.appID),
		Mchid:       core.String(p.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		TimeExpire:  req.TimeExpire,
		Attach:      optionalString(req.Attach),
		NotifyUrl:   core.String(req.NotifyURL),
		GoodsTag:    optionalString(req.GoodsTag),
		Amount: &sdknative.Amount{
			Total:    core.Int64(req.Total),
			Currency: optionalCurrency(req.Currency),
		},
	}
}

func (p *Pay) h5PrepayRequest(req H5PrepayRequest) sdkh5.PrepayRequest {
	return sdkh5.PrepayRequest{
		Appid:       core.String(p.appID),
		Mchid:       core.String(p.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		TimeExpire:  req.TimeExpire,
		Attach:      optionalString(req.Attach),
		NotifyUrl:   core.String(req.NotifyURL),
		GoodsTag:    optionalString(req.GoodsTag),
		Amount: &sdkh5.Amount{
			Total:    core.Int64(req.Total),
			Currency: optionalCurrency(req.Currency),
		},
		SceneInfo: &sdkh5.SceneInfo{
			PayerClientIp: core.String(req.PayerClientIP),
			H5Info: &sdkh5.H5Info{
				Type:        core.String(req.SceneType),
				AppName:     optionalString(req.AppName),
				AppUrl:      optionalString(req.AppURL),
				BundleId:    optionalString(req.BundleID),
				PackageName: optionalString(req.PackageName),
			},
		},
	}
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return core.String(v)
}

func optionalCurrency(v string) *string {
	if strings.TrimSpace(v) == "" {
		return core.String(CurrencyCNY)
	}
	return core.String(v)
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
