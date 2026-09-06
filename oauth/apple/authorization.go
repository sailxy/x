package apple

import (
	"errors"
	"net/url"
	"strings"

	"github.com/sailxy/x/oauth"
)

const authorizationEndpoint = "https://appleid.apple.com/auth/authorize"

// AuthorizationURL creates an Apple Web authorization URL. Apple returns the
// authorization response with form_post, which is required when requesting
// name or email.
func (c *Client) AuthorizationURL(request AuthorizationRequest) (string, error) {
	clientID, err := c.clientID(request.ClientID, "create authorization URL")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(request.RedirectURI) == "" {
		return "", invalidInput("create authorization URL", errors.New("redirect URI is required"))
	}
	if strings.TrimSpace(request.State) == "" {
		return "", invalidInput("create authorization URL", errors.New("state is required"))
	}

	scopes, err := authorizationScopes(request.Scopes)
	if err != nil {
		return "", err
	}
	query := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {request.RedirectURI},
		"response_type": {"code"},
		"response_mode": {"form_post"},
		"state":         {request.State},
	}
	if request.Nonce != "" {
		query.Set("nonce", request.Nonce)
	}
	if len(scopes) != 0 {
		query.Set("scope", strings.Join(scopes, " "))
	}
	return authorizationEndpoint + "?" + query.Encode(), nil
}

func authorizationScopes(values []string) ([]string, error) {
	scopes := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		scope := strings.TrimSpace(value)
		if scope != "name" && scope != "email" {
			return nil, invalidInput("create authorization URL", errors.New("unsupported scope"))
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	return scopes, nil
}

func (c *Client) clientID(value, operation string) (string, error) {
	clientID := strings.TrimSpace(value)
	if clientID == "" {
		return "", invalidInput(operation, errors.New("client ID is required"))
	}
	if _, ok := c.clientIDs[clientID]; !ok {
		return "", invalidInput(operation, errors.New("client ID is not configured"))
	}
	return clientID, nil
}

func invalidInput(operation string, cause error) error {
	return oauth.NewError(oauth.ProviderApple, oauth.ErrorKindInvalidInput, operation, cause)
}
