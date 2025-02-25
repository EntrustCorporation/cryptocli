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

// error checking for exceeding 256 chars
func charCheck(length int) {
    if (length > 256) {
        fmt.Println("Exceeded 256 characters. Please provide an " +
		    "input length <= 256 characters")
	os.Exit(1)
    }
}

// createLocalUserCmd represents the create-local-user command
var createLocalUserCmd = &cobra.Command{
	Use:	"create-local-user",
	Short:	"Create a Local User",
	Run: func(cmd *cobra.Command, args []string) {
	    flags := cmd.Flags()
	    params := map[string]interface{}{}

	    // create request payload
	    params["account_state"] = true

	    email, _ := flags.GetString("email")
	    charCheck(len(email))
	    params["email"] = email

	    name, _ := flags.GetString("name")
	    charCheck(len(name))
	    params["name"] = name

	    // JSONify
	    jsonParams, err := json.Marshal(params)
	    if (err != nil) {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
	    }

	    // now POST
	    endpoint := GetEndPoint("", "1.0", "CreateLocalUser")
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
		       fmt.Println("\nAction denied\n")
		   os.Exit(5)
	       }

	       fmt.Println("\n" + retStr + "\n")

	       // make a decision on what to exit with
	       retMap := JsonStrToMap(retStr)
	       if _, present := retMap["error"]; present {
		   os.Exit(3)
	       } else {
		   fmt.Println("Local user successfully created for", name,
		               "with username", email, "\n")
		   os.Exit(0)
	       }
           }
    },
}

func init() {
    rootCmd.AddCommand(createLocalUserCmd)
    createLocalUserCmd.Flags().StringP("email", "e", "",
                                  "Email of the User")
    createLocalUserCmd.Flags().StringP("name", "n", "",
                                  "Full name of the user")

    // mark mandatory fields as required
    createLocalUserCmd.MarkFlagRequired("email")
    createLocalUserCmd.MarkFlagRequired("name")
}
