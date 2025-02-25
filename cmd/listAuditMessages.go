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
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	// external
	"github.com/spf13/cobra"
)

const (
	listAuditOptionJSONOutput  = "json-output"
	listAuditOptionIncludeInfo = "include-info"
	listAuditOptionLocalTime   = "local-time"
)

type auditMessage struct {
	CreatedAt   string                 `json:"created_at"`
	UserContext string                 `json:"user_context"`
	Message     string                 `json:"message"`
	Info        map[string]interface{} `json:"info,omitempty"`
}

type listAuditMessagesResponse struct {
	Messages []auditMessage `json:"audit_messages"`
}

func printAuditMessages(data []byte, cmd *cobra.Command) {
	flags := cmd.Flags()

	// JSON output
	if ok, _ := flags.GetBool(listAuditOptionJSONOutput); ok {
		dst := &bytes.Buffer{}
		if err := json.Indent(dst, data, "", "  "); err != nil {
			// probably this is not of json format, print & exit
			fmt.Println(string(data))
			os.Exit(4)
		} else {
			fmt.Println(dst.String())
		}
		return
	}

	// text line-by-line custom output
	// without additional info:
	// (timestamp user-context message)
	// with additional info:
	// (timestamp user-context message)
	// name: value
	// name: value
	var resp listAuditMessagesResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Printf("\nInvalid response - %v\n", err)
		return
	}

	includeInfo, _ := flags.GetBool(listAuditOptionIncludeInfo)
	localTime, _ := flags.GetBool(listAuditOptionLocalTime)
	var location *time.Location
	if localTime {
		location, err = time.LoadLocation("Local")
		if err != nil {
			// lets print timestamp in UTC instead of failing the call
			localTime = false
		}
	}

	for _, msg := range resp.Messages {
		var createdAt time.Time
		createdAt, err = time.Parse(time.RFC3339, msg.CreatedAt)
		if err != nil {
			continue
		}

		var ts string
		if localTime {
			createdAt = createdAt.In(location)
			ts = createdAt.Format(time.RFC1123)
		} else {
			ts = createdAt.Format(time.RFC1123)
		}

		fmt.Printf("%s %s %s\n", ts, msg.UserContext, msg.Message)
		if !includeInfo || msg.Info == nil {
			continue
		}
		for name, value := range msg.Info {
			fmt.Printf("%s: %v\n", name, value)
		}
	}
}

func listAuditMessageAPI(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	params := map[string]interface{}{}

	// create request payload
	if flags.Changed("filters") {
		filters, _ := flags.GetString("filters")
		params["filters"] = filters
	}

	if flags.Changed("max-items") {
		maxItems, _ := flags.GetInt("max-items")
		params["max_items"] = maxItems
	}

	if flags.Changed("field") {
		fieldsArray, _ := flags.GetStringArray("field")
		params["fields"] = fieldsArray
	}

	if flags.Changed("next-token") {
		nextToken, _ := flags.GetString("next-token")
		params["next_token"] = nextToken
	}

	// JSONify
	jsonParams, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	}

	// now POST
	endpoint := GetEndPoint("", "1.0", "ListAuditMessages")
	ret, err := DoPost2(endpoint, GetCACertFile(),
		AuthTokenKV(),
		jsonParams, ContentTypeJSON, nil, nil)
	if err != nil {
		var apiError APIError
		if errors.As(err, &apiError) && apiError.HttpStatusCode == http.StatusNotFound {
			fmt.Println("\nAudit messages not found\n")
			os.Exit(5)
		}
		fmt.Printf("\nHTTP request failed:\n%v\n", err)
		os.Exit(4)
	} else {
		data := ret.([]byte)
		if len(data) == 0 {
			fmt.Printf("\nEmpty response\n")
			os.Exit(3)
		} else {
			printAuditMessages(data, cmd)
			os.Exit(0)
		}
	}
}

// listAuditMessagesCmd represents the list-audit-message command
var listAuditMessagesCmd = &cobra.Command{
	Use:   "list-audit-messages",
	Short: "List all audit messages",
	Run:   listAuditMessageAPI,
}

func init() {
	rootCmd.AddCommand(listAuditMessagesCmd)
	listAuditMessagesCmd.Flags().StringP("filters", "l", "",
		"Conditional filter expression to apply")
	listAuditMessagesCmd.Flags().IntP("max-items", "m", 0,
		"Maximum number of items to include in "+
			"response")
	listAuditMessagesCmd.Flags().StringArrayP("field", "f", []string{},
		"Audit message fields to include "+
			"in the response")
	listAuditMessagesCmd.Flags().StringP("next-token", "n", "",
		"Token from which subsequent Audit "+
			"messages would be listed")
	listAuditMessagesCmd.Flags().BoolP(listAuditOptionJSONOutput, "j", false,
		"Show JSON formatted audit messages")
	listAuditMessagesCmd.Flags().BoolP(listAuditOptionIncludeInfo, "i", false,
		"Show Additional Information if available")
	listAuditMessagesCmd.Flags().BoolP(listAuditOptionLocalTime, "t", false,
		"Convert audit message timestamp to local time.")
}
