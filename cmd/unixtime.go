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
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// unixtimeCmd represents the unixtime command
var unixtimeCmd = &cobra.Command{
	Use:   "unixtime",
	Short: "Get the current time as unix timestamp",
	Long:  "Get the current time as unix timestamp",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(int32(time.Now().Unix()))
	},
}

func init() {
	RootCmd.AddCommand(unixtimeCmd)
}
