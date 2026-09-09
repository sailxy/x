package pay

import (
	"context"
	"errors"
	"strings"

	sdk "github.com/smartwalle/alipay/v3"
)

func (p *Pay) QueryOrder(ctx context.Context, outTradeNo string) (*Order, error) {
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return nil, errors.New("alipay pay: out trade no is required")
	}

	rsp, err := p.order.QueryOrder(ctx, outTradeNo)
	if err != nil {
		return nil, err
	}
	return mapOrder(rsp), nil
}

func (p *Pay) CloseOrder(ctx context.Context, outTradeNo string) error {
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return errors.New("alipay pay: out trade no is required")
	}
	return p.order.CloseOrder(ctx, outTradeNo)
}

type sdkOrderService struct {
	client *sdk.Client
}

func newSDKOrderService(client *sdk.Client) orderService {
	return &sdkOrderService{client: client}
}

func (s *sdkOrderService) QueryOrder(ctx context.Context, outTradeNo string) (*sdk.TradeQueryRsp, error) {
	return s.client.TradeQuery(ctx, sdk.TradeQuery{OutTradeNo: outTradeNo})
}

func (s *sdkOrderService) CloseOrder(ctx context.Context, outTradeNo string) error {
	_, err := s.client.TradeClose(ctx, sdk.TradeClose{OutTradeNo: outTradeNo})
	return err
}

func mapOrder(rsp *sdk.TradeQueryRsp) *Order {
	if rsp == nil {
		return nil
	}
	return &Order{
		TradeNo:      rsp.TradeNo,
		OutTradeNo:   rsp.OutTradeNo,
		TradeStatus:  string(rsp.TradeStatus),
		TotalAmount:  rsp.TotalAmount,
		BuyerLogonID: rsp.BuyerLogonId,
		BuyerUserID:  rsp.BuyerUserId,
		BuyerOpenID:  rsp.BuyerOpenId,
		SendPayDate:  rsp.SendPayDate,
	}
}
