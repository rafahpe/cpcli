package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rafahpe/cpcli/model"
	"github.com/theherk/viper"
)

// DefaultPageSize is the default page size for pagination
const DefaultPageSize = 24

// get page size
func getPageSize() (int, bool) {
	pageSize := viper.GetInt("pagesize")
	if pageSize <= 0 {
		return DefaultPageSize, false
	}
	return pageSize, true
}

// Paginate the feed of replies, filtering by args
func paginate(pages chan model.Reply, skipHeaders bool, format []string) error {
	defer model.Exhaust(pages)
	// If output is CSV-like, dump the header
	if format != nil && len(format) > 0 && !skipHeaders {
		sep := ""
		for _, name := range format {
			fmt.Print(sep, name)
			sep = ";"
		}
		fmt.Println("")
	}
	// Keep reading pages of data
	reader := bufio.NewReader(os.Stdin)
	lineno := 0
	pageSize, doPagination := getPageSize()
	for page := range pages {
		if format != nil && len(format) > 0 {
			sep := ""
			for _, name := range format {
				fmt.Print(sep, pick(page, name))
				sep = ";"
			}
			fmt.Println("")
		} else {
			output, err := json.MarshalIndent(page, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
		}
		if doPagination {
			lineno++
			if lineno >= pageSize {
				log.Println("Press q to quit, enter to continue")
				r, _, err := reader.ReadRune()
				if err != nil {
					return err
				}
				if strings.ToLower(string(r)) == "q" {
					return nil
				}
				lineno = 0
			}
		}
	}
	return nil
}

// Pick particular attributes form a Reply object
func pick(data model.Reply, attrib string) string {
	parts := strings.Split(attrib, ".")
	lenp := len(parts)
	// If the string is a dot-separated path, go deep
	for i := 0; i < lenp-1; i++ {
		newData, ok := data[parts[i]]
		if !ok {
			return ""
		}
		data = make(model.Reply)
		if err := json.Unmarshal(newData, &data); err != nil {
			return ""
		}
	}
	result, ok := data[parts[lenp-1]]
	if !ok {
		return ""
	}
	repr, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(repr)
}