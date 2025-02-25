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

// UpdateADSetting command line arguments
const (
	adDomainName             = "domain-name"
	adSettingID              = "ad-setting-id"
	adSettingUIDAttribute    = "uid-attribute"
	adSettingType            = "type"
    adSettingNetBIOSName     = "netbios-name"
	adSettingServiceAccount  = "service-account"
	adSettingServicePassword = "service-password"
	adSettingServersJSONFile = "servers-json-file"
	adSettingRevision        = "revision"
)

type ADServer struct {
	ServerURL   string `json:"server_url"`
	TLS         bool   `json:"tls,omitempty"`
	UserBaseDN  string `json:"user_base_dn,omitempty"`
	GroupBaseDN string `json:"group_base_dn,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
	CACert      string `json:"cacert,omitempty"`
}

// UpdateADSetting REST API Request
// (and possible for CreateADSetting in the future)
type ADSetting struct {
	Name            string          `json:"name,omitempty"`
	ID              string          `json:"ad_setting_id,omitempty"`
	Revision        int             `json:"revision,omitempty"`
	ServiceAccount  string          `json:"service_account,omitempty"`
	ServicePassword string          `json:"service_password,omitempty"`
	UIDAttribute    string          `json:"uid_attribute,omitempty"`
	Type            string          `json:"type,omitempty"`
	NetBIOSName     string          `json:"netbios_name,omitempty"`
	Servers         []ADServer `json:"servers,omitempty"`
}

// listADSettingsCmd represents the list-box command
var listADSettingsCmd = &cobra.Command{
	Use:   "list-ad-settings",
	Short: "List all AD settings",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		params := map[string]interface{}{}

		// create request payload
		if flags.Changed("prefix") {
			prefix, _ := flags.GetString("prefix")
			params["prefix"] = prefix
		}
		
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
		endpoint := GetEndPoint("", "1.0", "ListADSettings")
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
				fmt.Println("\nAD Settings not found\n")
				os.Exit(5)
			}

			fmt.Println("\n" + retStr + "\n")

			// make a decision on what to exit with
			retMap := JsonStrToMap(retStr)
			if _, present := retMap["error"]; present {
				os.Exit(3)
			} else {
				os.Exit(0)
			}
		}
	},
}

func parseServersJSONFile(JSONFile string, v interface{}) error {
	file, err := os.Open(JSONFile)
	if err != nil {
		fmt.Printf("Error opening file %s - %v\n", JSONFile, err)
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(v)
	if err != nil {
		fmt.Printf("Error decoding Servers JSON - %v\n", err)
		return err
	}
	return nil
}

func parseCACertFile(Servers []ADServer) error {
	for idx := range Servers {
		if Servers[idx].CACert != "" {
			encoded, err := LoadAndEncodeCACertFile(Servers[idx].CACert)
			if err != nil {
				fmt.Printf("%v\n", err)
				return err
			}
			Servers[idx].CACert = encoded
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Commmand: update-ad-setting

// UpdateADSetting REST API Response
type updateADSettingResponse struct {
	Name     string `json:"name"`
	ID       string `json:"ad_setting_id"`
	Revision int    `json:"revision"`
}

func updateADSetting(cmd *cobra.Command, args []string) {

	var adSetting ADSetting

	// Build adSetting from Command Line Arguments
	flags := cmd.Flags()

	serviceAccountReset := false
	adSetting.ID, _ = flags.GetString(adSettingID)
	adSetting.Revision, _ = flags.GetInt(adSettingRevision)
	adSetting.ServiceAccount, _ = flags.GetString(adSettingServiceAccount)
	if adSetting.ServiceAccount == "unset" {
		serviceAccountReset = true
	}
	adSetting.ServicePassword, _ = flags.GetString(adSettingServicePassword)
	adSetting.UIDAttribute, _ = flags.GetString(adSettingUIDAttribute)
	adSetting.Type, _ = flags.GetString(adSettingType)
	adSetting.NetBIOSName, _ = flags.GetString(adSettingNetBIOSName)

	// Servers JSON
	JSONFile, _ := flags.GetString(adSettingServersJSONFile)
	if JSONFile != "" {
		err := parseServersJSONFile(JSONFile, &adSetting.Servers)
		if err != nil {
			os.Exit(1)
		}
		// parse the CACert if provided
		err = parseCACertFile(adSetting.Servers)
		if err != nil {
			os.Exit(1)
		}
	}

	// one of the AD Setting attribute must be updated
	if adSetting.ServiceAccount == "" && adSetting.ServicePassword == "" &&
        adSetting.UIDAttribute == "" && adSetting.NetBIOSName == "" &&
	adSetting.Type == "" && len(adSetting.Servers) == 0 {
		fmt.Printf("\nOne or more parameter(s) is missing. Usage below.\n\n")
		cmd.Help()
		os.Exit(1)
	}

	// Encode request param into JSON
	jsonParams, err := json.Marshal(adSetting)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	}

	// this is handle if service account needs to be cleared
	if serviceAccountReset {
		params := JsonStrToMap(string(jsonParams))
		params["service_account"] = ""

		// JSONify
		jsonParams, err = json.Marshal(params)
		if (err != nil) {
		    fmt.Println("Error building JSON request: ", err)
		    os.Exit(1)
		}
	}

	// send request
	var respData updateADSettingResponse
	endpoint := GetEndPoint("", "1.0", "UpdateADSetting")
	_, err = DoPost2(endpoint, GetCACertFile(),
		AuthTokenKV(),
		jsonParams, ContentTypeJSON, &respData, nil)
	if err != nil {
		fmt.Printf("\nUpdating Active Directory Setting failed:\n\n%v\n", err)
		os.Exit(1)
	}

	// print the response
	fmt.Printf("\nActive Directory Setting updated successfully.\n")
	fmt.Printf("\nRevision : %d\n", respData.Revision)
	fmt.Printf("Name     : %s\n", respData.Name)
	fmt.Printf("ID       : %s\n", respData.ID)
}

////////////////////////////////////////////////////////////////////////////////
// Commmand: change-ad-domain

func changeADDomain(cmd *cobra.Command, args []string) {

	var adSetting ADSetting

	// Build adSetting from Command Line Arguments
	flags := cmd.Flags()

	adSetting.Name, _ = flags.GetString(adDomainName)
	adSetting.Type, _ = flags.GetString(adSettingType)
	adSetting.ServiceAccount, _ = flags.GetString(adSettingServiceAccount)
	adSetting.ServicePassword, _ = flags.GetString(adSettingServicePassword)
	adSetting.UIDAttribute, _ = flags.GetString(adSettingUIDAttribute)
	adSetting.NetBIOSName, _ = flags.GetString(adSettingNetBIOSName)

	// Servers JSON
	JSONFile, _ := flags.GetString(adSettingServersJSONFile)
	if JSONFile != "" {
		err := parseServersJSONFile(JSONFile, &adSetting.Servers)
		if err != nil {
			os.Exit(1)
		}
		// parse the CACert if provided
		err = parseCACertFile(adSetting.Servers)
		if err != nil {
			os.Exit(1)
		}
	}

	// Encode request param into JSON
	jsonParams, err := json.Marshal(adSetting)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	}

	// send request
	var respData updateADSettingResponse
	endpoint := GetEndPoint("", "1.0", "ChangeADDomain")
	_, err = DoPost2(endpoint, GetCACertFile(),
		AuthTokenKV(),
		jsonParams, ContentTypeJSON, &respData, nil)
	if err != nil {
		fmt.Printf("\nChanging AD Domain failed:\n\n%v\n", err)
		os.Exit(1)
	}

	// print the response
	fmt.Printf("\nAD Domain changed successfully.\n")
	fmt.Printf("\nRevision : %d\n", respData.Revision)
	fmt.Printf("Name     : %s\n", respData.Name)
	fmt.Printf("ID       : %s\n", respData.ID)
}

func init() {
	// Commmand: list-ad-setting
	rootCmd.AddCommand(listADSettingsCmd)
	listADSettingsCmd.Flags().StringP("prefix", "p", "",
		"List only those AD settings prefixed with this "+
			"string")
	listADSettingsCmd.Flags().StringP("filters", "l", "",
		"Conditional expression to filter "+
			"AD settings")
	listADSettingsCmd.Flags().IntP("max-items", "m", 0,
		"Maximum number of items to include in "+
			"response")
	listADSettingsCmd.Flags().StringArrayP("field", "f", []string{},
		"AD setting fields to include in the response")
	listADSettingsCmd.Flags().StringP("next-token", "n", "",
		"Token from which subsequent AD settings would "+
			"be listed")

	// Commmand: update-ad-setting
	var updateADSettingsCommand = &cobra.Command{
		Use:   "update-ad-settings",
		Short: "Update Active Directory Settings",
		Run:   updateADSetting,
	}
	updateADSettingsCommand.Flags().StringP(adSettingID, "a", "",
		"Active Directory Setting ID or Name")
	updateADSettingsCommand.Flags().StringP(adSettingUIDAttribute, "u", "",
		"Active Directory UID Attribute")
	updateADSettingsCommand.Flags().StringP(adSettingType, "t", "",
		"Domain type")
	updateADSettingsCommand.Flags().StringP(adSettingNetBIOSName, "n", "",
		"Active Directory NetBIOS name")
	updateADSettingsCommand.Flags().StringP(adSettingServiceAccount, "s", "",
		"Active Directory Service Account User Name. " +
	        "To clear, set it to \"unset\".")
	updateADSettingsCommand.Flags().StringP(adSettingServicePassword, "p", "",
		"Active Directory Service Account Password")
	updateADSettingsCommand.Flags().StringP(adSettingServersJSONFile, "j", "",
		"Active Directory Domain Controller List JSON File. This is to be " +
        "a array of JSON objects, each object representing a Domain Controller. " +
        "Following keys are supported, \n" +
        "server_url (mandatory) full url of the Domain Controller\n" +
        "cacert (optional) path to CA Certificiate to verify with\n" +
        "user_base_dn (optional) user base DN\n" +
        "group_base_dn (optional) group base DN\n" +
        "timeout (optional) connection timeout in seconds, defaults to 5 seconds\n" +
        "tls (optional) enable StartTLS or not, defaults to false\n" +
        "\n" +
        "Example: \n" +
        "\n" +
        "[\n" +
        "    {\n" +
        "        \"server_url\": \"ldaps://dc1.mycompany.eng.com\",\n" +
        "        \"cacert\": \"/root/cacert.pem\",\n" +
        "        \"user_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n" +
        "        \"group_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n" +
        "        \"timeout\": 10,\n" +
        "        \"tls\": false,\n" +
        "    }\n" +
        "]\n")
	updateADSettingsCommand.Flags().IntP(adSettingRevision, "r", 0,
		"Active Directory Setting Current Revision")
	// mark mandatory fields as required
	updateADSettingsCommand.MarkFlagRequired(adSettingID)
	updateADSettingsCommand.MarkFlagRequired(adSettingRevision)
	rootCmd.AddCommand(updateADSettingsCommand)

	// Commmand: change-ad-domain
	var changeADDomainCommand = &cobra.Command{
		Use:   "change-ad-domain",
		Short: "Change Active Directory Domain",
		Run:   changeADDomain,
	}
	changeADDomainCommand.Flags().StringP(adDomainName, "d", "",
		"Active Directory Domain Name")
	changeADDomainCommand.Flags().StringP(adSettingUIDAttribute, "u", "",
		"Active Directory UID Attribute")
	changeADDomainCommand.Flags().StringP(adSettingType, "t", "",
		"Domain type")
	changeADDomainCommand.Flags().StringP(adSettingNetBIOSName, "n", "",
		"Active Directory NetBIOS name")
	changeADDomainCommand.Flags().StringP(adSettingServiceAccount, "s", "",
		"Active Directory Service Account User Name. " +
	        "To clear, set it to \"unset\".")
	changeADDomainCommand.Flags().StringP(adSettingServicePassword, "p", "",
		"Active Directory Service Account Password")
	changeADDomainCommand.Flags().StringP(adSettingServersJSONFile, "j", "",
		"Active Directory Domain Controller List JSON File. This is to be " +
        "a array of JSON objects, each object representing a Domain Controller. " +
        "Following keys are supported, \n" +
        "server_url (mandatory) full url of the Domain Controller\n" +
        "cacert (optional) path to CA Certificiate to verify with\n" +
        "user_base_dn (optional) user base DN\n" +
        "group_base_dn (optional) group base DN\n" +
        "timeout (optional) connection timeout in seconds, defaults to 5 seconds\n" +
        "tls (optional) enable StartTLS or not, defaults to false\n" +
        "\n" +
        "Example: \n" +
        "\n" +
        "[\n" +
        "    {\n" +
        "        \"server_url\": \"ldaps://dc1.mycompany.eng.com\",\n" +
        "        \"cacert\": \"/root/cacert.pem\",\n" +
        "        \"user_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n" +
        "        \"group_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n" +
        "        \"timeout\": 10,\n" +
        "        \"tls\": false,\n" +
        "    }\n" +
        "]\n")
	// mark mandatory fields as required
	changeADDomainCommand.MarkFlagRequired(adDomainName)
	changeADDomainCommand.MarkFlagRequired(adSettingType)
	changeADDomainCommand.MarkFlagRequired(adSettingServersJSONFile)
	changeADDomainCommand.MarkFlagRequired(adSettingUIDAttribute)
	rootCmd.AddCommand(changeADDomainCommand)
}
