package qq

import "github.com/sailxy/x/oauth"

// AuthenticateResult contains QQ token data and the verified stable identity.
type AuthenticateResult struct {
	Token    Token
	Identity Identity
}

// Token contains the QQ authorization-code exchange response.
type Token struct {
	AccessToken  oauth.SensitiveString `json:"access_token"`
	RefreshToken oauth.SensitiveString `json:"refresh_token"`
	ExpiresIn    string                `json:"expires_in"`
}

// Identity identifies a QQ user for the configured application and across
// applications that share a UnionID.
type Identity struct {
	OpenID  string `json:"openid"`
	UnionID string `json:"unionid"`
}

// User contains QQ user-info response fields.
type User struct {
	Ret             int    `json:"ret"`
	Msg             string `json:"msg"`
	IsLost          int    `json:"is_lost"`
	Nickname        string `json:"nickname"`
	Gender          string `json:"gender"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Year            string `json:"year"`
	Constellation   string `json:"constellation"`
	FigureURL       string `json:"figureurl"`
	FigureURL1      string `json:"figureurl_1"`
	FigureURL2      string `json:"figureurl_2"`
	FigureURLQQ     string `json:"figureurl_qq"`
	FigureURLQQ1    string `json:"figureurl_qq_1"`
	FigureURLQQ2    string `json:"figureurl_qq_2"`
	FigureURLType   string `json:"figureurl_type"`
	IsYellowVIP     string `json:"is_yellow_vip"`
	YellowVIPLevel  string `json:"yellow_vip_level"`
	VIP             string `json:"vip"`
	Level           string `json:"level"`
	IsYellowYearVIP string `json:"is_yellow_year_vip"`
}
