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

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key-csr",
	Short: "Generate Key and CSR",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		cipher, _ := flags.GetString("cipher")
		params["cipher"] = cipher

		if flags.Changed("public_key") {
			public_key, _ := flags.GetString("public_key")
			params["public_key"] = public_key
		}

		if flags.Changed("keyset_guid") {
			keyset_guid, _ := flags.GetString("keyset_guid")
			params["keyset_guid"] = keyset_guid
		}
		
		if flags.Changed("subject_dn") {
			subject_dn, _ := flags.GetString("subject_dn")
			params["subject_dn"] = subject_dn
		}

		if flags.Changed("sans") {
			sans, _ := flags.GetString("sans")
			params["sans"] = sans
		}

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "generate_key_csr")
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
		fmt.Println("Key successfully geneated\n")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)
	generateKeyCmd.Flags().StringP("keyset_guid", "k", "",
		"Keyset to be used for generating the key")
	generateKeyCmd.Flags().StringP("public_key", "p", "",
		"Public Key used to wrap the generated key")
	generateKeyCmd.Flags().StringP("cipher", "c", "",
		"Cipher for this key")
	generateKeyCmd.Flags().StringP("subject_dn", "S", "",
		"Subject distinguished name required for CSR generation. If not provided then a CSR with default values (CN=kcv.entrust.com, C=US) is created")
	generateKeyCmd.Flags().StringP("sans", "s", "",
	"Subject Alternative Names. It will only be considered when subject_dn is provided")

	generateKeyCmd.MarkFlagRequired("cipher")}