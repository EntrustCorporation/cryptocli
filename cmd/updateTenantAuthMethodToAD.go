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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	// custom
	"cli/getpasswd"
	
	// external
	"github.com/spf13/cobra"
)

const (
	authADDomainName     = "ad-domain-name"
	adDomainType         = "ad-domain-type"
	adServiceAccountName = "ad-service_account_name"
	adServiceAccountPw   = "ad-service-account-pw"
	adUID                = "ad-uid"
	adServers            = "ad-servers"

	Name = "name"

	initialADMemberCN                = "initial-ad-member-cn"
	initialADMemberDistinguishedName = "initial-ad-member-distinguished-name"
	initialADMemberMail              = "initial-ad-member-mail"
	initialADMemberUPN               = "initial-ad-member-upn"
)

type AuthToADServer struct {
	ServerURL   string `json:"server_url"`
	TLS         bool   `json:"tls,omitempty"`
	UserBaseDN  string `json:"user_base_dn,omitempty"`
	GroupBaseDN string `json:"group_base_dn,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
	CACert      string `json:"cacert,omitempty"`
}

type ADDomain struct {
	DomainName         string           `json:"domain_name,omitempty"`
	DomainType         string           `json:"type,omitempty"`
	ServiceAccountName string           `json:"service_account_name,omitempty"`
	ServiceAccountPw   string           `json:"service_account_pw,omitempty"`
	UID                string           `json:"uid,omitempty"`
	Servers            []AuthToADServer `json:"servers,omitempty"`
}

type InitialADmember struct {
	CN                string `json:"cn,omitempty"`
	DistinguishedName string `json:"distinguishedName,omitempty"`
	Mail              string `json:"mail,omitempty"`
	UPN               string `json:"upn,omitempty"`
}


type updateTenentAuthModeToADApiResponse struct {
	Result string `json:"result"`
}

func adDomainParseServersJSONFile(JSONFile string, v interface{}) error {
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

func adDomainParseCACertFile(Servers []AuthToADServer) error {
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

func getADServiceCredentials(prefix, user, password string) (string, string) {

	if user == "" {
		fmt.Printf("%s User Name: ", prefix)
		reader := bufio.NewReader(os.Stdin)
		user, _ = reader.ReadString('\n')
		// the returned string includes \r\n on Windows, \n on *nix
		if strings.HasSuffix(user, "\n") {
			user = user[:len(user)-1]
		}
		if strings.HasSuffix(user, "\r") {
			user = user[:len(user)-1]
		}
	}

	if password == "" {
		fmt.Printf("%s Password: ", prefix)
		password = getpasswd.ReadPassword()
		fmt.Printf("\n")
	}

	return user, password
}

func updateTenantAuthMethodToAD(cmd *cobra.Command, args []string) {

	params := map[string]interface{}{}
	var adDomain ADDomain
	var initialADmember InitialADmember

	flags := cmd.Flags()

	serviceUsername, _ := flags.GetString(adServiceAccountName)
	servicePassword, _ := flags.GetString(adServiceAccountPw)
	if servicePassword == "" || serviceUsername == "" {
		fmt.Printf("\n")
		serviceUsername, servicePassword = getADServiceCredentials("Service", serviceUsername, servicePassword)
	}

	adDomain.DomainName, _ = flags.GetString(authADDomainName)
	adDomain.DomainType, _ = flags.GetString(adDomainType)
	adDomain.ServiceAccountName = serviceUsername
	adDomain.ServiceAccountPw = servicePassword
	adDomain.UID, _ = flags.GetString(adUID)

	JSONFile, _ := flags.GetString(adServers)
	if JSONFile != "" {
		err := adDomainParseServersJSONFile(JSONFile, &adDomain.Servers)
		if err != nil {
			os.Exit(1)
		}
		err = adDomainParseCACertFile(adDomain.Servers)
		if err != nil {
			os.Exit(1)
		}
	}

	initialADmember.CN, _ = flags.GetString(initialADMemberCN)
	initialADmember.DistinguishedName, _ = flags.GetString(initialADMemberDistinguishedName)
	initialADmember.Mail, _ = flags.GetString(initialADMemberMail)
	initialADmember.UPN, _ = flags.GetString(initialADMemberUPN)

	params["ad_domain"] = adDomain

	params["name"], _ = flags.GetString(Name)

	params["initial_admember"] = initialADmember


	jsonParams, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Error building JSON request: ", err)
		os.Exit(1)
	}

	var respData updateTenentAuthModeToADApiResponse
	endpoint := GetEndPoint("", "1.0", "UpdateTenantAuthMethodToAD")
	_, err = DoPost2(endpoint, GetCACertFile(),
		AuthTokenKV(),
		jsonParams, ContentTypeJSON, &respData, nil)
	if err != nil {
		fmt.Printf("\nUpdating auth method to Active Directory failed:\n\n%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAuth method updated to Active Directory successfully.\n")
	fmt.Printf("\nResult : %s\n", respData.Result)
}

func init() {

	var updateTenantAuthMethodToADCommand = &cobra.Command{
		Use:   "update-tenant-auth-method-to-ad",
		Short: "Update Tenant Auth Method To AD",
		Run:   updateTenantAuthMethodToAD,
	}
	updateTenantAuthMethodToADCommand.Flags().StringP(authADDomainName, "a", "",
		"Active Directory Name")
	updateTenantAuthMethodToADCommand.Flags().StringP(adUID, "u", "",
		"Active Directory UID Attribute")
	updateTenantAuthMethodToADCommand.Flags().StringP(adDomainType, "t", "",
		"Active Directory domain type. Supported values: microsoft_Ad, openLDAP")
	updateTenantAuthMethodToADCommand.Flags().StringP(adServiceAccountName, "s", "",
		"Active Directory Service Account User Name. "+
		"Users have the option to input values either through the console or by using a flag "+
			"To clear, set it to \"unset\".")
	updateTenantAuthMethodToADCommand.Flags().StringP(adServiceAccountPw, "p", "",
		"Active Directory Service Account Password. "+
		"Users have the option to input values either through the console or by using a flag")
	updateTenantAuthMethodToADCommand.Flags().StringP(adServers, "j", "",
		"Path to the Active Directory Domain server List JSON File. The file should  "+
			"contain an array of JSON objects, each object representing a Domain Controller. "+
			"Following keys are supported within each domain controller JSON object. \n"+
			"server_url (mandatory) full url of the Domain Controller\n"+
			"cacert (optional) path to CA Certificiate to verify with\n"+
			"user_base_dn (optional) user base DN\n"+
			"group_base_dn (optional) group base DN\n"+
			"timeout (optional) connection timeout in seconds, defaults to 5 seconds\n"+
			"tls (optional) enable StartTLS or not, defaults to false\n"+
			"\n"+
			"Example: \n"+
			"\n"+
			"[\n"+
			"    {\n"+
			"        \"server_url\": \"ldaps://dc1.mycompany.eng.com\",\n"+
			"        \"cacert\": \"/root/cacert.pem\",\n"+
			"        \"user_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n"+
			"        \"group_base_dn\": \"DC=mycompany,DC=eng,DC=com\",\n"+
			"        \"timeout\": 10,\n"+
			"        \"tls\": false,\n"+
			"    }\n"+
			"]\n")

	updateTenantAuthMethodToADCommand.Flags().StringP(Name, "n", "",
		"Name of the tenant to be updated")

	updateTenantAuthMethodToADCommand.Flags().StringP(initialADMemberCN, "k", "",
		"Initial Active Directory member CN")
	updateTenantAuthMethodToADCommand.Flags().StringP(initialADMemberDistinguishedName, "d", "",
		"Initial Active Directory member distinguished name")
	updateTenantAuthMethodToADCommand.Flags().StringP(initialADMemberMail, "m", "",
		"Initial Active Directory member mail")
	updateTenantAuthMethodToADCommand.Flags().StringP(initialADMemberUPN, "o", "",
		"Initial Active Directory member UPN")

	// mark mandatory fields as required
	updateTenantAuthMethodToADCommand.MarkFlagRequired(authADDomainName)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(adUID)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(adDomainType)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(adServers)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(Name)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(initialADMemberMail)
	updateTenantAuthMethodToADCommand.MarkFlagRequired(initialADMemberUPN)

	rootCmd.AddCommand(updateTenantAuthMethodToADCommand)
}
