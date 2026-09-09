# 阿里云号码认证

该包通过阿里云号码认证服务，用客户端返回的 access token 换取手机号。

```go
client, err := pns.New(pns.Config{
	AccessKeyID:     "...",
	AccessKeySecret: "...",
	Endpoint:        "...",
})
if err != nil {
	return err
}

mobile, err := client.GetMobile(accessToken)
```

旧导入路径 `github.com/sailxy/x/aliyun/pns` 已移除，请改用 `github.com/sailxy/x/phone/aliyun`。
