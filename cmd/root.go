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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cpcli",
	Short: "Command-line interface for Aruba Clearpass API",
	Long: `cpcli (cp for short) is a command line application that interacts with
    Aruba Clearpass through the REST API.

It performs:

  - Authentication against Clearpass with the "login" command.
  - GET, POST, PUT, PATCH, DELETE requests to the API.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("execute Error: ", err)
		os.Exit(1)
	}
}

// DefaultPageSize is the default page size for pagination
const DefaultPageSize = 24

func init() {
	cobra.OnInitialize(func() { Singleton.OnInit() })

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Flags not stored in the config file
	RootCmd.PersistentFlags().StringVar(&Singleton.ConfigFile, "config", "", "config file (default is $HOME/.cpcli)")
	RootCmd.PersistentFlags().BoolVarP(&(Singleton.Options.SkipHeaders), "skip-headers", "H", false, "Skip headers when dumping CSV")
	RootCmd.PersistentFlags().StringArrayVarP(&(Singleton.Query), "query", "q", nil, "Query params (e.g. -q sort=-id -q filter={mac:'00:86:df:11:22:33'}")
	RootCmd.PersistentFlags().BoolVarP(&(Singleton.Options.PrettyPrint), "prettyprint", "p", false, "Pretty print json output")
	RootCmd.PersistentFlags().BoolVarP(&(Singleton.Force), "force", "F", false, "When used with 'login', force new authentication")

	// Flags stored in config file / viper
	RootCmd.PersistentFlags().StringP("server", "s", "", "CPPM Server name or IP address")
	RootCmd.PersistentFlags().StringP("client", "c", "", "Client ID for accesing the CPPM API")
	RootCmd.PersistentFlags().StringP("user", "u", "", "User name for accesing the CPPM API")
	RootCmd.PersistentFlags().StringP("token", "t", "", "OAUTH token")
	RootCmd.PersistentFlags().StringP("refresh", "r", "", "OAUTH refresh token")
	RootCmd.PersistentFlags().BoolP("unsafe", "U", false, "Skip server certificate verification")
	RootCmd.PersistentFlags().IntP("pagesize", "P", DefaultPageSize, "Pagesize of the requests")

	viper.BindPFlag("server", RootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("client", RootCmd.PersistentFlags().Lookup("client"))
	viper.BindPFlag("user", RootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("refresh", RootCmd.PersistentFlags().Lookup("refresh"))
	viper.BindPFlag("unsafe", RootCmd.PersistentFlags().Lookup("unsafe"))
	viper.BindPFlag("pagesize", RootCmd.PersistentFlags().Lookup("pagesize"))
}
