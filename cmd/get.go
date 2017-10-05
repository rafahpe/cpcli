// Copyright © 2017 Rafael Rivero
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Make a GET request",
	Long: `Make a GET request, allows pagination.

  - If no parameters provided, list the whole endpoint data as a JSON object.
  - If some parameters are provided, they are considered attributes to dump,
    for instance "mac_address", "attributes.Username"`,
	Run: func(cmd *cobra.Command, args []string) {
		pageSize, _ := getPageSize()
		opt := getOptions(cmd, args)
		if len(opt.Args) < 1 {
			log.Print("Error: debe indicar un path para hacer el GET")
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		path, args := opt.Args[0], opt.Args[1:len(opt.Args)]
		feed, err := globalClearpass.Get(ctx, path, nil, pageSize)
		if err != nil {
			log.Print(err)
			return
		}
		// If pretty printing, output is console.
		if globalOptions.PrettyPrint {
			if err := paginate(feed, opt.SkipHeaders, args); err != nil {
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
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}
