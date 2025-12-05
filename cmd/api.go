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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// ContentTypeJSON defines JSON content type
const (
	ContentTypeJSON = "application/json"
)

func GetTLSConfig(caCertPool *x509.CertPool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, // Skip default verification
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return err
			}

			opts := x509.VerifyOptions{
				Roots:         caCertPool,
				Intermediates: x509.NewCertPool(),
				// DNSName is omitted to skip CN/SAN check
			}

			for _, certBytes := range rawCerts[1:] {
				intermediate, err := x509.ParseCertificate(certBytes)
				if err != nil {
					return err
				}
				opts.Intermediates.AddCert(intermediate)
			}

			_, err = cert.Verify(opts)
			return err
		},
	}
}

// DoPostFormData sends an API request
func DoPostFormData(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string) (map[string]interface{}, error) {

	params := map[string]interface{}{}
	json.Unmarshal(jsonParams, &params)
	values := map[string]io.Reader{}
	
	if _, keyPresent := params["public_key"]; keyPresent {
		values["public_key"] = mustOpen(params["public_key"].(string))
	}
	if _, keyPresent := params["csv_file"]; keyPresent {
		values["csv_file"] = mustOpen(params["csv_file"].(string))
	}
	if _, keyPresent := params["secret_type"]; keyPresent {
		values["secret_type"] = strings.NewReader(params["secret_type"].(string))
	}
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				// this is to avoid Go's unused error message
				_ = fw
				return nil, err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				// this is to avoid Go's unused error message
				_ = fw
				return nil, err
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			return nil, err
		}
	}
	w.Close()
	request, err := http.NewRequest("POST", endpoint, &b)
	if err != nil {
		return nil, err
	}

	// close connection once done
	request.Close = true
	request.Header.Set("Content-Type", w.FormDataContentType())
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	ret := map[string]interface{}{}
	response, err := client.Do(request)
	if err != nil {
		return ret, err
	} else {
		data, _ := io.ReadAll(response.Body)
		dst := &bytes.Buffer{}
		if len(data) != 0 {
			if err := json.Indent(dst, data, "", "  "); err != nil {
				// probably this is not of json format, print & exit
				fmt.Println("\n" + string(data) + "\n")
				os.Exit(4)
			}
		}

		ret["status"] = response.StatusCode
		ret["data"] = dst
		return ret, nil
	}
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

// DoGet sends an API request
func DoGet(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string) (map[string]interface{}, error) {
	request, _ := http.NewRequest("GET", endpoint, bytes.NewBuffer(jsonParams))

	// close connection once done
	request.Close = true
	request.Header.Set("Content-Type", contentType)
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	ret := map[string]interface{}{}
	response, err := client.Do(request)
	if err != nil {
		return ret, err
	} else {
		data, _ := io.ReadAll(response.Body)
		dst := &bytes.Buffer{}
		if len(data) != 0 {
			if err := json.Indent(dst, data, "", "  "); err != nil {
				// probably this is not of json format, print & exit
				fmt.Println("\n" + string(data) + "\n")
				os.Exit(4)
			}
		}

		ret["status"] = response.StatusCode
		ret["data"] = dst
		return ret, nil
	}
}

// DoDelete sends an API request
func DoDelete(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string) (map[string]interface{}, error) {
	request, _ := http.NewRequest("DELETE", endpoint, bytes.NewBuffer(jsonParams))

	// close connection once done
	request.Close = true
	request.Header.Set("Content-Type", contentType)
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	ret := map[string]interface{}{}
	response, err := client.Do(request)
	if err != nil {
		return ret, err
	} else {
		data, _ := io.ReadAll(response.Body)
		dst := &bytes.Buffer{}
		if len(data) != 0 {
			if err := json.Indent(dst, data, "", "  "); err != nil {
				// probably this is not of json format, print & exit
				fmt.Println("\n" + string(data) + "\n")
				os.Exit(4)
			}
		}

		ret["status"] = response.StatusCode
		ret["data"] = dst
		return ret, nil
	}
}

func DoPatch(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string) (map[string]interface{}, error) {
	request, _ := http.NewRequest("PATCH", endpoint, bytes.NewBuffer(jsonParams))

	request.Close = true
	request.Header.Set("Content-Type", contentType)
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	ret := map[string]interface{}{}
	response, err := client.Do(request)
	if err != nil {
		return ret, err
	}
	data, _ := io.ReadAll(response.Body)
	dst := &bytes.Buffer{}
	if len(data) != 0 {
		if err := json.Indent(dst, data, "", "  "); err != nil {
			fmt.Println("\n" + string(data) + "\n")
			os.Exit(4)
		}
	}

	ret["status"] = response.StatusCode
	ret["data"] = dst
	return ret, nil
}

// DoPost sends an API request
func DoPost(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string) (map[string]interface{}, error) {
	request, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonParams))

	// close connection once done
	request.Close = true
	request.Header.Set("Content-Type", contentType)
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	ret := map[string]interface{}{}
	response, err := client.Do(request)
	if err != nil {
		return ret, err
	} else {
		data, _ := io.ReadAll(response.Body)
		dst := &bytes.Buffer{}
		if len(data) != 0 {
			if err := json.Indent(dst, data, "", "  "); err != nil {
				// probably this is not of json format, print & exit
				fmt.Println("\n" + string(data) + "\n")
				os.Exit(4)
			}
		}

		ret["status"] = response.StatusCode
		ret["data"] = dst
		return ret, nil
	}
}

// APIError contains error details of API request failure
type APIError struct {
	RequestURL     string
	HttpStatusCode int    // e.g. 200
	HttpStatus     string // e.g. "200 OK"
	ErrorJSON      []byte // Error Message from the Server
}

func (e APIError) Error() string {
	if len(e.ErrorJSON) > 0 {
		return fmt.Sprintf("%s\n%s\n%s", e.RequestURL, e.HttpStatus,
			string(e.ErrorJSON))
	}
	return fmt.Sprintf("%s\n%s", e.RequestURL, e.HttpStatus)
}

func newAPIError(requestURL string, response *http.Response) APIError {

	var apiError APIError
	apiError.RequestURL = requestURL
	apiError.HttpStatusCode = response.StatusCode
	apiError.HttpStatus = response.Status
	contentType := response.Header.Get("Content-Type")
	if contentType == ContentTypeJSON {
		apiError.ErrorJSON, _ = io.ReadAll(response.Body)
	}
	return apiError
}

// ResponseHandler can process the HTTP Response fully
type ResponseHandler interface {
	ProcessResponse(response *http.Response, responseData interface{},
		request *http.Request, requestURL string) (interface{}, error)
}

// DoPost2 sends an API request
func DoPost2(endpoint string,
	cacert string,
	headers map[string]string,
	jsonParams []byte,
	contentType string,
	responseData interface{},
	respCallback ResponseHandler) (interface{}, error) {

	request, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonParams))

	// close connection once done
	request.Close = true
	request.Header.Set("Content-Type", contentType)
	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	// invoke ResponseHandler if provided
	if cb, ok := respCallback.(ResponseHandler); ok {
		return cb.ProcessResponse(response, responseData, request, endpoint)
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		if responseData == nil {
			return io.ReadAll(response.Body)
		}

		contentType := response.Header.Get("Content-Type")
		if contentType != ContentTypeJSON {
			return nil, fmt.Errorf("Invalid Content-Type: %v", contentType)
		}
		return nil, json.NewDecoder(response.Body).Decode(responseData)

	default:
		return nil, newAPIError(endpoint, response)
	}
}

// DoDownload sends an API request
func DoDownload(endpoint string,
	method string,
	cacert string,
	headers map[string]string,
	jsonParams []byte) (string, error) {
	request, _ := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonParams))
	// close connection once done
	request.Close = true

	for header, value := range headers {
		request.Header.Set(header, value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	response, err := client.Do(request)
	if err != nil {
		return "", err
	} else {
		// server responded with some error - capture & exit
		if response.StatusCode != 200 {
			data, _ := io.ReadAll(response.Body)
			fmt.Println("\n" + string(data) + "\n")
			os.Exit(3)
		}

		fname, err := GetValueFromKVString(
			response.Header.Get("Content-Disposition"), "filename")
		if err != nil {
			return "", err
		}

		// create file
		outFile, err := os.Create(fname)
		if err != nil {
			return "", err
		}
		defer outFile.Close()

		// write body to file
		_, err = io.Copy(outFile, response.Body)
		if err != nil {
			return "", err
		}
		return fname, nil
	}
}

// DoGetDownload sends an API request
func DoGetDownload(endpoint string,
	cacert string,
	headers map[string]string) (string, error) {
	request, _ := http.NewRequest("GET", endpoint, nil)
	// close connection once done
	request.Close = true

	for header, value := range headers {
		request.Header.Set(string(header), value)
	}

	client := &http.Client{}
	tr := &http.Transport{}
	if cacert != "" {
		// Create a CA certificate pool and add cacert to it
		caCert, err := os.ReadFile(cacert)
		if err != nil {
			fmt.Println("Error reading CA Certificate: ", err)
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tr.TLSClientConfig = GetTLSConfig(caCertPool)
	} else {
		fmt.Println("\n###############################################################################\n" +
			"Insecure request. Entrust Vault certificate not verified. \n" +
			"It is strongly recommended to verify the same by specifying CA \n" +
			"Certificate, using the --cacert option, to mitigate Man-in-the-middle attack\n" +
			"###############################################################################")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = tr
	response, err := client.Do(request)
	if err != nil {
		return "", err
	} else {
		// server responded with some error - capture & exit
		if response.StatusCode != 200 {
			data, _ := io.ReadAll(response.Body)
			fmt.Println("\n" + string(data) + "\n")
			os.Exit(3)
		}

		fname, err := GetValueFromKVString(
			response.Header.Get("Content-Disposition"), "filename")
		if err != nil {
			return "", err
		}

		// create file
		outFile, err := os.Create(fname)
		if err != nil {
			return "", err
		}
		defer outFile.Close()

		// write body to file
		_, err = io.Copy(outFile, response.Body)
		if err != nil {
			return "", err
		}
		return fname, nil
	}
}

func GetValueFromKVString(kvString string, key string) (string, error) {
	regex := fmt.Sprintf("(%s)=([a-z]+)", key)
	re := regexp.MustCompile(regex)
	result := re.FindAllStringSubmatchIndex(kvString, -1)
	if len(result) == 0 || len(result[0]) < 5 {
		errMsg := fmt.Sprintf("%s not found", key)
		return "", errors.New(errMsg)
	}
	match := result[0]
	return string(kvString[match[4]:]), nil
}
