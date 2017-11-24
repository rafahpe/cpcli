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

	"github.com/peterh/liner"
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
    variable, CPPM_SERVER environment variable, or with the -h flag.
  - The OAUTH token can be provided in the 'token' configuration variable,
	the CPPM_TOKEN environment variable, or the -t flag.
  - If you have an OAUTH refresh token, it can be provided in the 'refresh'
    configuration variable, the CPPM_REFRESH environment variable, or the -r flag.
  - If OAUTH token is missing, invalid or expired, then client_id can be
	provided in the 'client' config variable, CPPM_CLIENT environment
	variable, or -c flag.
  - If you are using username/password based auth, besides the client ID,
	you will need to provide your username with the 'user' config variable,
	CPPM_USER environment variable, or -u flag`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		p := loginCmdP{cmd}
		if token, refresh, err := p.run(); err != nil {
			fmt.Println("Login error: ", err)
		} else {
			if err := p.save(token, refresh); err != nil {
				fmt.Println("login Error saving config data: ", err)
			}
			fmt.Println("login OK. Token: ", token)
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

// Run login
func (p loginCmdP) run() (string, string, error) {
	server := viper.GetString("server")
	if server == "" {
		return "", "", ErrMissingserver
	}
	client := viper.GetString("client")
	if client == "" {
		return "", "", ErrMissingCreds
	}
	token := viper.GetString("token")
	refresh := viper.GetString("refresh")
	ctx := context.Background()
	if token != "" {
		token, refresh, err := globalClearpass.Validate(ctx, server, client, "", token, refresh)
		if err == nil {
			return token, refresh, nil
		}
		fmt.Println("login.run Error: ", err)
	}
	secret, err := readline(fmt.Sprintf("Secret for '%s' (leave blank if public client): ", client), true)
	if err != nil {
		return "", "", err
	}
	user, password := viper.GetString("user"), ""
	if user != "" {
		password, err = readline(fmt.Sprintf("Password for '%s' (leave blank if auth type is 'client_credentials'): ", user), true)
		if err != nil {
			return "", "", err
		}
	}
	return globalClearpass.Login(ctx, server, client, secret, user, password)
}

// Save login parameters
func (p loginCmdP) save(token, refresh string) error {
	if token != "" {
		viper.Set("token", token)
	}
	if refresh != "" {
		viper.Set("refresh", refresh)
	}
	return viper.WriteConfig()
}

// Readline reads a single line of input
func readline(prompt string, password bool) (string, error) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	if !password {
		return line.Prompt(prompt)
	}
	return line.PasswordPrompt(prompt)
}
