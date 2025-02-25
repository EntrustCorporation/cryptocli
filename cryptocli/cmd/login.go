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
	"bufio"
	"cli/getpasswd"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strings"
)

const (
	loginOptionLoginURL  = "login-URL"
	loginOptionUserName  = "username"
	loginOptionPassword  = "password"
	loginOptionCACert    = "cacert"
	loginOptionTokenFile = "token-file"
)

type accessToken struct {
	Token      string `json:"access_token"`
	Expiration string `json:"expires_at"`
	User       string `json:"user"`
}

func getCredentials(prefix, user, password string) (string, string) {

	if user == "" {
		fmt.Printf("%sUser Name: ", prefix)
		reader := bufio.NewReader(os.Stdin)
		user, _ = reader.ReadString('\n')
		// the returned string includes \r\n on Windows, \n on *nix
		if strings.HasSuffix(user, "\n") {
			user = user[:len(user)-1]
		}
		if strings.HasSuffix(user, "\r") {
			user = user[:len(user)-1]
		}
	}

	if password == "" {
		fmt.Printf("%sPassword: ", prefix)
		password = getpasswd.ReadPassword()
		fmt.Printf("\n")
	}

	return user, password
}

func loginAPI(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	params := map[string]interface{}{}

	username, _ := flags.GetString(loginOptionUserName)
	password, _ := flags.GetString(loginOptionPassword)
	if password == "" || username == "" {
		fmt.Printf("\n")
		username, password = getCredentials("", username, password)
	}

	params["username"] = username
	params["password"] = password

	loginURL, _ := flags.GetString(loginOptionLoginURL)
	if !strings.HasPrefix(loginURL, "https://") {
		fmt.Printf("Provide complete login URL\n")
		os.Exit(1)
	}
	if !strings.HasSuffix(loginURL, "/") {
		loginURL = loginURL + "/"
	}
	uri, err := url.Parse(loginURL)
	if err != nil {
		fmt.Printf("Invalid URL %s - %v\n", loginURL, err)
		os.Exit(2)
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(2)
	}

	cacert, _ := flags.GetString(loginOptionCACert)
	var respData accessToken
	_, err = DoPost2(loginURL, cacert, map[string]string{},
		jsonParams, ContentTypeJSON, &respData, nil)
	if err != nil {
		fmt.Printf("\nLogin failed:\n%v\n", err)
		os.Exit(3)
	}

	tokenFile, _ := flags.GetString(loginOptionTokenFile)
	tokenFile, err = SaveAccessToken(tokenFile, respData.Token, uri.Hostname(), cacert)
	if err != nil {
		fmt.Printf("\nError saving access token to %s - %v\n", tokenFile, err)
		os.Exit(4)
	}

	fmt.Printf("\nLogin is successful.\nThe login session expires at %s.\n",
		formatLoginExpiration(respData.Expiration))
	fmt.Printf("Access Token is saved in %s.\n", tokenFile)
	fmt.Printf("\n")
	os.Exit(0)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Tokenization Vault",
	Run:   loginAPI,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP(loginOptionCACert, "C", "",
		"CA Certificate to verify Tokenization Vault server with")
	loginCmd.Flags().StringP(loginOptionUserName, "u", "",
		"Login username. You will be prompted to enter if not provided.")
	loginCmd.Flags().StringP(loginOptionPassword, "p", "",
		"Login password. You will be prompted to enter if not provided.")
	loginCmd.Flags().StringP(loginOptionLoginURL, "l", "",
		"Login URL")

	loginCmd.MarkFlagRequired(loginOptionCACert)
	loginCmd.MarkFlagRequired(loginOptionLoginURL)
}
