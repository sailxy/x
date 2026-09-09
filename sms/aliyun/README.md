# 阿里云短信

该包通过阿里云短信服务发送验证码。

```go
client, err := sms.New(sms.Config{
	AccessKeyID:     "...",
	AccessKeySecret: "...",
	Endpoint:        "...",
	SigName:         "...",
	TemplateCode:    "...",
})
if err != nil {
	return err
}

_, err = client.SendSMS(phone, code)
```

旧导入路径 `github.com/sailxy/x/aliyun/sms` 已移除，请改用 `github.com/sailxy/x/sms/aliyun`。
