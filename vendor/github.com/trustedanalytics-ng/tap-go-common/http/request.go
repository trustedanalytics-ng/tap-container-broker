/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package http

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type BasicAuth struct {
	User     string
	Password string
}

type OAuth2 struct {
	TokenType string
	Token     string
}

func RestGETWithBody(url string, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("GET", url, body, "application/json", authHeader, client)
}

func RestGET(url, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("GET", url, "", "application/json", authHeader, client)
}

func RestPUT(url, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("PUT", url, body, "application/json", authHeader, client)
}

func RestUrlEncodedPOST(url, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("POST", url, body, "application/x-www-form-urlencoded", authHeader, client)
}

func RestPOST(url, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("POST", url, body, "application/json", authHeader, client)
}

func RestDELETE(url, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("DELETE", url, body, "", authHeader, client)
}

func RestPATCH(url, body, authHeader string, client *http.Client) (int, []byte, error) {
	return makeRequest("PATCH", url, body, "application/json-patch+json", authHeader, client)
}

func makeRequest(reqType, url, body, contentType, authHeader string, client *http.Client) (int, []byte, error) {
	logger.Info("Doing:  ", reqType, url)

	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(reqType, url, bytes.NewBuffer([]byte(body)))
	} else {
		req, _ = http.NewRequest(reqType, url, nil)
	}

	req.Header.Add("Authorization", authHeader)
	SetContentType(req, contentType)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("sending http request %v failed: %v", reqType, err))
		return -1, nil, err
	}
	ret_code := resp.StatusCode
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("reading http request %v response body failed: %v", reqType, err))
		return -1, nil, err
	}

	if resp.Header.Get("Content-Type") == "application/octet-stream" {
		logger.Debug("CODE:", ret_code, "BODY: [ Binary Data ]", resp.ContentLength)
	} else {
		logger.Debug("CODE:", ret_code, "BODY:", string(data))
	}

	return ret_code, data, nil
}

func GetBasicAuthHeader(basicAuth *BasicAuth) string {
	if basicAuth == nil {
		return ""
	}
	auth := basicAuth.User + ":" + basicAuth.Password
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}

func GetOAuth2Header(oauth2 *OAuth2) string {
	if oauth2 == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", oauth2.TokenType, oauth2.Token)
}

func SetContentType(req *http.Request, contentType string) {
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
}

func DownloadBinary(url, authHeader string, client *http.Client, dest io.Writer) (int64, error) {
	return binaryStreamRequest(url, authHeader, client, dest)
}

func binaryStreamRequest(url, authHeader string, client *http.Client, dest io.Writer) (int64, error) {
	logger.Info("Doing:  ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("making http request failed", err)
		return -1, err
	}

	req.Header.Add("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("sending http request failed", err)
		return -1, err
	}

	defer resp.Body.Close()
	_, err = io.CopyN(dest, resp.Body, resp.ContentLength)
	if err != nil {
		logger.Error("copying http request response body failed", err)
		return -1, err
	}

	logger.Debug("CODE:", resp.StatusCode, "BODY: [ Binary Data ] Size:", resp.ContentLength)
	return resp.ContentLength, nil
}
