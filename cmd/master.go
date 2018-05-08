package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rafahpe/cpcli/model"
	"github.com/rafahpe/cpcli/term"
	"github.com/spf13/viper"
)

// Master is the master application object
type Master struct {
	cppm model.Clearpass

	// Logger for error messages
	Log *log.Logger

	// Options to mamage with Cobra
	ConfigFile string
	Options    term.Options
	Force      bool
	Query      []string
}

// Error type for predefined errors
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	// ErrMissingPath returned when no path is provided for a REST op
	ErrMissingPath = Error("Missing path parameter for REST operation")
	// ErrMissingserver returned when there is no server to log in
	ErrMissingserver = Error("Missing CPPM server name or IP address")
	// ErrMissingCreds returned when there are no credentials to log in
	ErrMissingCreds = Error("Missing credentials to log into CPPM")
	// ErrInvalidCreds returned when not allowed to log in
	ErrInvalidCreds = Error("Credentials are invalid or expired")
)

// Singleton is the config holder for all commands
var Singleton Master

// OnInit reads in config file and ENV variables if set.
func (master *Master) OnInit() {

	// Find home directory.
	master.Log = log.New(os.Stderr, "", 0)
	home, err := homedir.Dir()
	if err != nil {
		master.Log.Fatal("Could not find home directory: ", err)
	}

	// Get config file from command line, default ".cpcli"
	if master.ConfigFile != "" {
		viper.SetConfigFile(master.ConfigFile)
	} else {
		// Search config in home directory with name ".cpcli" (without extension).
		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.SetConfigName(".cpcli")
	}
	viper.SetEnvPrefix("cppm")
	viper.AutomaticEnv() // read in environment variables that match

	// Read or create the config file
	if err := viper.ReadInConfig(); err != nil {
		master.ConfigFile = path.Join(home, ".cpcli.yaml")
		primeConfigFile(master.ConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			master.Log.Fatal("initConfig Error: ", err)
		}
	}

	pageSize := viper.GetInt("pagesize")
	if pageSize <= 0 {
		master.Options.PageSize = DefaultPageSize
		master.Options.Paginate = false
	} else {
		master.Options.PageSize = pageSize
		master.Options.Paginate = true
	}

	// Init the connection to clearpass
	server := viper.GetString("server")
	token := viper.GetString("token")
	refresh := viper.GetString("refresh")
	unsafe := viper.GetBool("unsafe")
	master.cppm = model.New(server, token, refresh, unsafe)
}

// Make sure the file exists, otherwise Viper complains when saving
func primeConfigFile(cfgFile string) {
	fd, err := os.OpenFile(cfgFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("primeConfigFile Error: ", err)
		os.Exit(1)
	}
	defer fd.Close()
}

// Save login parameters
func (master *Master) Save(token, refresh string) error {
	if token != "" {
		viper.Set("token", token)
	}
	if refresh != "" {
		viper.Set("refresh", refresh)
	}
	return viper.WriteConfig()
}

// Login into the ClearPass. Return access and refresh token
func (master *Master) Login() (string, string, error) {
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
	if token != "" && !master.Force {
		token, refresh, err := master.cppm.Validate(ctx, server, client, "", token, refresh)
		if err == nil {
			return token, refresh, nil
		}
		fmt.Println("login.run Error: ", err)
	}
	secret, err := term.Readline(fmt.Sprintf("Secret for '%s' (leave blank if public client): ", client), true)
	if err != nil {
		return "", "", err
	}
	user, password := viper.GetString("user"), ""
	if user != "" {
		password, err = term.Readline(fmt.Sprintf("Password for '%s' (leave blank if auth type is 'client_credentials'): ", user), true)
		if err != nil {
			return "", "", err
		}
	}
	return master.cppm.Login(ctx, server, client, secret, user, password)
}

// Run runs a command against the Clearpass
func (master *Master) Run(method model.Method, args []string) error {
	if len(args) < 1 {
		return ErrMissingPath
	}
	// Read the filter
	query, err := master.readQuery()
	if err != nil {
		return err
	}
	// Check if we are in a pipe
	reader, err := term.Stdin()
	if err != nil {
		return err
	}
	path, format := args[0], args[1:]
	// If stdin is a tty, run just once
	if reader == nil {
		return master.do(method, path, query, nil, format)
	}
	// Otherwise, iterate over the pipe
	for reader.Next() {
		item := reader.Get()
		if err := master.do(method, path, query, item, format); err != nil {
			return err
		}
	}
	return reader.Error()
}

// Runs the request and outputs the result
func (master *Master) do(method model.Method, path string, query model.Params, request interface{}, format []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := master.cppm.Request(method, path, query, request)
	return term.Output(ctx, master.Options, feed, format)
}
