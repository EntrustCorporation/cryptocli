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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var updateKeyStateCmd = &cobra.Command{
	Use:   "update-key-state",
	Short: "Update Key State",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		key_guid, _ := flags.GetString("key_guid")

		state, _ := flags.GetString("state")
		params["state"] = state

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "key/"+key_guid+"/state")
		ret, err := DoPost(endpoint,
			GetCACertFile(),
			AuthTokenKV(),
			jsonParams,
			"application/json")
		if err != nil {
			fmt.Printf("\nHTTP request failed: %s\n", err)
			os.Exit(4)
		}
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
		}
		fmt.Println("\n" + retStr + "\n")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(updateKeyStateCmd)
	updateKeyStateCmd.Flags().StringP("key_guid", "k", "", "Key GUID")
	updateKeyStateCmd.Flags().StringP("state", "s", "", "State. Supported states are enable and disable.")
	
	updateKeyStateCmd.MarkFlagRequired("key_guid")
	updateKeyStateCmd.MarkFlagRequired("state")
}