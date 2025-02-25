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

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		keyGuid, _ := flags.GetString("keyGuid")
		params["keyGuid"] = keyGuid

		data, _ := flags.GetString("data")
		params["data"] = data

		mode, _ := flags.GetString("mode")
		params["mode"] = mode

		iv, _ := flags.GetString("iv")
		params["iv"] = iv

		aad, _ := flags.GetString("aad")
		params["aad"] = aad

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "encrypt")
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
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringP("keyGuid", "k", "", "Key GUID to be used for encryption")
	encryptCmd.Flags().StringP("data", "d", "", "Data to be encrypted")
	encryptCmd.Flags().StringP("mode", "m", "", "Mode of encryption")
	encryptCmd.Flags().StringP("iv", "i", "", "Initialization vector")
	encryptCmd.Flags().StringP("aad", "a", "", "Additional authentication data")

	encryptCmd.MarkFlagRequired("keyGuid")
	encryptCmd.MarkFlagRequired("data")
	encryptCmd.MarkFlagRequired("mode")
}
