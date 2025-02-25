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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var updatePersonalAccessTokenCmd = &cobra.Command{
	Use:   "update-personal-access-token",
	Short: "Update a Personal Access Token",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		name, _ := flags.GetString("name")
		params["name"] = name

		if flags.Changed("description") {
			description, _ := flags.GetString("description")
			params["description"] = description
		}

		if flags.Changed("expiry") {
			expiry, _ := flags.GetString("expiry")
			expiry = expiry + "T10:00:00Z"
			thetime, e := time.Parse(time.RFC3339, expiry)
			if e != nil {
				fmt.Println("Can't parse time format\n")
			}
			params["expiry"] = thetime.Unix()
		}

		if flags.Changed("revoked") {
			revoked, _ := flags.GetBool("revoked")
			params["revoked"] = revoked
		}

		if len(params) < 2 {
			fmt.Println("nothing to update")
			os.Exit(1)
		}

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "UpdatePersonalAccessToken")
		ret, err := DoPost(endpoint,
			GetCACertFile(),
			AuthTokenKV(),
			jsonParams,
			"application/json")
		if err != nil {
			fmt.Printf("\nHTTP request failed: %s\n", err)
			os.Exit(4)
		} else {
			retBytes := ret["data"].(*bytes.Buffer)
			retStatus := ret["status"].(int)
			retStr := retBytes.String()

			if retStr == "" && retStatus == 404 {
				fmt.Println("\nAction denied\n")
				os.Exit(5)
			}

			retMap := JsonStrToMap(retStr)
			if _, present := retMap["error"]; present {
				fmt.Println("\n" + retStr + "\n")
				os.Exit(3)
			} else {
				fmt.Println("\n" + retStr + "\n")
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(updatePersonalAccessTokenCmd)
	updatePersonalAccessTokenCmd.Flags().StringP("description", "d", "",
		"Description of personal access token")
	updatePersonalAccessTokenCmd.Flags().StringP("name", "n", "",
		"Name of personal access token")
	updatePersonalAccessTokenCmd.Flags().StringP("expiry", "e", "",
		"Date on which personal access token expires in yyyy-mm-dd format")
	updatePersonalAccessTokenCmd.Flags().BoolP("revoked", "r", false,
		"True if you want to revoke this token")

	updatePersonalAccessTokenCmd.MarkFlagRequired("name")
}
