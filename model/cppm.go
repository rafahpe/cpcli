package model

import (
	"context"
	"fmt"
	"net/url"
)

// ErrPageTooSmall when paginated commands are givel a page size too small (<=0)
const ErrPageTooSmall = Error("Page size is too small")

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM. Returns access and refresh tokens, or error.
	// - 'address' is the CPPM server address (:port, if different from 443).
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret. Empty if client is public (trusted).
	// - 'user', 'pass' are the username and password for "password"
	//   authentication. If any of them is empty,  'client_credentials'
	//   authentication is used instead.
	Login(ctx context.Context, address, clientID, secret, user, pass string) (string, string, error)
	// Validate / Refresh credentials.
	// - 'address' is the CPPM server address (:port, if different from 443).
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret.
	// - 'token', 'refresh': the authentication and refresh tokens.
	//   If a refresh token is provided, attempt to refresh the
	//   authentication token. Otherwise, just check it is valid.
	Validate(ctx context.Context, address, clientID, secret, token, refresh string) (string, string, error)
	// Token obtained after authentication / validation
	Token() string
	// Request made to the CPPM.
	Request(method Method, path string, params Params, request interface{}) *Reply
}

// Clearpass model
type clearpass struct {
	unsafe  bool
	url     string
	token   string
	refresh string
}

// apiURL returns the URL of the API
func apiURL(address string) string {
	return fmt.Sprintf("https://%s/api", url.PathEscape(address))
}

// New creates a Clearpass object with cached IP and token
func New(address, token, refresh string, skipVerify bool) Clearpass {
	return &clearpass{
		unsafe:  skipVerify,
		url:     apiURL(address),
		token:   token,
		refresh: refresh,
	}
}

// Token implements Clearpass interface
func (c *clearpass) Token() string {
	return c.token
}

// Follow a stream of results from an endpoint.
func (c *clearpass) Request(method Method, path string, params Params, request interface{}) *Reply {
	if c.url == "" || c.token == "" {
		return NewReply(nil, ErrNotLoggedIn)
	}
	// Clone params, if any
	var defaults Params
	if params != nil && len(params) > 0 {
		defaults := make(Params)
		for k, v := range params {
			defaults[k] = v
		}
		if _, ok := defaults["limit"]; ok {
			defaults["calculate_count"] = "false"
		}
		if filter, ok := defaults["filter"]; ok {
			norm, err := normalize(filter, path)
			if err != nil {
				return NewReply(nil, err)
			}
			defaults["filter"] = norm
		}
	}
	return Request(method, c.url+"/"+path, c.token, defaults, request, c.unsafe)
}
