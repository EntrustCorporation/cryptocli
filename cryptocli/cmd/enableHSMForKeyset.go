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

var enableHSMForKeysetCmd = &cobra.Command{
	Use:   "enable-hsm-for-keyset",
	Short: "Enable HSM for Keyset",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		keyset_guid, _ := flags.GetString("keyset_guid")
		params["keyset_guid"] = keyset_guid

		part_label, _ := flags.GetString("part_label")
		params["part_label"] = part_label

		part_password, _ := flags.GetString("part_password")
		params["part_password"] = part_password

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "EnableKeysetHSM")
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

		if (retStr == "" && retStatus == 404) {
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
	rootCmd.AddCommand(enableHSMForKeysetCmd)
	enableHSMForKeysetCmd.Flags().StringP("keyset_guid", "k", "",
		"Keyset for which HSM is to be enabled")
	enableHSMForKeysetCmd.Flags().StringP("part_label", "l", "",
		"Partition Label")
	enableHSMForKeysetCmd.Flags().StringP("part_password", "p", "",
		"Partition Password")

	enableHSMForKeysetCmd.MarkFlagRequired("keyset_guid")
}