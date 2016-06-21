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

package models

import (
	"log"
	"time"

	"github.com/trustedanalytics-ng/tap-go-common/util"
)

var (
	ProcessorsTries              = getEnvValueOrDefault("PROCESSORS_TRIES", 12)
	ProcessorsServiceBrokerTries = getEnvValueOrDefault("PROCESSORS_SERVICE_BROKER_TRIES", 60)
	WaitForDepsTries             = ProcessorsTries + 2

	ProcessorsIntervalSec         = time.Second * time.Duration(getEnvValueOrDefault("PROCESSORS_INTERVAL_SEC", 10))
	UpdateDeploymentTimeout       = time.Second * time.Duration(getEnvValueOrDefault("UPDATE_DEPLOYMENT_TIMEOUT", 30))
	WaitForScaledReplicasTimeout  = time.Second * time.Duration(getEnvValueOrDefault("WAIT_FOR_SCALED_REPLICAS_TIMEOUT", 60))
	WaitForScaledReplicasInterval = time.Second * time.Duration(getEnvValueOrDefault("WAIT_FOR_SCALED_REPLICAS_INTERVAL", 5))
)

func getEnvValueOrDefault(envName string, defaultValue int32) int32 {
	value, err := util.GetInt32EnvValueOrDefault(envName, defaultValue)
	if err != nil {
		log.Fatalf("Can't parse env %s value: %v", envName, value)
	}
	return value
}
