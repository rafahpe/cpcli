package model

import (
	"encoding/json"
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

// Normalize some known parameters that change format, such as MAC addresses.
func normalize(f map[string]string, path string) (map[string]interface{}, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		path = parts[0]
	}
	normap, mustNorm := normalizerTable[path]
	result := make(map[string]interface{}, len(f))
	for k, v := range f {
		var newVal interface{}
		// If value looks like json, parse it.
		if v[0] == '{' || v[0] == '[' || v[0] == '\'' || v[0] == '"' {
			if err := json.Unmarshal(([]byte)(v), &newVal); err != nil {
				newVal = v
			}
		} else {
			if mustNorm {
				if normalizer, ok := normap[k]; ok {
					norm, err := normalizer(v)
					if err != nil {
						return nil, err
					}
					v = norm
				}
			}
			newVal = v
		}
		result[k] = newVal
	}
	return result, nil
}
