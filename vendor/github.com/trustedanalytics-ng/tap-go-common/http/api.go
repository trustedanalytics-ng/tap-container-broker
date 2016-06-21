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
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiConnector struct {
	BasicAuth *BasicAuth
	OAuth2    *OAuth2
	Client    *http.Client
	Url       string
}

type CallFunc func(url string, requestBody, authHeader string, client *http.Client) (int, []byte, error)

func GetModel(apiConnector ApiConnector, expectedStatus int, result interface{}) (int, error) {
	return callModelTemplateWithBody(RestGETWithBody, apiConnector, "", expectedStatus, result)
}

func PatchModel(apiConnector ApiConnector, requestBody interface{}, expectedStatus int, result interface{}) (int, error) {
	return callModelTemplateWithBody(RestPATCH, apiConnector, requestBody, expectedStatus, result)
}

func PostModel(apiConnector ApiConnector, requestBody interface{}, expectedStatus int, result interface{}) (int, error) {
	return callModelTemplateWithBody(RestPOST, apiConnector, requestBody, expectedStatus, result)
}

func PutModel(apiConnector ApiConnector, requestBody interface{}, expectedStatus int, result interface{}) (int, error) {
	return callModelTemplateWithBody(RestPUT, apiConnector, requestBody, expectedStatus, result)
}

func DeleteModel(apiConnector ApiConnector, expectedStatus int) (int, error) {
	return callModelTemplateWithBody(RestDELETE, apiConnector, "", expectedStatus, "")
}

func DeleteModelWithBody(apiConnector ApiConnector, requestBody interface{}, expectedStatus int) (int, error) {
	return callModelTemplateWithBody(RestDELETE, apiConnector, requestBody, expectedStatus, "")
}

func callModelTemplateWithBody(callFunc CallFunc, apiConnector ApiConnector, requestBody interface{}, expectedStatus int, result interface{}) (status int, err error) {
	requestBodyByte := []byte{}

	if requestBody != "" {
		requestBodyByte, err = json.Marshal(requestBody)
		if err != nil {
			return http.StatusBadRequest, err
		}
	}

	authHeader := ""
	if apiConnector.OAuth2 != nil {
		authHeader = GetOAuth2Header(apiConnector.OAuth2)
	}
	if apiConnector.BasicAuth != nil {
		authHeader = GetBasicAuthHeader(apiConnector.BasicAuth)
	}

	status, body, err := callFunc(apiConnector.Url, string(requestBodyByte), authHeader, apiConnector.Client)
	if err != nil {
		return status, err
	}

	if status != expectedStatus {
		return status, getWrongStatusError(status, expectedStatus, string(body))
	}

	if result != "" {
		err = json.Unmarshal(body, result)
		if err != nil {
			return status, err
		}
	}

	return status, nil
}

func getWrongStatusError(status, expectedStatus int, body string) error {
	return fmt.Errorf("Bad response status: %d, expected status was: % d. Response body: %s", status, expectedStatus, body)
}
