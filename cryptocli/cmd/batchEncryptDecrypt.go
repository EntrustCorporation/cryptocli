/*
 Copyright 2025 Entrust Corporation

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

var batchEncryptDecryptCmd = &cobra.Command{
	Use:   "batch-encrypt-decrypt",
	Short: "Batch Encrypt/Decrypt. Please provide keyGuid, data, mode, iv, aad and operation alternatively.",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()

		data, _ := flags.GetStringArray("data")
		keyGuid, _ := flags.GetStringArray("keyGuid")
		mode, _ := flags.GetStringArray("mode")
		iv, _ := flags.GetStringArray("iv")
		aad, _ := flags.GetStringArray("aad")
		operation, _ := flags.GetStringArray("operation")

		if len(mode) != len(data) || len(mode) != len(keyGuid) || len(mode) != len(iv) || len(mode) != len(aad) || len(mode) != len(operation) {
			fmt.Println("Missing parameters. Please check and try again")
			os.Exit(1)
		}

		request := []interface{}{}

		for i := 0; i < len(mode); i++ {
			params := map[string]interface{}{}
			params["keyGuid"] = keyGuid[i]
			params["data"] = data[i]
			params["operation"] = operation[i]
			params["mode"] = mode[i]
			
			if iv[i] != "0" {
				params["iv"] = iv[i]
			}
			if aad[i] != "0" {
				params["aad"] = aad[i]
			}
			request = append(request, params)
		}

		jsonParams, err := json.Marshal(request)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "batch/encrypt-decrypt")
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
			
			retMap := JsonArrayStrToMap(retStr)
			jsonData, _ := JSONMarshalIndent(retMap)
			fmt.Println("\n" + string(jsonData) + "\n")
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(batchEncryptDecryptCmd)
	batchEncryptDecryptCmd.Flags().StringArrayP("data", "d", []string{}, "Data to be encrypted")
	batchEncryptDecryptCmd.Flags().StringArrayP("keyGuid", "k", []string{}, "Key GUID to be used for encryption")
	batchEncryptDecryptCmd.Flags().StringArrayP("mode", "m", []string{}, "Mode of encryption")
	batchEncryptDecryptCmd.Flags().StringArrayP("iv", "i", []string{}, "Enter initialization vector if required else provide 0")
	batchEncryptDecryptCmd.Flags().StringArrayP("aad", "a", []string{}, "Enter Additional authentication data if required else provide 0")
	batchEncryptDecryptCmd.Flags().StringArrayP("operation", "o", []string{}, "Operation type: Encrypt/Decrypt")

	batchEncryptDecryptCmd.MarkFlagRequired("data")
	batchEncryptDecryptCmd.MarkFlagRequired("keyGuid")
	batchEncryptDecryptCmd.MarkFlagRequired("mode")
	batchEncryptDecryptCmd.MarkFlagRequired("iv")
	batchEncryptDecryptCmd.MarkFlagRequired("aad")
	batchEncryptDecryptCmd.MarkFlagRequired("operation")
}
