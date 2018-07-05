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
	"github.com/spf13/cobra"
	// Using this forst instead of the original because of write file support
	// See: https://github.com/spf13/viper/pull/287
)

type webLogoutCmdP struct{ *cobra.Command }

// logoutCmd represents the login command
var webLogoutCmd = &cobra.Command{
	Use:   "weblogout",
	Short: "Log out from the CPPM HTTP interface",
	Long:  `Logs out from the the CPPM HTTP interface`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Log to Stderr without timestamps
		if err := Singleton.WebLogout(); err != nil {
			Singleton.Log.Fatal("WebLogout error: ", err)
		}
		Singleton.Log.Print("Logout completed")
	},
}

/* Logout does not seem to work
func init() {
	RootCmd.AddCommand(webLogoutCmd)
}
*/
