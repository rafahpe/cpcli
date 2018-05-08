package cmd

import (
	"encoding/json"
	"strings"

	hjson "github.com/hjson/hjson-go"
)

func (master *Master) readQuery() (map[string]string, error) {
	query := make(map[string]string)
	if master.Query != nil && len(master.Query) > 0 {
		for _, current := range master.Query {
			// If filter has json format, parse it using json.
			current = strings.TrimSpace(current)
			parts := strings.SplitN(current, "=", 2)
			if len(parts) < 2 {
				query[current] = "{'$exists': true}"
			} else {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// If val has json format, parse it using hjson and normalize back to standard json
				if strings.HasPrefix(val, "{") {
					var obj interface{}
					if err := hjson.Unmarshal([]byte(val), &obj); err != nil {
						return nil, err
					}
					norm, err := json.Marshal(obj)
					if err != nil {
						return nil, err
					}
					val = string(norm)
				}
				query[key] = val
			}
		}
	}
	return query, nil
}
