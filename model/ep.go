package model

import (
	"context"
	"strings"

	"github.com/rafahpe/cpcli/lib"
)

// Endpoints enumerates all endpoints
func (c *clearpass) Endpoints(ctx context.Context, pageSize int) (chan lib.Reply, error) {
	return c.Follow(ctx, "endpoint", nil, pageSize)
}

// EndpointByMac finds an endpoint information from its MAC address
func (c *clearpass) EndpointByMac(ctx context.Context, pageSize int, mac lib.MAC) (chan lib.Reply, error) {
	params := map[string]string{
		"mac_address": strings.ToLower(string(mac)),
	}
	return c.Follow(ctx, "endpoint", params, pageSize)
}
