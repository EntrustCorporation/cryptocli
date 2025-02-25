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
    "bytes"
    "encoding/json"
    // external
    "github.com/spf13/cobra"
)

// updateLocalUserCmd represents the update-local-user command
var updateLocalUserCmd = &cobra.Command{
    Use:   "update-local-user",
    Short: "Update a given Local User",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        // create request payload
        user, _ := flags.GetString("user")
        params["username"] = user

        revision, _ := flags.GetInt("revision")
        params["revision"] = revision

        if flags.Changed("name") {
            name, _ := flags.GetString("name")
            charCheck(len(name))
            params["name"] = name
        }

        if flags.Changed("account-status") {
            accountStatus, _ := flags.GetString("account-status")
            if accountStatus == "enable" {
                params["account_state"] = true
            } else if accountStatus == "disable" {
                params["account_state"] = false
            } else {
                fmt.Printf("\n Valid values: enable or disable")
                os.Exit(1)
            }
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "UpdateLocalUser")
        ret, err := DoPost(endpoint,
                               GetCACertFile(),
                               AuthTokenKV(),
                               jsonParams,
                               "application/json")
        if err != nil {
            fmt.Printf("\nHTTP request failed: %s\n", err)
            os.Exit(4)
        } else {
            // type assertion
            retBytes := ret["data"].(*bytes.Buffer)
            retStatus := ret["status"].(int)
            retStr := retBytes.String()

            if (retStr == "" && retStatus == 404) {
                fmt.Println("\nUser not found\n")
                os.Exit(5)
            }

            fmt.Println("\n" + retStr + "\n")

            // make a decision on what to exit with
            retMap := JsonStrToMap(retStr)
            if _, present := retMap["error"]; present {
                os.Exit(3)
            } else {
                os.Exit(0)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(updateLocalUserCmd)
    updateLocalUserCmd.Flags().StringP("user", "u", "",
                                    "username of the user to update")
    updateLocalUserCmd.Flags().IntP("revision", "R", 0,
                                 "Revision number of the user")
    updateLocalUserCmd.Flags().StringP("name", "n", "",
                                    "Full Name of the User")
    updateLocalUserCmd.Flags().StringP("account-status", "s", "",
                                    "Account status of the User")

    // mark mandatory fields as required
    updateLocalUserCmd.MarkFlagRequired("user")
    updateLocalUserCmd.MarkFlagRequired("revision")
}
