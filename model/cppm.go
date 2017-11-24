package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM. Returns access and refresh tokens, or error.
	// - 'address' is the CPPM server address.
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret. Empty if client is public (trusted).
	// - 'user', 'pass' are the username and password for "password"
	//   authentication. If any of them is empty,  'client_credentials'
	//   authentication is used instead.
	Login(ctx context.Context, address, clientID, secret, user, pass string) (string, string, error)
	// Validate / Refresh credentials.
	// "secret" is only needed when request_type is 'password' and
	// - 'address' is the CPPM server address.
	// - 'clientID' is the OAuth2 Client ID
	// - 'secret' is the OAuth2 Client secret. Required if
	//   authentication type is "password". Othewise, leave blank.
	// - 'token', 'refresh': the authentication and refresh tokens.
	//   If a refresh token is provided, attempt to refresh the
	//   authentication token. Otherwise, just check it is valid.
	Validate(ctx context.Context, address, clientID, secret, token, refresh string) (string, string, error)
	// Token obtained after authentication / validation
	Token() string
	// Do a REST request to the CPPM.
	Do(ctx context.Context, method Method, path string, filter Filter, request interface{}, pageSize int) (chan Reply, error)
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
	return fmt.Sprintf("https://%s:443/api/", url.PathEscape(address))
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
// Filter is a map of fields to filter by (e.g. "mac": "00:01:02:03:04:05")
func (c *clearpass) Do(ctx context.Context, method Method, path string, filter Filter, request interface{}, pageSize int) (chan Reply, error) {
	if c.url == "" || c.token == "" {
		return nil, ErrNotLoggedIn
	}
	if pageSize <= 0 {
		return nil, ErrPageTooSmall
	}
	defaults := map[string]string{
		"filter":          "{}",
		"sort":            "-id",
		"offset":          "0",
		"limit":           fmt.Sprintf("%d", pageSize),
		"calculate_count": "false",
	}
	if filter != nil && len(filter) > 0 {
		norm, err := normalize(filter, path)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(norm)
		if err != nil {
			return nil, err
		}
		defaults["filter"] = string(val)
	}
	return follow(ctx, method, c.url+path, c.token, defaults, request, c.unsafe)
}
