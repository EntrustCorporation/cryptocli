/*
 Copyright 2023-2025 Entrust Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var gAccessTokenFile string

const defaultCfgFileName = "cryptocli.cfg"

var rootCmd = &cobra.Command{
	Use:   "cryptocli",
	Short: "Entrust Tokenization Vault CLI",
	Long:  `Perform Tokenization Vault operations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initAccessToken)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is "+defaultCfgFileName+" on your home/profile directory)")
	rootCmd.PersistentFlags().StringVar(&gAccessTokenFile, loginOptionTokenFile, "",
		"Name of the File (with full path) for saving and reusing Access Token "+
			"and Server details. Login command creates this file while other commands "+
			"use this file. If a token file is not specified, default file tokenization_token.txt "+
			"is created in cryptocli.data/ under your home or profile directory.")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(defaultCfgFileName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
	}
}

func initAccessToken() {

	if len(os.Args) < 2 {
		return
	}

	excludedCommands := [...]string{"login", "version", "help"}
	for _, cmd := range excludedCommands {
		if os.Args[1] == cmd {
			return
		}
	}

	for i := 2; i < len(os.Args); i++ {
		var param = os.Args[i]
		if strings.HasPrefix(param, "-") && len(param) > 2 && param[1] != '-' {
			fmt.Printf("\nInvalid parameter %q. Parameter names must be prefixed with --\nE.g. -%s\n\n",
				param, param)
			os.Exit(1)
		}
	}

	tokenFile, err := LoadAccessToken(gAccessTokenFile)
	if err != nil {
		fmt.Printf("\nError getting Server information from %s. %v.\nIf you are not logged in yet, log into the vault by running login command.\n\n",
			tokenFile, err)
		os.Exit(1)
	}
}