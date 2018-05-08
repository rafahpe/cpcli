package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Normalizer function that can "normalize" some parameters
// that depend on the query. E.g. MACs for endpoints have a
// different format than MACs for guests.
type normalizer func(v string) (string, error)

type normap map[string]normalizer

// Known normalization rules
var normalizerTable = map[string]normap{
	"endpoint": normap{
		// MAC format: lower, no separators
		"mac_address": normalizer(func(v string) (string, error) {
			return strings.ToLower(string(NewMAC(v))), nil
		}),
	},
	"guest": normap{
		// MAC format: upper, hyphenated
		"mac": normalizer(func(v string) (string, error) {
			return NewMAC(v).Hyphen(), nil
		}),
	},
	"insight": normap{
		// MAC format: lower, no separators
		"mac": normalizer(func(v string) (string, error) {
			return strings.ToLower(string(NewMAC(v))), nil
		}),
	},
}

// Normalize some known parameters that have different formats
// for different endpoints, such as MAC addresses.
func normalize(filter string, path string) (string, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		path = parts[0]
	}
	// Filter must be key/value pairs
	var f map[string]interface{}
	if err := json.Unmarshal([]byte(filter), &f); err != nil {
		return "", err
	}
	normap, mustNorm := normalizerTable[path]
	result := make(map[string]interface{}, len(f))
	for key, val := range f {
		if mustNorm {
			if normalizer, ok := normap[key]; ok {
				text, ok := val.(string)
				if !ok {
					return "", fmt.Errorf("Expected string for %s but got (%T) %v", key, val, val)
				}
				norm, err := normalizer(text)
				if err != nil {
					return "", err
				}
				val = norm
			}
		}
		result[key] = val
	}
	b, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
