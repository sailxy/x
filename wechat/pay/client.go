package pay

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

type Config struct {
	AppID                      string
	MchID                      string
	MchCertificateSerialNumber string
	MchPrivateKeyPath          string
	MchPrivateKeyPEM           string
	MchAPIv3Key                string
}

type Pay struct {
	appID string
	mchID string

	client        *core.Client
	notifyHandler *notify.Handler
	payment       paymentService
	order         orderService
	notifyParser  notifyParser
}

type clientParts struct {
	client        *core.Client
	notifyHandler *notify.Handler
	payment       paymentService
	order         orderService
	notifyParser  notifyParser
}

type clientFactory func(Config, *rsa.PrivateKey) (*clientParts, error)

func New(c Config) (*Pay, error) {
	return newWithFactory(c, defaultClientFactory)
}

func newWithFactory(c Config, factory clientFactory) (*Pay, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	privateKey, err := loadPrivateKey(c)
	if err != nil {
		return nil, err
	}

	parts, err := factory(c, privateKey)
	if err != nil {
		return nil, err
	}

	return &Pay{
		appID:         c.AppID,
		mchID:         c.MchID,
		client:        parts.client,
		notifyHandler: parts.notifyHandler,
		payment:       parts.payment,
		order:         parts.order,
		notifyParser:  parts.notifyParser,
	}, nil
}

func (c Config) validate() error {
	switch {
	case strings.TrimSpace(c.AppID) == "":
		return errors.New("appid is required")
	case strings.TrimSpace(c.MchID) == "":
		return errors.New("mchid is required")
	case strings.TrimSpace(c.MchCertificateSerialNumber) == "":
		return errors.New("merchant certificate serial number is required")
	case strings.TrimSpace(c.MchAPIv3Key) == "":
		return errors.New("merchant api v3 key is required")
	case strings.TrimSpace(c.MchPrivateKeyPEM) == "" && strings.TrimSpace(c.MchPrivateKeyPath) == "":
		return errors.New("merchant private key path or pem is required")
	default:
		return nil
	}
}

func loadPrivateKey(c Config) (*rsa.PrivateKey, error) {
	keyPEM := strings.TrimSpace(c.MchPrivateKeyPEM)
	if keyPEM == "" {
		b, err := os.ReadFile(c.MchPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read merchant private key: %w", err)
		}
		keyPEM = string(b)
	}

	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, errors.New("decode merchant private key pem")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse merchant private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("merchant private key must be rsa private key")
	}
	return rsaKey, nil
}

func defaultClientFactory(c Config, privateKey *rsa.PrivateKey) (*clientParts, error) {
	ctx := context.Background()
	client, err := core.NewClient(
		ctx,
		option.WithWechatPayAutoAuthCipher(
			c.MchID,
			c.MchCertificateSerialNumber,
			privateKey,
			c.MchAPIv3Key,
		),
	)
	if err != nil {
		return nil, err
	}

	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(c.MchID)
	handler, err := notify.NewRSANotifyHandler(c.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		return nil, err
	}

	return &clientParts{
		client:        client,
		notifyHandler: handler,
		payment:       newSDKPaymentService(client),
		order:         newSDKOrderService(client, c.MchID),
		notifyParser:  &sdkNotifyParser{handler: handler},
	}, nil
}
