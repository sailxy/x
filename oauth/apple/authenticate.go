package apple

import "context"

// Authenticate exchanges an authorization code and verifies the identity
// token returned by Apple using the same configured Client ID.
func (c *Client) Authenticate(ctx context.Context, request AuthenticateRequest) (*AuthenticateResult, error) {
	token, err := c.exchange(ctx, request.ClientID, request.Code, request.RedirectURI)
	if err != nil {
		return nil, err
	}
	identity, err := c.verifyIdentityToken(ctx, request.ClientID, token.IdentityToken, request.ExpectedNonceClaim)
	if err != nil {
		return nil, err
	}
	return &AuthenticateResult{
		Token:    token,
		Identity: *identity,
	}, nil
}
