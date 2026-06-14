package pay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	sdk "github.com/smartwalle/alipay/v3"
)

type Config struct {
	AppID              string
	PrivateKeyPath     string
	PrivateKeyPEM      string
	AlipayPublicKeyPEM string
	NotifyURL          string
	ReturnURL          string
	Sandbox            bool
}

type Pay struct {
	appID      string
	production bool
	notifyURL  string
	returnURL  string

	client  *sdk.Client
	payment paymentService
	order   orderService
	notify  notifyParser
}

type clientParts struct {
	client  *sdk.Client
	payment paymentService
	order   orderService
	notify  notifyParser
}

type clientFactory func(Config, string) (*clientParts, error)

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
		appID:      c.AppID,
		production: !c.Sandbox,
		notifyURL:  c.NotifyURL,
		returnURL:  c.ReturnURL,
		client:     parts.client,
		payment:    parts.payment,
		order:      parts.order,
		notify:     parts.notify,
	}, nil
}

func (c Config) validate() error {
	switch {
	case strings.TrimSpace(c.AppID) == "":
		return errors.New("appid is required")
	case strings.TrimSpace(c.PrivateKeyPEM) == "" && strings.TrimSpace(c.PrivateKeyPath) == "":
		return errors.New("private key path or pem is required")
	case strings.TrimSpace(c.AlipayPublicKeyPEM) == "":
		return errors.New("alipay public key is required")
	default:
		return nil
	}
}

func loadPrivateKey(c Config) (string, error) {
	privateKey := strings.TrimSpace(c.PrivateKeyPEM)
	if privateKey != "" {
		if _, err := parsePrivateKey(privateKey); err != nil {
			return "", err
		}
		return privateKey, nil
	}

	b, err := os.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("read private key: %w", err)
	}
	privateKey = string(b)
	if _, err = parsePrivateKey(privateKey); err != nil {
		return "", err
	}
	return privateKey, nil
}

func defaultClientFactory(c Config, privateKey string) (*clientParts, error) {
	client, err := sdk.New(c.AppID, privateKey, !c.Sandbox)
	if err != nil {
		return nil, err
	}
	if err = client.LoadAliPayPublicKey(c.AlipayPublicKeyPEM); err != nil {
		return nil, err
	}

	return &clientParts{
		client:  client,
		payment: newSDKPaymentService(client),
		order:   newSDKOrderService(client),
		notify:  newSDKNotifyParser(client),
	}, nil
}

func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = strings.TrimSpace(privateKey)
	block, _ := pem.Decode([]byte(privateKey))
	var der []byte
	if block != nil {
		der = block.Bytes
	} else {
		var err error
		der, err = base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			return nil, errors.New("decode private key pem or base64 der")
		}
	}

	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key must be rsa private key")
		}
		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return key, nil
}
