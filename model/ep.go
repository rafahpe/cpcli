package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/rafahpe/cpcli/lib"
)

// Endpoints enumerates all endpoints
func (c *clearpass) Endpoints(ctx context.Context, filters map[Filter]string, pageSize int) (chan lib.Reply, error) {
	var params map[string]string
	if filters != nil && len(filters) > 0 {
		params = make(map[string]string)
		for key, val := range filters {
			switch key {
			case MAC:
				// MAC format: lower, no separators
				params["mac_address"] = strings.ToLower(string(lib.NewMAC(val)))
			default:
				return nil, fmt.Errorf("Parameter %d = '%s' not supported", key, val)
			}
		}
	}
	return c.follow(ctx, "endpoint", params, pageSize)
}
