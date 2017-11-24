package model

import (
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

// Filter to apply to GET queries
type Filter = map[string]interface{}

// Normalize some known parameters that have different formats
// for different endpoints, such as MAC addresses.
func normalize(f Filter, path string) (Filter, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		path = parts[0]
	}
	normap, mustNorm := normalizerTable[path]
	result := make(Filter, len(f))
	for key, val := range f {
		if text, ok := val.(string); ok {
			if mustNorm {
				if normalizer, ok := normap[key]; ok {
					norm, err := normalizer(text)
					if err != nil {
						return nil, err
					}
					val = norm
				}
			}
		}
		result[key] = val
	}
	return result, nil
}
