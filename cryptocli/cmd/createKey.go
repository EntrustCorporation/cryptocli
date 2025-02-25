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

var createKeyCmd = &cobra.Command{
	Use:   "create-key",
	Short: "Create Key",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		name, _ := flags.GetString("name")
		params["name"] = name

		if flags.Changed("description") {
			description, _ := flags.GetString("description")
			params["description"] = description
		}

		if flags.Changed("keyset_guid") {
			keyset_guid, _ := flags.GetString("keyset_guid")
			params["keyset_guid"] = keyset_guid
		}

		cipher, _ := flags.GetString("cipher")
		params["cipher"] = cipher

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "key")
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

		fmt.Println("\n" + retStr + "\n")

		retMap := JsonStrToMap(retStr)
		if _, present := retMap["error"]; present {
			os.Exit(3)
		}
		fmt.Println("Key successfully created:", name,
			"\n")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(createKeyCmd)
	createKeyCmd.Flags().StringP("keyset_guid", "k", "",
		"Keyset to be used for creating this key")
	createKeyCmd.Flags().StringP("name", "n", "",
		"Name of the key to be created")
	createKeyCmd.Flags().StringP("description", "d", "",
		"Key description")
	createKeyCmd.Flags().StringP("cipher", "c", "",
		"Cipher for this key")

	createKeyCmd.MarkFlagRequired("name")
	createKeyCmd.MarkFlagRequired("cipher")
}