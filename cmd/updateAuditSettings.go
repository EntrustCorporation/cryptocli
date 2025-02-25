/*
 Copyright 2020-2025 Entrust Corporation

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
	// standard
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	// external
	"github.com/spf13/cobra"
)

// updateAuditSettingsCmd represents the update-audit-settings command
var updateAuditSettingsCmd = &cobra.Command{
	Use:   "update-audit-settings",
	Short: "Update audit settings",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		// create request payload
		if flags.Changed("retention-days") {
			retentionDays, _ := flags.GetInt32("retention-days")
			if retentionDays < 0 {
				fmt.Printf("\nRetention days must be 0 (retain all) or greater\n\n")
				os.Exit(1)
			}
			params["auditlog_retention"] = retentionDays
		}

		if flags.Changed("max-logs-size") {
			maxLogsSize, _ := flags.GetInt64("max-logs-size")
			if maxLogsSize < 0 {
				fmt.Printf("\nMaximum log size must be 0 (no limit) or greater\n\n")
				os.Exit(1)
			}
			params["auditlog_totsize"] = maxLogsSize
		}

		if len(params) <= 0 {
			cmd.Usage()
			os.Exit(1)
		}

		// JSONify
		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("Error building JSON request: ", err)
			os.Exit(1)
		}

		// now POST
		endpoint := GetEndPoint("", "1.0", "UpdateAuditSetting")
		ret, err := DoPost(endpoint,
			GetCACertFile(),
			AuthTokenKV(),
			jsonParams,
			"application/json")
		if err != nil {
			fmt.Printf("\nHTTP request failed: %s\n", err)
			os.Exit(4)
		} else {
			// type assertion
			retBytes := ret["data"].(*bytes.Buffer)
			retStatus := ret["status"].(int)
			retStr := retBytes.String()

			if retStr == "" && retStatus == 404 {
				fmt.Println("\nAudit settings not found\n")
				os.Exit(5)
			}

			if retStatus == 204 {
				fmt.Println("\nUpdate successful\n")
				os.Exit(0)
			} else {
				fmt.Println("\n" + retStr + "\n")
				// make a decision on what to exit with
				retMap := JsonStrToMap(retStr)
				if _, present := retMap["error"]; present {
					os.Exit(3)
				} else {
					fmt.Println("\nUnknown error\n")
					os.Exit(0)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(updateAuditSettingsCmd)
	// auditlog_retention
	updateAuditSettingsCmd.Flags().Int32P("retention-days", "r", 0,
		"Number of days to retain the audit logs. Retention days can be set between " +
        "30 - 365 days. Set 0 to retain all.")
	updateAuditSettingsCmd.Flags().Int64P("max-logs-size", "m", 0,
		"Maximum size for the audit logs in bytes. Set 0 for no limit.")
}
