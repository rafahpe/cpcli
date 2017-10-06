package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rafahpe/cpcli/model"
	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string, method model.Method) {
	opt := updateOptions(cmd, args)
	if len(opt.Args) < 1 {
		log.Print("Error: missing path for ", method)
		return
	}
	// Build the filters
	var filter map[string]string
	if opt.Filter != nil && len(opt.Filter) > 0 {
		filter = make(map[string]string, len(opt.Filter))
		for _, f := range opt.Filter {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) < 2 {
				log.Print("Error: only simple equality filters supported righ now")
			} else {
				filter[parts[0]] = parts[1]
			}
		}
	}
	// Check if we are in a pipe
	items, err := readInput()
	if err != nil {
		log.Print(err)
		return
	}
	// Run just once, or iterate over the pipe
	path, format := opt.Args[0], opt.Args[1:len(opt.Args)]
	if items == nil {
		doRequest(model.GET, path, filter, nil, format)
	} else {
		for _ = range items {
			doRequest(model.GET, path, filter, nil, format)
		}
	}
}

// Read input line by line, if it is piped from somewhere.
func readInput() (chan model.Reply, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("Error stating os.Stdin: %s", err)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		replies := make(chan model.Reply)
		go func(replies chan model.Reply, scanner *bufio.Scanner) {
			defer close(replies)
			for scanner.Scan() {
				text := scanner.Text()
				reply := model.Reply{}
				if err := json.Unmarshal(([]byte)(text), &reply); err != nil {
					log.Print("Error unmarshalling '", text, "': ", err)
				} else {
					replies <- reply
				}
			}
		}(replies, scanner)
		return replies, nil
	}
	// No data being piped in
	return nil, nil
}

// Runs the request and outputs the result
func doRequest(method model.Method, path string, filter map[string]string, request interface{}, format []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed, err := globalClearpass.Do(ctx, method, path, filter, request, globalOptions.PageSize)
	if err != nil {
		log.Print(err)
		return
	}
	// If pretty printing, output is console. Use pagination.
	if globalOptions.PrettyPrint {
		if err := paginate(feed, format); err != nil {
			log.Print(err)
		}
		return
	}
	// Otherwise, output may be pipe. Use newline-delimited json.
	for reply := range feed {
		txt, err := json.Marshal(reply)
		if err != nil {
			log.Print(err)
		} else {
			fmt.Println(string(txt))
		}
	}
}
