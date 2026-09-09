# Sign in with Apple

`oauth/apple` 封装 Sign in with Apple 的授权码交换和 identity token 验证。调用方提交一次授权结果，`Authenticate` 会在内部生成 client secret、换取 Apple token、获取并缓存公钥，然后返回完整 token 数据和经过验证的 Apple 身份。

## 接入

```go
package login

import (
	"context"
	"os"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/apple"
)

func appleLogin(code string) (*apple.AuthenticateResult, error) {
	privateKey, err := os.ReadFile(os.Getenv("APPLE_PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	clientID := os.Getenv("APPLE_WEB_CLIENT_ID")
	client, err := apple.New(apple.Config{
		TeamID:        os.Getenv("APPLE_TEAM_ID"),
		ClientIDs:     []string{clientID},
		KeyID:         os.Getenv("APPLE_KEY_ID"),
		PrivateKeyPEM: oauth.SensitiveString(privateKey),
	})
	if err != nil {
		return nil, err
	}

	const redirectURI = "https://example.com/oauth/apple/callback"
	const expectedNonceClaim = "nonce-associated-with-backend-state"

	// Web 登录时，把返回地址交给浏览器跳转。state 和 nonce 由业务服务管理。
	_, err = client.AuthorizationURL(apple.AuthorizationRequest{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		State:       "state-generated-by-backend",
		Nonce:       expectedNonceClaim,
		Scopes:      []string{"name", "email"},
	})
	if err != nil {
		return nil, err
	}

	// 回调时先校验并消费 state，再提交授权码。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Authenticate(ctx, apple.AuthenticateRequest{
		ClientID:           clientID,
		Code:               code,
		RedirectURI:        redirectURI,
		ExpectedNonceClaim: expectedNonceClaim,
	})
}
```

`ClientID` 必须由服务端从 `Config.ClientIDs` 中选择，不能直接信任客户端提交的 audience。Web 流程交换授权码时使用的 `redirectURI` 必须和授权请求一致。原生客户端如果对原始 nonce 做了哈希，应把 identity token 中预期出现的哈希值传给 `ExpectedNonceClaim`；未启用 nonce 时该字段留空。

## 返回数据

`Authenticate` 返回：

- `Identity`：经过签名、issuer、audience、有效期和 nonce 校验的 `Subject`、`Audience`、`Email`、`EmailVerified`、`IsPrivateEmail`、`Nonce`。
- `Token`：Apple 返回的 `AccessToken`、`RefreshToken`、`IdentityToken`、`TokenType`、`ExpiresIn`。

业务绑定 Apple 身份时应使用 `(Identity.Audience, Identity.Subject)` 作为身份键。不能使用未经本包验证的客户端 `user` 或 email 作为身份键。

Apple 的姓名只会在用户首次授权时出现在回调的 `user` 表单字段中，不在 identity token 内，因此不属于 `AuthenticateResult`；需要姓名的业务应在回调层单独解析并保存这个一次性字段。

三个 token 都是 `oauth.SensitiveString`，普通格式化和 JSON 编码会显示为 `[REDACTED]`。确实需要原值时显式调用 `.Value()`。只有业务明确需要刷新或撤销能力时才应保存相应 token，且任何 token、授权码、私钥和 client secret 都不能写入日志。

本包不负责生成或消费 `state`/nonce、查找或创建业务用户、合并账号，也不签发业务登录令牌。完整的可编译示例见 [example_test.go](./example_test.go)。
