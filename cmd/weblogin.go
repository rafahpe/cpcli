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

type webLoginCmdP struct{ *cobra.Command }

// loginCmd represents the login command
var webLoginCmd = &cobra.Command{
	Use:   "weblogin",
	Short: "Log into the CPPM HTTP interface",
	Long: `Logs into the CPPM HTTP interface using cached cookies, or providing an username and password to reauthenticate.

  - The ClearPass server address is provided in the 'server' configuration variable, CPPM_SERVER environment variable, or with the -h flag.
  - The username is provided with the 'user' config variable, CPPM_USER environment variable, or -u flag`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Log to Stderr without timestamps
		if cookie, err := Singleton.WebLogin(); err != nil {
			Singleton.Log.Fatal("WebLogin error: ", err)
		} else {
			if err := Singleton.SaveCookie(cookie); err != nil {
				Singleton.Log.Fatal("WebLogin Error saving config data: ", err)
			}
			Singleton.Log.Println("WebLogin OK. Cookie: ", cookie)
		}
	},
}

func init() {
	RootCmd.AddCommand(webLoginCmd)
}
