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

package logger

import (
	"fmt"
	"os"
	"strings"

	logging "github.com/op/go-logging"
)

const (
	LevelCritical = "CRITICAL"
	LevelError    = "ERROR"
	LevelWarning  = "WARNING"
	LevelNotice   = "NOTICE"
	LevelInfo     = "INFO"
	LevelDebug    = "DEBUG"
)

func validateLogLevel(level string) error {
	levels := []string{LevelCritical, LevelError, LevelWarning, LevelNotice, LevelInfo, LevelDebug}
	for _, x := range levels {
		if strings.ToUpper(x) == strings.ToUpper(level) {
			return nil
		}
	}
	return fmt.Errorf("logger level %q is invalid", level)
}

// InitLogger initializes logger
func InitLogger(module string) (*logging.Logger, error) {
	logger := logging.MustGetLogger(module)

	var logLevel = os.Getenv("BROKER_LOG_LEVEL")
	if logLevel == "" {
		logLevel = LevelCritical
	}
	if err := SetLoggerLevel(logger, logLevel); err != nil {
		return logger, fmt.Errorf("cannot set logger level %q: %v", logLevel, err)
	}

	return logger, nil
}

// SetLoggerLevel sets logger backend with proper log level
func SetLoggerLevel(logger *logging.Logger, level string) error {
	if err := validateLogLevel(level); err != nil {
		return err
	}

	logLevel, err := logging.LogLevel(level)
	if err != nil {
		return err
	}

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)

	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{level:.4s} ▶ [%{shortfunc}]: %{color:reset} %{message}`,
	)

	// For messages written to backend1 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend1Formatter := logging.NewBackendFormatter(backend1, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	backend1Leveled.SetLevel(logLevel, logger.Module)

	// Set the backends to be used.
	logger.SetBackend(backend1Leveled)

	return nil
}
