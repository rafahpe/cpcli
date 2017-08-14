package model

import (
	"context"
	"fmt"

	"github.com/rafahpe/cpcli/lib"
)

// Guests enumerates all guests
func (c *clearpass) Guests(ctx context.Context, filters map[Filter]string, pageSize int) (chan lib.Reply, error) {
	var params map[string]string
	if filters != nil && len(filters) > 0 {
		params = make(map[string]string)
		for key, val := range filters {
			switch key {
			case MAC:
				// MAC format: upper, hyphenated
				params["mac"] = lib.NewMAC(val).Hyphen()
			default:
				return nil, fmt.Errorf("Parameter %d = '%s' not supported", key, val)
			}
		}
	}
	return c.follow(ctx, "guest", params, pageSize)
}
