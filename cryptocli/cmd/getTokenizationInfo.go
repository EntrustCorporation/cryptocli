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

var getTokenizationInfoCmd = &cobra.Command{
	Use:   "get-tokenization-info",
	Short: "Get Tokenization Info",
	Run: func(cmd *cobra.Command, args []string) {
		params := map[string]interface{}{}

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "GetTokenizationInfo")
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
			fmt.Println("\nTokenization not found\n")
			os.Exit(5)
		}

		fmt.Println("\n" + retStr + "\n")

		retMap := JsonStrToMap(retStr)
		if _, present := retMap["error"]; present {
			os.Exit(3)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(getTokenizationInfoCmd)
}
