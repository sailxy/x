package wechat_test

import (
	"context"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/wechat"
)

func ExampleClient_Authenticate() {
	client, _ := wechat.New(wechat.Config{AppID: "wechat-app-id", AppSecret: oauth.SensitiveString("wechat-app-secret")})
	_, _ = client.AuthorizationURL("https://example.com/oauth/wechat/callback", "state-managed-by-backend")
	_ = context.Background() // Pass this context to Authenticate after WeChat redirects with code.
}
