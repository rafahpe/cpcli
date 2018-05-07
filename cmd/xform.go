package cmd

import (
	"fmt"
	"strings"

	hjson "github.com/hjson/hjson-go"
	"github.com/rafahpe/cpcli/model"
)

func (master *Master) readJSON(specs []string) (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if specs != nil && len(specs) > 0 {
		for _, current := range master.Filter {
			// If filter has json format, parse it using json.
			current = strings.TrimSpace(current)
			if strings.HasPrefix(current, "{") {
				var partial model.Filter
				if err := hjson.Unmarshal([]byte(current), &partial); err != nil {
					return nil, fmt.Errorf("Wrong filter format: %s", err.Error())
				}
				for k, v := range partial {
					filter[k] = v
				}
			} else {
				// If not json, consider it a key / value pair.
				// If the value is not specified, then only require that
				// the item exists,
				parts := strings.SplitN(current, "=", 2)
				if len(parts) < 2 {
					filter[current] = map[string]bool{"$exists": true}
				} else {
					filter[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return filter, nil
}
