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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func GetConnectionHostAndPortFromEnvs(componentName string) (string, int, error) {
	hostEnvName := componentName + "_HOST"
	portEnvName := componentName + "_PORT"

	var errorArray []error

	host, err := GetEnvOrError(hostEnvName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	portStr, err := GetEnvOrError(portEnvName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		errorArray = append(errorArray, fmt.Errorf("port has incorrect value %s: %v", portStr, err))
	}

	if len(errorArray) > 0 {
		return "", -1, sumErrors(errorArray)
	}

	return host, port, nil
}

func GetConnectionAddressFromEnvs(componentName string) (string, error) {
	host, port, err := GetConnectionHostAndPortFromEnvs(componentName)
	if err != nil {
		return "", err
	}

	address := fmt.Sprintf("%s:%d", host, port)

	return address, nil
}

func GetConnectionCredentialsFromEnvs(componentName string) (username, password string, err error) {
	userEnvName := componentName + "_USER"
	passEnvName := componentName + "_PASS"

	var errorArray []error

	username, err = GetEnvOrError(userEnvName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	password, err = GetEnvOrError(passEnvName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	if len(errorArray) > 0 {
		err = sumErrors(errorArray)
	}
	return
}

func GetConnectionParametersFromEnv(componentName string) (address, username, password string, err error) {
	var errorArray []error

	address, err = GetConnectionAddressFromEnvs(componentName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	username, password, err = GetConnectionCredentialsFromEnvs(componentName)
	if err != nil {
		errorArray = append(errorArray, err)
	}

	if len(errorArray) > 0 {
		err = sumErrors(errorArray)
	}
	return
}

func GetEnvOrError(envName string) (string, error) {
	value := os.Getenv(envName)
	if value == "" {
		return value, errors.New(envName + " not set!")
	}
	return value, nil
}

func sumErrors(errorsArray []error) error {
	var buffer bytes.Buffer

	for _, err := range errorsArray {
		buffer.WriteString(fmt.Sprintf("%s, ", err.Error()))
	}

	return errors.New(buffer.String())
}
