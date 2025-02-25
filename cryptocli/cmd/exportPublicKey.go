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
    "os"
    "fmt"
    // external
    "github.com/spf13/cobra"
)


var exportPublicKeyCmd = &cobra.Command{
    Use:   "export-public-key",
    Short: "Export public key of asymmetric key",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        key_guid, _ := flags.GetString("key_guid")
        download, _ := flags.GetBool("download")

        if download == true {
            endpoint := GetEndPoint2("", "1.0", "key/"+key_guid+"/export/public?download=yes")
            fname, err := DoGetDownload(endpoint, GetCACertFile(),
                        AuthTokenKV())
            if err != nil {
                fmt.Printf("\nHTTP request failed: %s\n", err)
                os.Exit(4)
            } else {
                fmt.Println("\nSuccessfully downloaded public key " +
                "as - " + fname + "\n")
                os.Exit(0)
            }
        }
        endpoint := GetEndPoint("", "1.0", "key/"+key_guid+"/export/public")
        ret, err := DoGet(endpoint,
            GetCACertFile(),
            AuthTokenKV(),
            nil,
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
    rootCmd.AddCommand(exportPublicKeyCmd)
    exportPublicKeyCmd.Flags().StringP("key_guid", "k", "",
	"Key GUID")
    exportPublicKeyCmd.Flags().BoolP("download", "d", false,
    "True if you want to download the public key")

	exportPublicKeyCmd.MarkFlagRequired("key_guid")
}
