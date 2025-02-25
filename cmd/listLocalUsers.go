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
	"bytes"

	// external
	"github.com/spf13/cobra"
)


// listLocalUsersCmd represents the list-local-users command
var listLocalUsersCmd = &cobra.Command{
	Use:	"list-local-users",
	Short:	"List all Local Users",
	Run: func(cmd *cobra.Command, args []string) {
	    flags := cmd.Flags()
        params := map[string]interface{}{}

        // create request payload
        if flags.Changed("prefix") {
            prefix, _ := flags.GetString("prefix")
            params["prefix"] = prefix
        }
        
        if flags.Changed("filters") {
            filters, _ := flags.GetString("filters")
            params["filters"] = filters
        }
        
        if flags.Changed("max-items") {
            maxItems, _ := flags.GetInt("max-items")
            params["max_items"] = maxItems
        }
        
        if flags.Changed("fields") {
            fieldsArray, _ := flags.GetStringArray("fields")
            params["fields"] = fieldsArray
        }
        
        if flags.Changed("next-token") {
            nextToken, _ := flags.GetString("next-token")
            params["next_token"] = nextToken
        }

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

		// now POST
	    endpoint := GetEndPoint("", "1.0", "ListLocalUsers")
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

	    	// make a decision on what to exit with
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
    rootCmd.AddCommand(listLocalUsersCmd)
	listLocalUsersCmd.Flags().StringP("prefix", "p", "",
									"Prefix to list Users with")
	listLocalUsersCmd.Flags().StringP("filters", "l", "",
									"Conditional expression to list filtered " +
									"users")
	listLocalUsersCmd.Flags().IntP("max-items", "m", 0,
								"Maximum number of items to include " +
								"in the response")
	listLocalUsersCmd.Flags().StringArrayP("fields", "f", []string{},
									"Fields to include in the response")
	listLocalUsersCmd.Flags().StringP("next-token", "n", "",
								"Token from which subsequent Users would " +
								"be listed")
}
