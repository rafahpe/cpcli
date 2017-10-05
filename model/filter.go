package model

import "strings"

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
func normalize(f map[string]string, path string) (map[string]string, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		path = parts[0]
	}
	if normap, ok := normalizerTable[path]; ok {
		for k, v := range f {
			if normalizer, ok := normap[k]; ok {
				newVal, err := normalizer(v)
				if err != nil {
					return nil, err
				}
				f[k] = newVal
			}
		}
	}
	return f, nil
}
