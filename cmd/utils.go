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
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"math/big"
    "bytes"
	"time"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"strings"
)

const RandomStringCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

type tokenInfo struct {
	AccessToken string `json:"access_token"`
	Server      string `json:"server"`
	CACertFile  string `json:"cacert_file"`
}

var gTokenInfo tokenInfo

func SaveAccessToken(tokenFile, accessToken, server, caCertFile string) (string, error) {
	if tokenFile == "" {
		tokenDir, err := GetDataDir()
		if err != nil {
			return "", err
		}
		tokenFile = filepath.Join(tokenDir, DefaultTokenFilename)
	}

	info := tokenInfo{
		AccessToken: accessToken,
		Server:      server,
		CACertFile:  caCertFile}

	file, err := os.Create(tokenFile)
	if err != nil {
		return tokenFile, err
	}
	defer file.Close()
	return tokenFile, json.NewEncoder(file).Encode(&info)
}

func LoadAccessToken(tokenFile string) (string, error) {
	if tokenFile == "" {
		tokenDir, err := GetDataDir()
		if err != nil {
			return "", err
		}
		tokenFile = filepath.Join(tokenDir, DefaultTokenFilename)
	}

	file, err := os.Open(tokenFile)
	if err != nil {
		return tokenFile, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&gTokenInfo)
	if err != nil {
		return tokenFile, err
	}

	if gTokenInfo.AccessToken == "" || gTokenInfo.Server == "" {
		return tokenFile, fmt.Errorf("Invalid or corrupt Token File - access_token or server is missing")
	}

	return tokenFile, nil
}

func GetAccessToken() string {
	return gTokenInfo.AccessToken
}

func GetServer() string {
	return gTokenInfo.Server
}

func GetCACertFile() string {
	return gTokenInfo.CACertFile
}

func JsonStrToMap(jsonStr string) map[string]interface{} {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonMap)
	if err != nil {
		fmt.Println("Error parsing json string: ", err)
		os.Exit(1)
	}
	return jsonMap
}

func JsonArrayStrToMap(jsonArrayStr string) []map[string]interface{} {
	var jsonArray []map[string]interface{}
	err := json.Unmarshal([]byte(jsonArrayStr), &jsonArray)
	if err != nil {
		fmt.Println("Error parsing json array string: ", err)
		os.Exit(1)
	}
	return jsonArray
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func LoadAndEncodeCACertFile(CACertFile string) (string, error) {
	file, err := os.Open(CACertFile)
	if err != nil {
		return "", fmt.Errorf("Error opening CA Certificate File %s - %v",
			CACertFile, err)
	}
	defer file.Close()

	// Read entire CA CertFile content into a byte slice
	reader := bufio.NewReader(file)
	content, _ := io.ReadAll(reader)

	encoded := base64.StdEncoding.EncodeToString(content)
	if encoded == "" {
		return "", fmt.Errorf("Error encoding certificate data - %s", CACertFile)
	}
	return encoded, nil
}

// equivalent of json.MarshalIndent, while escaping "<", ">" and "&"
// which otherwise would be escaped to their unicode equivalents
// see https://golang.org/pkg/encoding/json/#SetEscapeHTML
func JSONMarshalIndent(t interface{}) ([]byte, error) {
    buffer := &bytes.Buffer{}
    encoder := json.NewEncoder(buffer)
    encoder.SetEscapeHTML(false)
    encoder.SetIndent("", "  ")
    err := encoder.Encode(t)
    return buffer.Bytes(), err
}

func B64File(File string) (string, error) {
	file, err := os.Open(File)
	if err != nil {
		return "", fmt.Errorf("Error opening file %s - %v",
			File, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content, _ := io.ReadAll(reader)

	encoded := base64.StdEncoding.EncodeToString(content)
	if encoded == "" {
		return "", fmt.Errorf("Error base64 encoding file - %s", File)
	}
	return encoded, nil
}

func B64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func B64Decode(b64Str string) string {
	dec, _ := base64.StdEncoding.DecodeString(b64Str)
	return string(dec)
}

func AESGCMDecrypt(b64CipherText string, b64Key string, b64Nonce string,
	b64Tag string, authData string) string {
	key := []byte(B64Decode(b64Key))
	ciphertext := []byte(B64Decode(b64CipherText))
	nonce := []byte(B64Decode(b64Nonce))
	tag := []byte(B64Decode(b64Tag))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

// crypto library expects these together & separates it out by itself
	ciphertextTag := append(ciphertext, tag...)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextTag, []byte(authData))
	if err != nil {
		panic(err.Error())
	}
	return string(plaintext)
}

func KeyExists(kvMap map[string]interface{}, key string) bool {
	val, ok := kvMap[key]
	return ok && val != nil
}

func convertUTCtoLocal() *time.Location {
	location, err := time.LoadLocation("Local")
	if err != nil {
		os.Exit(10) // unexpected
	}
	return location
}

func convertTimeFrom24HourTo12HourFormat(hour int) (int, string) {
	ampm := "AM"
	if hour > 11 {
		ampm = "PM"
	}
	if hour == 0 {
		hour = 12
	} else if hour > 12 {
		hour = hour - 12
	}
	return hour, ampm
}

func formatLoginExpiration(expiresAt string) string {
	expiration, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		fmt.Printf("\nInvalid Expiration %q - %v\n", expiresAt, err)
		os.Exit(5)
	}

	expiration = expiration.In(convertUTCtoLocal())

	hour, min, sec := expiration.Clock()
	hour, ampm := convertTimeFrom24HourTo12HourFormat(hour)

	year, month, day := expiration.Date()

	return fmt.Sprintf("%s, %02d %s %d %02d:%02d:%02d %s", expiration.Weekday(),
		day, month, year, hour, min, sec, ampm)
}

// Generates random string of provided length. Charset used to generate the random string includes upper & lowercase
// alphabets and numerals from 0 to 9. Crypthographically strong random generator is used while generating string so
// that it can be used as password/passphrase.
func GenerateRandomString(length int) string {
	var randomString strings.Builder
	maxIndex := big.NewInt(int64(len(RandomStringCharSet)))
	for i := 0; i < length; i++ {
		randIndex, _ := rand.Int(rand.Reader, maxIndex)
		randomString.WriteByte(RandomStringCharSet[int(randIndex.Int64())])
	}
	return randomString.String()
}
