# QQ OAuth

```go
client, _ := qq.New(qq.Config{ClientID: id, ClientSecret: secret})
result, err := client.Authenticate(ctx, code, redirectURI)
// 先按 result.Identity.UnionID 查找已绑定用户。
if isNewUser {
	profile, err := client.User(ctx, result)
	_ = profile
}
```

`AuthorizationURL` 的 `state` 由业务服务生成和消费。`AccessToken`、`RefreshToken` 默认脱敏；只有确实需要原值时调用 `Value()`。本包不管理业务用户、绑定或登录会话。
