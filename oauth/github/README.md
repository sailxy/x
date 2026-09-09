# GitHub OAuth

`oauth/github` 封装 GitHub OAuth Web Application Flow。调用方只需要生成授权地址，并在回调后调用一次 `Authenticate`；授权码换 token 和获取当前 GitHub 用户由包内部完成。

## 接入

```go
package login

import (
	"context"
	"os"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/github"
)

func githubLogin(code, callbackState string) (*github.AuthenticateResult, error) {
	client, err := github.New(github.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: oauth.SensitiveString(os.Getenv("GITHUB_CLIENT_SECRET")),
	})
	if err != nil {
		return nil, err
	}

	const redirectURI = "https://example.com/oauth/github/callback"

	// 在发起登录时，由业务服务生成并保存短时、一次性的 state。
	_, err = client.AuthorizationURL(redirectURI, "state-generated-by-backend")
	if err != nil {
		return nil, err
	}

	// 回调时先由业务服务校验并消费 callbackState，再提交授权码。
	_ = callbackState
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Authenticate(ctx, code, redirectURI)
}
```

实际流程中，应把 `AuthorizationURL` 返回的地址交给浏览器跳转。换取授权码时使用的 `redirectURI` 必须和生成授权地址时一致。

## 返回数据

`Authenticate` 返回：

- `Token`：`AccessToken`、`Scope`、`TokenType`。
- `User`：`AccountID`、`NodeID`、`Login`、`Name`、`AvatarURL`、`Email`。

`User.AccountID` 是业务绑定 GitHub 身份时应使用的稳定标识。其他用户资料有值就返回，没有值时保持为空；尤其是用户未公开邮箱时，`Email` 为空不会影响登录。

`AccessToken` 是 `oauth.SensitiveString`，普通格式化和 JSON 编码会显示为 `[REDACTED]`。确实需要使用原值时调用 `result.Token.AccessToken.Value()`，使用后应立即丢弃，不能写入日志。

本包不负责生成或校验 `state`、查找或创建业务用户、合并邮箱账号、保存 token，也不签发业务登录令牌。完整的可编译示例见 [example_test.go](./example_test.go)。
