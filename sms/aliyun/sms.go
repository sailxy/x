package aliyun

import (
	"encoding/json"
	"errors"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
)

type TemplateParam struct {
	Code string `json:"code"`
}

func (t TemplateParam) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string

	SigName      string
	TemplateCode string
}

type SMS struct {
	client       *dysmsapi.Client
	sigName      string
	templateCode string
}

func New(c Config) (*SMS, error) {
	if c.AccessKeyID == "" || c.AccessKeySecret == "" || c.Endpoint == "" {
		return nil, errors.New("access key ID, access key secret, and endpoint are required")
	}
	cfg := &openapi.Config{
		AccessKeyId:     &c.AccessKeyID,
		AccessKeySecret: &c.AccessKeySecret,
		Endpoint:        &c.Endpoint,
	}

	cli, err := dysmsapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &SMS{
		client:       cli,
		sigName:      c.SigName,
		templateCode: c.TemplateCode,
	}, nil
}

func (s *SMS) SendSMS(phone, code string) (*dysmsapi.SendSmsResponse, error) {
	templateParam := TemplateParam{Code: code}.String()
	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  &phone,
		SignName:      &s.sigName,
		TemplateCode:  &s.templateCode,
		TemplateParam: &templateParam,
	}
	resp, err := s.client.SendSms(req)
	if err != nil {
		return nil, err
	}
	if *resp.Body.Code != "OK" {
		return nil, errors.New("code error")
	}
	return resp, nil
}
