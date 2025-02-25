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
	"fmt"
	"os"
	"encoding/json"
	"github.com/spf13/cobra"
)

var exportKeyCmd = &cobra.Command{
	Use:   "export-key",
	Short: "Export Key",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		key_guid, _ := flags.GetString("key_guid")
		sha256, _ := flags.GetBool("sha256")

		public_key, _ := flags.GetString("public_key")
		params["public_key"] = public_key

		jsonParams, err := json.Marshal(params)
	    if (err != nil) {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
	    }

		endpoint := GetEndPoint("", "1.0", "key/"+key_guid+"/export")
		if sha256 {
			endpoint = GetEndPoint2("", "1.0", "key/"+key_guid+"/export/?sha256=yes")
		}
		ret, err := DoPostFormData(endpoint,
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
	rootCmd.AddCommand(exportKeyCmd)
	exportKeyCmd.Flags().StringP("key_guid", "k", "", "Key GUID")
	exportKeyCmd.Flags().StringP("public_key", "p", "", "Public Key File")
	exportKeyCmd.Flags().BoolP("sha256", "s", false,
    "True if you want to use SHA256 hash for wrapping. Default hash is SHA1")

	exportKeyCmd.MarkFlagRequired("key_guid")
	exportKeyCmd.MarkFlagRequired("public_key")
}
