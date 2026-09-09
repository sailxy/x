package pns

import (
	"errors"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/client"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
}

type PNS struct {
	client *dypnsapi.Client
}

func New(c Config) (*PNS, error) {
	if c.AccessKeyID == "" || c.AccessKeySecret == "" || c.Endpoint == "" {
		return nil, errors.New("access key ID, access key secret, and endpoint are required")
	}
	cfg := &openapi.Config{
		AccessKeyId:     &c.AccessKeyID,
		AccessKeySecret: &c.AccessKeySecret,
		Endpoint:        &c.Endpoint,
	}

	cli, err := dypnsapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &PNS{
		client: cli,
	}, nil
}

func (u *PNS) GetMobile(token string) (*dypnsapi.GetMobileResponse, error) {
	req := &dypnsapi.GetMobileRequest{
		AccessToken: &token,
	}
	resp, err := u.client.GetMobile(req)
	if err != nil {
		return nil, err
	}

	if *resp.Body.Code != "OK" {
		return nil, errors.New("code error")
	}
	return resp, nil
}
