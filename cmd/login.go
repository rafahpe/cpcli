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

	"github.com/rafahpe/cpcli/lib"
	"github.com/spf13/cobra"
	// Using this forst instead of the original because of write file support
	// See: https://github.com/spf13/viper/pull/287
	"github.com/theherk/viper"
)

// ErrMissingserver returned when there is no server to log in
const ErrMissingserver = Error("Missing CPPM server name or IP address")

// ErrMissingCreds returned when there are no credentials to log in
const ErrMissingCreds = Error("Missing credentials to log into CPPM")

// ErrInvalidCreds returned when not allowed to log in
const ErrInvalidCreds = Error("Credentials are invalid or expired")

type loginCmdP struct{ *cobra.Command }

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log into the CPPM",
	Long: `Logs into the CPPM using cached credentials, or providing
a client_id and client_secret to reauthenticate.

  - The Clearpass server address is provided in the 'server' configuration
    variable, CPPM_server environment variable, or with the -h flag.
  - The OAUTH token can be provided in the 'token' configuration variable,
	or the CPPM_TOKEN environment variable, or the -t flag.
  - If OAUTH token is missing, invalid or expired, then client_id can be
	provided in the 'client' config variable, CPPM_CLIENT environment
	variable, or -c flag.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		p := loginCmdP{cmd}
		if token, err := p.run(); err != nil {
			fmt.Println("login Error: ", err)
		} else {
			if err := p.save(token); err != nil {
				fmt.Println("login Error saving config data: ", err)
			}
			fmt.Println("login OK. Token: ", globalClearpass.Token())
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

// Run login
func (p loginCmdP) run() (string, error) {
	server := viper.GetString("server")
	if server == "" {
		return "", ErrMissingserver
	}
	client := viper.GetString("client")
	if client == "" {
		return "", ErrMissingCreds
	}
	token := viper.GetString("token")
	if token != "" {
		err := globalClearpass.Validate(server, client, token)
		if err == nil {
			return token, nil
		}
		fmt.Println("login.run Error: ", err)
	}
	passwd, err := lib.Readline(fmt.Sprintf("Secret for '%s': ", client), true)
	if err != nil {
		return "", err
	}
	return globalClearpass.Login(server, client, passwd)
}

// Save login parameters
func (p loginCmdP) save(token string) error {
	if token != "" {
		viper.Set("token", token)
	}
	return viper.WriteConfig()
}
