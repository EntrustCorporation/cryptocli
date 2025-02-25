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

var createMaskPolicyCmd = &cobra.Command{
	Use:   "create-mask-policy",
	Short: "Create Mask Policy",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		name, _ := flags.GetString("name")
		params["name"] = name

		if flags.Changed("description") {
			description, _ := flags.GetString("description")
			params["description"] = description
		}

		new, _ := flags.GetBool("new")
		params["isNew"] = new

		preservedPrefixLength, _ := flags.GetInt("preservedPrefixLength")
		params["preservedPrefixLength"] = preservedPrefixLength

		preservedSuffixLength, _ := flags.GetInt("preservedSuffixLength")
		params["preservedSuffixLength"] = preservedSuffixLength

		charset, _ := flags.GetString("charset")
		params["charset"] = charset

		maskChar, _ := flags.GetString("maskChar")
		params["maskChar"] = maskChar

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "CreateMaskPolicy")
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

			retMap := JsonStrToMap(retStr)
			if _, present := retMap["error"]; present {
				os.Exit(3)
			} else {
				fmt.Println("Masking policy successfully created:", name,
					"\n")
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createMaskPolicyCmd)
	createMaskPolicyCmd.Flags().StringP("name", "n", "",
		"Name of the policy to be created")
	createMaskPolicyCmd.Flags().StringP("description", "d", "",
		"Policy description")
	createMaskPolicyCmd.Flags().IntP("preservedPrefixLength", "P", 0,
		"Length of prefix to be preserved")
	createMaskPolicyCmd.Flags().IntP("preservedSuffixLength", "S", 0,
		"Length of suffix to be preserved")
	createMaskPolicyCmd.Flags().StringP("charset", "c", "",
		"Character set to be used with this policy")
	createMaskPolicyCmd.Flags().BoolP("new", "N", true,
		"True if creating a new policy, False if updating existing policy")
	createMaskPolicyCmd.Flags().StringP("maskChar", "m", "#",
		"Mask Character")

	createMaskPolicyCmd.MarkFlagRequired("name")
	createMaskPolicyCmd.MarkFlagRequired("charset")
	createMaskPolicyCmd.MarkFlagRequired("maskChar")
}