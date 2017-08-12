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
	"fmt"
	"log"

	"github.com/rafahpe/cpcli/lib"
	"github.com/rafahpe/cpcli/model"
	"github.com/spf13/cobra"
	// Using this forst instead of the original because of write file support
	// See: https://github.com/spf13/viper/pull/287
)

// loginCmd represents the login command
var guestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List guest users from CPPM",
	Long: `List all the guest users of CPPM, allows pagination

  - If no parameters provided, list the whole guest user data as a JSON object.
  - If some parameters are provided, they are considered attributes to dump`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := login(); err != nil {
			fmt.Println("guestList Error: ", err)
			return
		}
		opt := getOptions(cmd, args)
		err := lib.Paginate(opt.SkipHeaders, opt.Args, func(ctx context.Context, pageSize int) (chan lib.Reply, error) {
			if opt.Mac != "" {
				return model.CPPM().GuestByMac(ctx, opt.PageSize, opt.Mac)
			}
			return model.CPPM().Guests(ctx, pageSize)
		})
		if err != nil {
			log.Print(err)
		}
	},
}

func init() {
	guestCmd.AddCommand(guestListCmd)
}
