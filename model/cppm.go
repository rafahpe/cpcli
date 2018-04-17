package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ErrPageTooSmall when paginated commands are givel a page size too small (<=0)
const ErrPageTooSmall = Error("Page size is too small")

// Params for ClearPass request
type Params struct {
	Sort     string // Field to sort by (default "-id")
	Offset   int    // Start offset (default 0)
	PageSize int    // Page size (default 25)
	Filter   Filter // Filter as a JSON object
}

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
	// Do a REST request to the CPPM.
	Do(ctx context.Context, method Method, path string, request interface{}, params Params) (Reply, error)
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
func (c *clearpass) Do(ctx context.Context, method Method, path string, request interface{}, params Params) (Reply, error) {
	if c.url == "" || c.token == "" {
		return nil, ErrNotLoggedIn
	}
	if params.PageSize < 0 {
		return nil, ErrPageTooSmall
	}
	defaults := map[string]string{
		"filter":          "{}",
		"sort":            "-id",
		"offset":          fmt.Sprintf("%d", params.Offset),
		"limit":           "25",
		"calculate_count": "false",
	}
	if params.Sort != "" {
		defaults["sort"] = params.Sort
	}
	if params.PageSize > 0 {
		defaults["limit"] = fmt.Sprintf("%d", params.PageSize)
	}
	if params.Filter != nil && len(params.Filter) > 0 {
		norm, err := normalize(params.Filter, path)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(norm)
		if err != nil {
			return nil, err
		}
		defaults["filter"] = string(val)
	}
	return follow(ctx, method, c.url+"/"+path, c.token, defaults, request, c.unsafe)
}
