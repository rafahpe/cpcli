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
	"github.com/rafahpe/cpcli/model"
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
		runCmd(cmd, args, model.GET)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}