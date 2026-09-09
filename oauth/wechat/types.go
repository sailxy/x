package wechat

import "github.com/sailxy/x/oauth"

// AuthenticateResult contains WeChat token data and stable identity.
type AuthenticateResult struct {
	Token    Token
	Identity Identity
}

// Token contains WeChat's authorization-code exchange response.
type Token struct {
	AccessToken  oauth.SensitiveString `json:"access_token"`
	RefreshToken oauth.SensitiveString `json:"refresh_token"`
	ExpiresIn    int64                 `json:"expires_in"`
	Scope        string                `json:"scope"`
}

// Identity identifies a WeChat user for the configured application and UnionID group.
type Identity struct {
	OpenID  string `json:"openid"`
	UnionID string `json:"unionid"`
}

// User contains WeChat user-info response fields.
type User struct {
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	UnionID    string   `json:"unionid"`
	Privilege  []string `json:"privilege"`
}
