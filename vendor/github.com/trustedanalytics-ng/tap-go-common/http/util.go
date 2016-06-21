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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gocraft/web"
)

type MessageResponse struct {
	Message string `json:"message"`
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func UuidToShortDnsName(uuid string) string {
	if len(uuid) < 15 {
		return "x" + strings.Replace(uuid, "-", "", -1)
	}
	return "x" + strings.Replace(uuid[0:15], "-", "", -1)
}

func ReadJsonFromByte(content []byte, retstruct interface{}) error {
	var err error
	body, err := ioutil.ReadAll(bytes.NewReader(content))
	if err != nil {
		logger.Error("Error reading content:", err)
		return err
	}
	b := []byte(body)
	err = json.Unmarshal(b, &retstruct)
	if err != nil {
		logger.Error("Error parsing content as json:", err)
		return err
	}
	logger.Debug("Content parsed as JSON: ", retstruct)
	return nil
}

func ReadJson(req *web.Request, retstruct interface{}) error {
	var err error
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("Error reading request body:", err)
		return err
	}
	b := []byte(body)
	err = json.Unmarshal(b, &retstruct)
	if err != nil {
		logger.Error("Error parsing request body json:", err)
		return err
	}
	logger.Debug("Request JSON parsed as: ", retstruct)
	return nil
}

func WriteJson(rw web.ResponseWriter, response interface{}, status_code int) error {
	b, err := json.Marshal(&response)
	if err != nil {
		logger.Error("Error marshalling response:", err)
		return err
	}
	rw.Header().Set("Content-Type", "application/json")
	logger.Debug("Responding with status", status_code, " and JSON:", string(b))
	rw.WriteHeader(status_code)
	_, err = fmt.Fprintf(rw, "%s", string(b))
	return err
}

func WriteJsonOrError(rw web.ResponseWriter, response interface{}, status int, err error) error {
	responseStatus := GetHttpStatusOrStatusError(status, err)
	if responseStatus >= http.StatusBadRequest {
		GenericRespond(responseStatus, rw, err)
		return err
	}
	return WriteJson(rw, response, responseStatus)
}

func Respond500(rw web.ResponseWriter, err error) {
	GenericRespond(http.StatusInternalServerError, rw, err)
}

func Respond400(rw web.ResponseWriter, err error) {
	GenericRespond(http.StatusBadRequest, rw, err)
}

func Respond403(rw web.ResponseWriter) {
	GenericRespond(http.StatusForbidden, rw, errors.New("Access Forbidden"))
}

func Respond404(rw web.ResponseWriter, err error) {
	GenericRespond(http.StatusNotFound, rw, err)
}

func Respond409(rw web.ResponseWriter, err error) {
	GenericRespond(http.StatusConflict, rw, err)
}

func GenericRespond(code int, rw web.ResponseWriter, err error) {
	logger.Error(fmt.Sprintf("Respond %d, reason: %v", code, err))
	WriteJson(rw, MessageResponse{err.Error()}, code)
}

func RespondUnauthorized(rw web.ResponseWriter) {
	rw.Header().Set("WWW-Authenticate", `Basic realm=""`)
	rw.WriteHeader(http.StatusUnauthorized)
	rw.Write([]byte("401 Unauthorized\n"))
}

//In order to get rid of reapeting 'return' statement all cases has to be handled in if{}else{}
func HandleError(rw web.ResponseWriter, err error) {
	logger.Debug("handling error", err)
	if IsNotFoundError(err) {
		Respond404(rw, err)
	} else if IsAlreadyExistsError(err) || IsConflictError(err) {
		Respond409(rw, err)
	} else if IsBadRequestError(err) {
		Respond400(rw, err)
	} else {
		Respond500(rw, err)
	}
}

func RespondErrorByStatus(rw web.ResponseWriter, statusCode int, operationName string) {
	if statusCode == http.StatusForbidden {
		Respond403(rw)
	} else {
		GenericRespond(statusCode, rw, fmt.Errorf("error doing: %s", operationName))
	}
}
