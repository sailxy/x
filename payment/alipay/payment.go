package pay

import (
	"context"
	"errors"
	"strings"

	sdk "github.com/smartwalle/alipay/v3"
)

const (
	defaultAppPayProductCode  = "QUICK_MSECURITY_PAY"
	defaultPagePayProductCode = "FAST_INSTANT_TRADE_PAY"
	defaultWapPayProductCode  = "QUICK_WAP_WAY"
)

func (p *Pay) AppPay(ctx context.Context, req PayRequest) (*AppPayResponse, error) {
	sdkReq, err := p.buildAppPayRequest(req, defaultAppPayProductCode)
	if err != nil {
		return nil, err
	}

	orderString, err := p.payment.AppPay(ctx, sdkReq)
	if err != nil {
		return nil, err
	}
	return &AppPayResponse{OrderString: orderString}, nil
}

func (p *Pay) PagePay(ctx context.Context, req PagePayRequest) (*PagePayResponse, error) {
	sdkReq, err := p.buildPagePayRequest(req)
	if err != nil {
		return nil, err
	}

	payURL, err := p.payment.PagePay(ctx, sdkReq)
	if err != nil {
		return nil, err
	}
	return &PagePayResponse{URL: payURL}, nil
}

func (p *Pay) WapPay(ctx context.Context, req WapPayRequest) (*WapPayResponse, error) {
	sdkReq, err := p.buildWapPayRequest(req)
	if err != nil {
		return nil, err
	}

	payURL, err := p.payment.WapPay(ctx, sdkReq)
	if err != nil {
		return nil, err
	}
	return &WapPayResponse{URL: payURL}, nil
}

func (p *Pay) buildAppPayRequest(req PayRequest, defaultProductCode string) (sdkAppPayRequest, error) {
	if err := validatePayRequest(req); err != nil {
		return sdkAppPayRequest{}, err
	}

	productCode := strings.TrimSpace(req.ProductCode)
	if productCode == "" {
		productCode = defaultProductCode
	}

	notifyURL := strings.TrimSpace(req.NotifyURL)
	if notifyURL == "" {
		notifyURL = p.notifyURL
	}

	return sdkAppPayRequest{
		Subject:     req.Subject,
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.TotalAmount,
		ProductCode: productCode,
		Body:        req.Body,
		NotifyURL:   notifyURL,
	}, nil
}

func (p *Pay) buildPagePayRequest(req PagePayRequest) (sdkPagePayRequest, error) {
	appReq, err := p.buildAppPayRequest(req.PayRequest, defaultPagePayProductCode)
	if err != nil {
		return sdkPagePayRequest{}, err
	}

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = p.returnURL
	}

	return sdkPagePayRequest{
		sdkAppPayRequest: appReq,
		ReturnURL:        returnURL,
	}, nil
}

func (p *Pay) buildWapPayRequest(req WapPayRequest) (sdkWapPayRequest, error) {
	appReq, err := p.buildAppPayRequest(req.PayRequest, defaultWapPayProductCode)
	if err != nil {
		return sdkWapPayRequest{}, err
	}

	returnURL := strings.TrimSpace(req.ReturnURL)
	if returnURL == "" {
		returnURL = p.returnURL
	}

	return sdkWapPayRequest{
		sdkAppPayRequest: appReq,
		ReturnURL:        returnURL,
		QuitURL:          req.QuitURL,
	}, nil
}

func validatePayRequest(req PayRequest) error {
	switch {
	case strings.TrimSpace(req.Subject) == "":
		return errors.New("alipay pay: subject is required")
	case strings.TrimSpace(req.OutTradeNo) == "":
		return errors.New("alipay pay: out trade no is required")
	case strings.TrimSpace(req.TotalAmount) == "":
		return errors.New("alipay pay: total amount is required")
	default:
		return nil
	}
}

type sdkPaymentService struct {
	client *sdk.Client
}

func newSDKPaymentService(client *sdk.Client) paymentService {
	return &sdkPaymentService{client: client}
}

func (s *sdkPaymentService) AppPay(_ context.Context, req sdkAppPayRequest) (string, error) {
	return s.client.TradeAppPay(sdk.TradeAppPay{Trade: buildSDKTrade(req, "")})
}

func (s *sdkPaymentService) PagePay(_ context.Context, req sdkPagePayRequest) (string, error) {
	payURL, err := s.client.TradePagePay(sdk.TradePagePay{Trade: buildSDKTrade(req.sdkAppPayRequest, req.ReturnURL)})
	if err != nil {
		return "", err
	}
	return payURL.String(), nil
}

func (s *sdkPaymentService) WapPay(_ context.Context, req sdkWapPayRequest) (string, error) {
	payURL, err := s.client.TradeWapPay(sdk.TradeWapPay{
		Trade:   buildSDKTrade(req.sdkAppPayRequest, req.ReturnURL),
		QuitURL: req.QuitURL,
	})
	if err != nil {
		return "", err
	}
	return payURL.String(), nil
}

func buildSDKTrade(req sdkAppPayRequest, returnURL string) sdk.Trade {
	return sdk.Trade{
		Subject:     req.Subject,
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.TotalAmount,
		ProductCode: req.ProductCode,
		Body:        req.Body,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   returnURL,
	}
}
