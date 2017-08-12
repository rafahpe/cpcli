package model

import (
	"context"

	"github.com/rafahpe/cpcli/lib"
)

// Guests enumerates all guests
func (c *clearpass) Guests(ctx context.Context, pageSize int) (chan lib.Reply, error) {
	return c.Follow(ctx, "guest", nil, pageSize)
}

// GuestByMac finds a guest information from its MAC address
func (c *clearpass) GuestByMac(ctx context.Context, pageSize int, mac lib.MAC) (chan lib.Reply, error) {
	params := map[string]string{
		"mac": mac.Hyphen(),
	}
	return c.Follow(ctx, "guest", params, pageSize)
}
