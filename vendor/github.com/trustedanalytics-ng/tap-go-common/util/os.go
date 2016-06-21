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
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"

	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

var logger, _ = commonLogger.InitLogger("util")

func GetTerminationObserverChannel() chan os.Signal {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	return channel
}

func TerminationObserver(waitGroup *sync.WaitGroup, appName string) {
	<-GetTerminationObserverChannel()
	logger.Info(appName, "is going to be stopped now...")
	waitGroup.Wait()
	logger.Info(appName, "stopped!")
	os.Exit(0)
}

func GetEnvValueOrDefault(envName string, defaultValue string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return defaultValue
}

func GetStringEnvValueOrDefault(envName string, defaultValue string) (string, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(string), err
}

func GetInt16EnvValueOrDefault(envName string, defaultValue int16) (int16, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(int16), err
}

func GetUint16EnvValueOrDefault(envName string, defaultValue uint16) (uint16, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(uint16), err
}

func GetInt32EnvValueOrDefault(envName string, defaultValue int32) (int32, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(int32), err
}

func GetUint32EnvValueOrDefault(envName string, defaultValue uint32) (uint32, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(uint32), err
}

func GetInt64EnvValueOrDefault(envName string, defaultValue int64) (int64, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(int64), err
}

func GetUint64EnvValueOrDefault(envName string, defaultValue uint64) (uint64, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(uint64), err
}

func GetFloat32EnvValueOrDefault(envName string, defaultValue float32) (float32, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(float32), err
}

func GetFloat64EnvValueOrDefault(envName string, defaultValue float64) (float64, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(float64), err
}

func GetBoolEnvValueOrDefault(envName string, defaultValue bool) (bool, error) {
	v, err := GetTypeEnvValueOrDefault(envName, defaultValue)
	return v.(bool), err
}

func GetTypeEnvValueOrDefault(envName string, defaultValue interface{}) (interface{}, error) {
	var value interface{}
	var valueString string = os.Getenv(envName)
	if valueString == "" {
		return defaultValue, nil
	}

	var err error

	switch defaultValue.(type) {
	case string:
		return valueString, nil
	case uint16:
		var ui64 uint64
		ui64, err = strconv.ParseUint(valueString, 10, 16)
		value = uint16(ui64)
	case int16:
		var i64 int64
		i64, err = strconv.ParseInt(valueString, 10, 16)
		value = int16(i64)
	case uint32:
		var ui64 uint64
		ui64, err = strconv.ParseUint(valueString, 10, 32)
		value = uint32(ui64)
	case int32:
		var i64 int64
		i64, err = strconv.ParseInt(valueString, 10, 32)
		value = int32(i64)
	case uint64:
		value, err = strconv.ParseUint(valueString, 10, 64)
	case int64:
		value, err = strconv.ParseInt(valueString, 10, 64)
	case float32:
		var f64 float64
		f64, err = strconv.ParseFloat(valueString, 32)
		value = float32(f64)

	case float64:
		value, err = strconv.ParseFloat(valueString, 64)
	case bool:
		value, err = strconv.ParseBool(valueString)
	default:
		err = fmt.Errorf("no type matching for %v", defaultValue)
	}

	if err != nil {
		return defaultValue, err
	}
	return value, nil
}
