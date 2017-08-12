package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rafahpe/cpcli/lib"
)

// Clearpass model
type clearpass struct {
	url   string
	token string
}

var globalClearpass clearpass

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM, return token or error
	Login(ip, clientID, secret string) (string, error)
	// Validate credentials
	Validate(ip, clientID, token string) error
	// Set credentials without validating
	SetCredentials(ip, clientID, secret string)
	// Get the token
	Token() string
	// Guests information, paginated
	Guests(ctx context.Context, pageSize int) (chan lib.Reply, error)
	// GuestByMac gets Guest user information from its MAC
	GuestByMac(ctx context.Context, pageSize int, mac lib.MAC) (chan lib.Reply, error)
	// Endpoints information, paginated
	Endpoints(ctx context.Context, pageSize int) (chan lib.Reply, error)
	// EndpointByMac gets Guest user information from its MAC
	EndpointByMac(ctx context.Context, pageSize int, mac lib.MAC) (chan lib.Reply, error)
}

// CPPM gets the Clearpass global instance
func CPPM() Clearpass {
	return &globalClearpass
}

// Token implements Clearpass interface
func (c *clearpass) Token() string {
	return c.token
}

// SetCredentials implements Clearpass interface
func (c *clearpass) SetCredentials(ip, clientID, token string) {
	c.url = fmt.Sprintf("https://%s:443/api/", url.PathEscape(ip))
	c.token = token
}

// Follow a stream of results from an endpoint.
// Filter is a map of fields to filter by (e.g. "mac": "00:01:02:03:04:05")
func (c *clearpass) Follow(ctx context.Context, method string, filter map[string]string, pageSize int) (chan lib.Reply, error) {
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
	if filter != nil {
		val, err := json.Marshal(filter)
		if err != nil {
			return nil, err
		}
		defaults["filter"] = string(val)
	}
	return Follow(ctx, c.url+method, c.token, defaults), nil
}
