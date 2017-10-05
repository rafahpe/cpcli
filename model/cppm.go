package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM, return token or error
	Login(ctx context.Context, ip, clientID, secret string) (string, error)
	// Validate credentials
	Validate(ctx context.Context, ip, clientID, token string) error
	// Token obtained after authentication / validation
	Token() string
	// Perform a Get request.
	Get(ctx context.Context, path string, filter map[string]string, pageSize int) (chan Reply, error)
}

// Clearpass model
type clearpass struct {
	unsafe bool
	url    string
	token  string
}

// New creates a Clearpass object with cached IP and token
func New(ip, token string, skipVerify bool) Clearpass {
	return &clearpass{
		unsafe: skipVerify,
		url:    fmt.Sprintf("https://%s:443/api/", url.PathEscape(ip)),
		token:  token,
	}
}

// Token implements Clearpass interface
func (c *clearpass) Token() string {
	return c.token
}

// Follow a stream of results from an endpoint.
// Filter is a map of fields to filter by (e.g. "mac": "00:01:02:03:04:05")
func (c *clearpass) follow(ctx context.Context, method method, path string, filter map[string]string, request interface{}, pageSize int) (chan Reply, error) {
	if c.url == "" || c.token == "" {
		return nil, ErrNotLoggedIn
	}
	if pageSize <= 0 {
		return nil, ErrPageTooSmall
	}
	defaults := map[string]string{
		"filter":          "",
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

// Run a GET request against CPPM
func (c *clearpass) Get(ctx context.Context, path string, filter map[string]string, pageSize int) (chan Reply, error) {
	return c.follow(ctx, GET, path, filter, nil, pageSize)
}
