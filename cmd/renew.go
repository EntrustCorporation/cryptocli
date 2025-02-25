/*
 Copyright 2020-2025 Entrust Corporation

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
    // standard
    "os"
    "fmt"
    "encoding/json"
    // external
    "github.com/spf13/cobra"
)

type renewAccessToken struct {
	Token      string `json:"access_token"`
	Expiration string `json:"expires_at"`
}

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
    Use:   "renew",
    Short: "Renew Access Token",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

	    var respData renewAccessToken

        // now POST
        endpoint := GetEndPoint("", "1.0", "Renew")
	    _, err = DoPost2(endpoint, GetCACertFile(),
        AuthTokenKV(),
		    jsonParams, ContentTypeJSON, &respData, nil)
	    if err != nil {
		    fmt.Printf("\nSession Renew failed:\n\n%v\n", err)
		    os.Exit(1)
        }

        // save access token to a file
        tokenFile, _ := flags.GetString(loginOptionTokenFile)
        tokenFile, err = SaveAccessToken(tokenFile,
                                               respData.Token,
                                               GetServer(),
                                               GetCACertFile())
        if err != nil {
            fmt.Printf("\nError saving access token to %s - %v\n", tokenFile, err)
            os.Exit(4)
        }

        fmt.Printf("\nSession is renewed.\nThe login session expires at %s.\n",
                formatLoginExpiration(respData.Expiration))
        fmt.Printf("New Access Token is saved in %s.\n", tokenFile)
        fmt.Printf("\n")
    },
}

func init() {
    rootCmd.AddCommand(renewCmd)
}
