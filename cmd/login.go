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

type loginCmdP struct{ *cobra.Command }

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log into the CPPM",
	Long: `Logs into the CPPM using cached credentials, or providing a client_id and client_secret to reauthenticate.

  - The ClearPass server address is provided in the 'server' configuration variable, CPPM_SERVER environment variable, or with the -h flag.
  - The OAUTH token can be provided in the 'token' configuration variable, the CPPM_TOKEN environment variable, or the -t flag.
  - If you have an OAUTH refresh token, it can be provided in the 'refresh' configuration variable, the CPPM_REFRESH environment variable, or the -r flag.
  - If OAUTH token is missing, invalid or expired, then client_id can be provided in the 'client' config variable, CPPM_CLIENT environment variable, or -c flag.
  - If you are using username/password based auth, besides the client ID, you will need to provide your username with the 'user' config variable, CPPM_USER environment variable, or -u flag`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Log to Stderr without timestamps
		if token, refresh, err := Singleton.Login(); err != nil {
			Singleton.Log.Fatal("Login error: ", err)
		} else {
			if err := Singleton.Save(token, refresh); err != nil {
				Singleton.Log.Fatal("login Error saving config data: ", err)
			}
			Singleton.Log.Println("login OK. Authorization: Bearer ", token)
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
