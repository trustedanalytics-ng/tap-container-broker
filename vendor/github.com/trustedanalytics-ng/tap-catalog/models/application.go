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
	RegisterType("InstanceDependencies", reflect.TypeOf(InstanceDependency{}))
	RegisterType("Metadata", reflect.TypeOf(Metadata{}))
}

type InstanceDependency struct {
	Id string `json:"id"`
}

type Application struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	Description          string               `json:"description"`
	ImageId              string               `json:"imageId"`
	Replication          int                  `json:"replication"`
	TemplateId           string               `json:"templateId"`
	AuditTrail           AuditTrail           `json:"auditTrail"`
	InstanceDependencies []InstanceDependency `json:"instanceDependencies"`
	Metadata             []Metadata           `json:"metadata"`
}

func (application *Application) ValidateApplicationStructCreate() error {
	if application.Id != "" {
		return GetIdFieldHasToBeEmptyError()
	}
	if application.TemplateId == "" {
		return fmt.Errorf("TemplateId is required")
	}
	err := CheckIfMatchingRegexp(application.Name, RegexpDnsLabelLowercase)
	if err != nil {
		return GetInvalidValueError("Name", application.Name, err)
	}

	return nil
}
