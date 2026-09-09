package qq_test

import (
	"context"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/qq"
)

func ExampleClient_Authenticate() {
	client, _ := qq.New(qq.Config{ClientID: "qq-client-id", ClientSecret: oauth.SensitiveString("qq-client-secret")})
	_, _ = client.AuthorizationURL("https://example.com/oauth/qq/callback", "state-managed-by-backend")
	_ = context.Background() // Pass this context to Authenticate after QQ redirects with code.
}
