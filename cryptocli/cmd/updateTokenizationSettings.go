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

var updateTokenizationSettingsCmd = &cobra.Command{
	Use:   "update-tokenization-settings",
	Short: "Update Tokenization settings",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		degradedModeAvailProvided := flags.Changed("degraded-mode-availability")
		oidcProvided := flags.Changed("oidc-enabled")
		if !degradedModeAvailProvided && !oidcProvided {
			fmt.Println("Specify any one of Tokenization settings to update")
			os.Exit(1)
		}

		if flags.Changed("degraded-mode-availability") {
			degradedModeAvailability, _ := flags.GetString("degraded-mode-availability")
			if degradedModeAvailability != "enable" && degradedModeAvailability != "disable" {
				fmt.Printf("\nInvalid -d, --degraded-mode-availability option %s. "+
					"Supported: enable (or) disable\n", degradedModeAvailability)
				os.Exit(1)
			}
			params["degraded_mode_availability"] = (degradedModeAvailability == "enable")
		}		

		if flags.Changed("oidc-enabled") {
			oidcEnabled, _ := flags.GetString("oidc-enabled")
			if oidcEnabled != "enable" && oidcEnabled != "disable" {
				fmt.Printf("\nInvalid -o, --oidc-enabled option %s. "+
								"Supported: enable (or) disable\n", oidcEnabled)
				os.Exit(1)
			}
			params["oidc_enabled"] = (oidcEnabled == "enable")
		}

		revision, _ := flags.GetInt("revision")
		params["revision"] = revision

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		endpoint := GetEndPoint("", "1.0", "UpdateTokenizationSettings")
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
			fmt.Println("\nTokenization Settings not found\n")
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
	rootCmd.AddCommand(updateTokenizationSettingsCmd)
	updateTokenizationSettingsCmd.Flags().StringP("degraded-mode-availability", "d", "",
		"Degraded mode availability. ")
	updateTokenizationSettingsCmd.Flags().StringP("oidc-enabled", "o", "",
		"OIDC enabled flag. ")
	updateTokenizationSettingsCmd.Flags().IntP("revision", "R", 0,
		"Revision number of tokenization settings")

	updateTokenizationSettingsCmd.MarkFlagRequired("revision")
}
