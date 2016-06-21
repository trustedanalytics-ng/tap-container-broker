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
	"fmt"
	"reflect"
	"strings"
)

func init() {
	RegisterType("Metadata", reflect.TypeOf(Metadata{}))
	RegisterType("Bindings", reflect.TypeOf(InstanceBindings{}))
}

const (
	BROKER_OFFERING_PREFIX    = "BROKER_OFFERING_"
	BROKER_TEMPLATE_ID        = "BROKER_TEMPLATE_ID"
	APPLICATION_IMAGE_ADDRESS = "APPLICATION_IMAGE_ADDRESS"
	OFFERING_PLAN_ID          = "PLAN_ID"
	LAST_STATE_CHANGE_REASON  = "LAST_STATE_CHANGE_REASON"
)

const ReasonDeleteFailure = "Instance was in FAILURE state. Removing..."

type Instance struct {
	Id         string             `json:"id"`
	Name       string             `json:"name"`
	Type       InstanceType       `json:"type"`
	ClassId    string             `json:"classId"`
	Bindings   []InstanceBindings `json:"bindings"`
	Metadata   []Metadata         `json:"metadata"`
	State      InstanceState      `json:"state"`
	AuditTrail AuditTrail         `json:"auditTrail"`
}

type InstanceState string

const (
	InstanceStateRequested       InstanceState = "REQUESTED"
	InstanceStateDeploying       InstanceState = "DEPLOYING"
	InstanceStateFailure         InstanceState = "FAILURE"
	InstanceStateStopped         InstanceState = "STOPPED"
	InstanceStateStartReq        InstanceState = "START_REQ"
	InstanceStateStarting        InstanceState = "STARTING"
	InstanceStateRunning         InstanceState = "RUNNING"
	InstanceStateReconfiguration InstanceState = "RECONFIGURATION"
	InstanceStateStopReq         InstanceState = "STOP_REQ"
	InstanceStateStopping        InstanceState = "STOPPING"
	InstanceStateDestroyReq      InstanceState = "DESTROY_REQ"
	InstanceStateDestroying      InstanceState = "DESTROYING"
	InstanceStateUnavailable     InstanceState = "UNAVAILABLE"
)

func (state InstanceState) String() string {
	return string(state)
}

type InstanceBindings struct {
	Id   string            `json:"id"`
	Data map[string]string `json:"data"`
}

type Metadata struct {
	Id    string `json:"key"`
	Value string `json:"value"`
}

type InstanceType string

const (
	InstanceTypeApplication   InstanceType = "APPLICATION"
	InstanceTypeService       InstanceType = "SERVICE"
	InstanceTypeServiceBroker InstanceType = "SERVICE_BROKER"
)

func GetValueFromMetadata(metadatas []Metadata, key string) string {
	for _, metadata := range metadatas {
		if metadata.Id == key {
			return metadata.Value
		}
	}
	return ""
}

func GetPrefixedOfferingName(offeringName string) string {
	return BROKER_OFFERING_PREFIX + offeringName
}

func IsServiceBrokerOfferingMetadata(metadata Metadata) bool {
	return strings.Contains(metadata.Id, BROKER_OFFERING_PREFIX)
}

func (instance *Instance) ValidateInstanceStructCreate(instanceType InstanceType) error {
	if instance.Id != "" {
		return GetIdFieldHasToBeEmptyError()
	}

	if instanceType == InstanceTypeService && GetValueFromMetadata(instance.Metadata, OFFERING_PLAN_ID) == "" {
		return fmt.Errorf("key %s not found!", OFFERING_PLAN_ID)
	}

	err := CheckIfMatchingRegexp(instance.Name, RegexpDnsLabelLowercase)
	if err != nil {
		return GetInvalidValueError("Name", instance.Name, err)
	}
	//although it copies for loop from instances.go, in this case we don't query etcd before being sure request is proper
	//in most cases bindings array will be small so no issue with performance should happen here
	for _, binding := range instance.Bindings {
		for k := range binding.Data {
			if err = CheckIfMatchingRegexp(k, RegexpProperSystemEnvName); err != nil {
				return GetInvalidValueError("Data", k, err)
			}
		}
	}

	return nil
}
