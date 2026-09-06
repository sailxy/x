package apple

import "context"

// Authenticate exchanges an authorization code and verifies the identity
// token returned by Apple using the same configured Client ID.
func (c *Client) Authenticate(ctx context.Context, request AuthenticateRequest) (*AuthenticateResult, error) {
	exchange, err := c.exchange(ctx, request.ClientID, request.Code, request.RedirectURI)
	if err != nil {
		return nil, err
	}
	identity, err := c.VerifyIdentityToken(ctx, VerifyIdentityTokenRequest{
		ClientID:           request.ClientID,
		IdentityToken:      exchange.identityToken,
		ExpectedNonceClaim: request.ExpectedNonceClaim,
	})
	if err != nil {
		return nil, err
	}
	return &AuthenticateResult{
		Identity:  *identity,
		TokenType: exchange.tokenType,
		ExpiresIn: exchange.expiresIn,
	}, nil
}
