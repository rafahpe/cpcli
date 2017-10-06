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
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rafahpe/cpcli/model"
	"github.com/spf13/cobra"
	// Using this fork instead of the original because of write file support
	// See: https://github.com/spf13/viper/pull/287
	"github.com/theherk/viper"
)

// Error type for predefined errors
type Error string

func (e Error) Error() string {
	return string(e)
}

var cfgFile string

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

type options struct {
	PageSize    int
	Paginate    bool
	Mac         string
	SkipHeaders bool
	PrettyPrint bool
	Filter      []string
	Args        []string
}

var globalOptions options
var globalClearpass model.Clearpass

func init() {
	cobra.OnInitialize(initConfig)

	globalOptions = options{}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cpcli.yaml)")

	RootCmd.PersistentFlags().StringP("server", "s", "", "CPPM Server name or IP address")
	RootCmd.PersistentFlags().StringP("client", "c", "", "Client ID for accesing the CPPM API")
	RootCmd.PersistentFlags().StringP("token", "t", "", "OAUTH token")
	RootCmd.PersistentFlags().BoolP("unsafe", "u", false, "Skip server certificate verification")
	RootCmd.PersistentFlags().IntP("pagesize", "P", DefaultPageSize, "Pagesize of the requests")

	// Flags that are shared by several commands.
	RootCmd.PersistentFlags().BoolVarP(&(globalOptions.SkipHeaders), "skip-headers", "H", false, "Skip headers when dumping CSV")
	RootCmd.PersistentFlags().StringArrayVarP(&(globalOptions.Filter), "filter", "f", nil, "Filter arguments (e.g. mac=00:86:df:11:22:33)")
	RootCmd.PersistentFlags().BoolVarP(&(globalOptions.PrettyPrint), "prettyprint", "p", false, "Pretty print json output")

	viper.BindPFlag("server", RootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("client", RootCmd.PersistentFlags().Lookup("client"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("unsafe", RootCmd.PersistentFlags().Lookup("unsafe"))
	viper.BindPFlag("pagesize", RootCmd.PersistentFlags().Lookup("pagesize"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("initConfig Error: ", err)
		os.Exit(1)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cpcli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cpcli")
	}

	viper.SetEnvPrefix("cppm")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		cfgFile := filepath.Join(home, ".cpcli.yaml")
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			primeConfigFile(cfgFile)
		}
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("initConfig Error: ", err)
		}
	}

	// Init the connection to clearpass
	server := viper.GetString("server")
	token := viper.GetString("token")
	unsafe := viper.GetBool("unsafe")
	globalClearpass = model.New(server, token, unsafe)
}

func primeConfigFile(cfgFile string) {
	// Make sure the file exists, otherwise Viper complains when saving
	fd, err := os.OpenFile(cfgFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("primeConfigFile Error: ", err)
		os.Exit(1)
	}
	defer fd.Close()
}

// Simplifies getting common options
func updateOptions(cmd *cobra.Command, args []string) options {
	pageSize := viper.GetInt("pagesize")
	if pageSize <= 0 {
		globalOptions.PageSize = DefaultPageSize
		globalOptions.Paginate = false
	} else {
		globalOptions.PageSize = pageSize
		globalOptions.Paginate = true
	}
	globalOptions.Args = args
	return globalOptions
}
