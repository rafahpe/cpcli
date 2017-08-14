package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rafahpe/cpcli/lib"
)

// Filter types - known filters
type Filter int

const (
	// MAC filtering - filter by MAC address
	MAC Filter = iota
	// IP filtering - filter by IP address
	IP
)

// Clearpass model
type clearpass struct {
	url   string
	token string
}

// Clearpass server interface
type Clearpass interface {
	// Login into CPPM, return token or error
	Login(ip, clientID, secret string) (string, error)
	// Validate credentials
	Validate(ip, clientID, token string) error
	// Token obtained after authentication / validation
	Token() string
	// Guests information, paginated
	Guests(ctx context.Context, filters map[Filter]string, pageSize int) (chan lib.Reply, error)
	// Endpoints information, paginated
	Endpoints(ctx context.Context, filters map[Filter]string, pageSize int) (chan lib.Reply, error)
}

// NewClearpass creates a Clearpass object with cached IP and token
func NewClearpass(ip, token string) Clearpass {
	return &clearpass{
		url:   fmt.Sprintf("https://%s:443/api/", url.PathEscape(ip)),
		token: token,
	}
}

// Token implements Clearpass interface
func (c *clearpass) Token() string {
	return c.token
}

// Follow a stream of results from an endpoint.
// Filter is a map of fields to filter by (e.g. "mac": "00:01:02:03:04:05")
func (c *clearpass) follow(ctx context.Context, method string, filter map[string]string, pageSize int) (chan lib.Reply, error) {
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
	return follow(ctx, c.url+method, c.token, defaults), nil
}
