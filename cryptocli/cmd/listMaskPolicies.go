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
	"github.com/spf13/cobra"
)

var listMaskPoliciesCmd = &cobra.Command{
	Use:   "list-mask-policies",
	Short: "Get list of Mask Policies",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()

		queryString := "?"
		nameFilter, _ := flags.GetString("nameFilter")
		if len(nameFilter) > 0 {
			queryString = queryString + "name=" + nameFilter
		}

		offset, _ := flags.GetString("offset")
		if len(offset) > 0 {
			if len(queryString) > 1 {
				queryString = queryString + "&"
			}
			queryString = queryString + "_offset=" + offset
		}

		counts, _ := flags.GetString("counts")
		if len(counts) > 0 {
			if len(queryString) > 1 {
				queryString = queryString + "&"
			}
			queryString = queryString + "_counts=" + counts
		}

		limit, _ := flags.GetString("limit")
		if len(limit) > 0 {
			if len(queryString) > 1 {
				queryString = queryString + "&"
			}
			queryString = queryString + "_limit=" + limit
		}

		ordering, _ := flags.GetString("sort")
		if len(ordering) > 0 {
			if len(queryString) > 1 {
				queryString = queryString + "&"
			}
			queryString = queryString + "_ordering=" + ordering
		}

		if len(queryString) < 2 {
			queryString = ""
		}

		endpoint := GetEndPoint2("", "1.0", "GetMaskPolicies"+queryString)
		ret, err := DoGet(endpoint,
			GetCACertFile(),
			AuthTokenKV(),
			nil,
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
	rootCmd.AddCommand(listMaskPoliciesCmd)
	listMaskPoliciesCmd.Flags().StringP("nameFilter", "n", "", "Name Filter")
	listMaskPoliciesCmd.Flags().StringP("offset", "o", "", "Offset")
	listMaskPoliciesCmd.Flags().StringP("limit", "l", "", "Limit")
	listMaskPoliciesCmd.Flags().StringP("counts", "c", "", "True if you want count information")
	listMaskPoliciesCmd.Flags().StringP("sort", "s", "", "Sort")
}