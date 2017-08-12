package lib

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/peterh/liner"
	"github.com/theherk/viper"
)

// Readline reads a single line of input
func Readline(prompt string, password bool) (string, error) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	if !password {
		return line.Prompt(prompt)
	}
	return line.PasswordPrompt(prompt)
}

// Reply generic version
type Reply map[string]json.RawMessage

// Exhaust a channel so its goroutine can end
func Exhaust(replies chan Reply) {
	for _ = range replies {
	}
}

// FeedFunc gets a feed of replies
type FeedFunc func(ctx context.Context, pageSize int) (chan Reply, error)

// Paginate the feed of replies, filtering by args
func Paginate(skipHeaders bool, args []string, feeder FeedFunc) error {
	pageSize := viper.GetInt("pagesize")
	doPagination := true
	if pageSize <= 0 {
		doPagination = false
		pageSize = 24
	}
	// Get the input channel
	ctx, cancel := context.WithCancel(context.Background())
	pages, err := feeder(ctx, pageSize)
	if err != nil {
		cancel()
		return err
	}
	// Cancel and exhaust the channel on exit
	defer func(cancel context.CancelFunc, pages chan Reply) {
		cancel()
		Exhaust(pages)
	}(cancel, pages)
	// If output is CSV-like, dump the header
	if args != nil && len(args) > 0 && !skipHeaders {
		sep := ""
		for _, name := range args {
			fmt.Print(sep, name)
			sep = ";"
		}
		fmt.Println("")
	}
	// Keep reading pages of data
	reader := bufio.NewReader(os.Stdin)
	lineno := 0
	for page := range pages {
		if args != nil && len(args) > 0 {
			sep := ""
			for _, name := range args {
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
func pick(data Reply, attrib string) string {
	parts := strings.Split(attrib, ".")
	lenp := len(parts)
	// If the string is a dot-separated path, go deep
	for i := 0; i < lenp-1; i++ {
		newData, ok := data[parts[i]]
		if !ok {
			return ""
		}
		data = make(Reply)
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
