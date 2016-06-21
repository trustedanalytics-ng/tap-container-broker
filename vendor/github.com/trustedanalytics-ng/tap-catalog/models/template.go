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

type Template struct {
	Id         string        `json:"templateId"`
	State      TemplateState `json:"state"`
	AuditTrail AuditTrail    `json:"auditTrail"`
}

type TemplateState string

const (
	TemplateStateInProgress  TemplateState = "IN_PROGRESS"
	TemplateStateReady       TemplateState = "READY"
	TemplateStateUnavailable TemplateState = "UNAVAILABLE"
)

func (template *Template) ValidateTemplateStructCreate() error {
	if template.Id != "" {
		return GetIdFieldHasToBeEmptyError()
	}
	return nil
}
