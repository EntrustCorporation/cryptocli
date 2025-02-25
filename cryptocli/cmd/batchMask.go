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

var batchMaskCmd = &cobra.Command{
	Use:   "batch-mask",
	Short: "Batch Mask. Please provide policyName and tokenData alternatively.",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()

		policyName, _ := flags.GetStringArray("policyName")
		tokenData, _ := flags.GetStringArray("tokenData")

		if len(policyName) != len(tokenData) {
			fmt.Println("Missing parameters. Please check and try again")
			os.Exit(1)
		}

		request := []interface{}{}

		for i := 0; i < len(policyName); i++ {
			params := map[string]interface{}{}
			params["policyName"] = policyName[i]
			params["tokenData"] = tokenData[i]
			request = append(request, params)
		}

		jsonParams, err := json.Marshal(request)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "batch/mask")
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

			fmt.Println("\n" + retStr + "\n")
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(batchMaskCmd)
	batchMaskCmd.Flags().StringArrayP("policyName", "n", []string{},
		"Name of the policy to be used to masking")
	batchMaskCmd.Flags().StringArrayP("tokenData", "d", []string{},
		"Data to be masked")

	batchMaskCmd.MarkFlagRequired("policyName")
	batchMaskCmd.MarkFlagRequired("tokenData")
}
