package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rafahpe/cpcli/model"
)

// Paginate the feed of replies, filtering by args
func paginate(pages model.Reply, format []string) error {
	defer model.Exhaust(pages)
	// If output is CSV-like, dump the header
	skipHeaders := globalOptions.SkipHeaders
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
	for pages.Next() {
		page := pages.Get()
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
		if globalOptions.Paginate {
			lineno++
			if lineno >= globalOptions.PageSize {
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
	if err := pages.Error(); pages != nil {
		fmt.Println("Request ended with error: ", err)
	}
	return nil
}

// Pick particular attributes form a Reply object
func pick(data model.RawReply, attrib string) string {
	parts := strings.Split(attrib, ".")
	lenp := len(parts)
	// If the string is a dot-separated path, go deep
	for i := 0; i < lenp-1; i++ {
		newData, ok := data[parts[i]]
		if !ok {
			return ""
		}
		data = make(model.RawReply)
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
