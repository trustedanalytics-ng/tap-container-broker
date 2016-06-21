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
)

func init() {
	RegisterType("Plans", reflect.TypeOf(ServicePlan{}))
	RegisterType("Dependencies", reflect.TypeOf(ServiceDependency{}))
	RegisterType("Metadata", reflect.TypeOf(Metadata{}))
}

type Service struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Bindable    bool          `json:"bindable"`
	TemplateId  string        `json:"templateId"`
	State       ServiceState  `json:"state"`
	Plans       []ServicePlan `json:"plans"`
	AuditTrail  AuditTrail    `json:"auditTrail"`
	Metadata    []Metadata    `json:"metadata"`
	Tags        []string      `json:"tags"`
}

type ServicePlan struct {
	Id           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Cost         string              `json:"cost"`
	Dependencies []ServiceDependency `json:"dependencies"`
	AuditTrail   AuditTrail          `json:"auditTrail"`
}

type ServiceDependency struct {
	Id          string `json:"id"`
	PlanName    string `json:"plan_name"`
	PlanId      string `json:"plan_id"`
	ServiceName string `json:"service_name"`
	ServiceId   string `json:"service_id"`
}

type ServiceState string

const (
	ServiceStateDeploying ServiceState = "DEPLOYING"
	ServiceStateReady     ServiceState = "READY"
	ServiceStateOffline   ServiceState = "OFFLINE"
)

func (servicePlan *ServicePlan) ValidateServicePlanStructCreate() error {
	if servicePlan.Id != "" {
		return GetIdFieldHasToBeEmptyError()
	}
	return nil
}
func (service *Service) ValidateServiceStructCreate() error {
	if service.Id != "" {
		return GetIdFieldHasToBeEmptyError()
	}

	if err := CheckIfMatchingRegexp(service.Name, RegexpDnsLabelLowercase); err != nil {
		return GetInvalidValueError("Name", service.Name, err)
	}

	if len(service.Plans) == 0 {
		return fmt.Errorf("offering should have at least one plan")
	}

	return nil
}
