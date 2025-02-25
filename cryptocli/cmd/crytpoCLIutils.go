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
	"fmt"
	"os"
	"path/filepath"
)

const DefaultTokenFilename = "crypto_token.txt"

const CryptoCLIDataSubdir = "cryptocli.data"

func GetDataDir() (string, error) {
	baseDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cryptoCLIDir := filepath.Join(baseDir, CryptoCLIDataSubdir)

	_, err = os.Stat(cryptoCLIDir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(cryptoCLIDir, 0755)
		if errDir != nil {
			return "", err
		}
	}
	return cryptoCLIDir, nil
}

func GetEndPoint(server string, version string, action string) string {
	if server == "" {
		server = GetServer()
	}
	return fmt.Sprintf("https://%s/token/%s/%s/", server, version, action)
}

func GetEndPoint2(server string, version string, action string) string {
	if server == "" {
		server = GetServer()
	}
	return fmt.Sprintf("https://%s/token/%s/%s", server, version, action)
}



func AuthTokenKV() map[string]string {
	return map[string]string{"X-TOKEN-AUTH": GetAccessToken()}
}
