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
package model

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

type HookType string

const (
	HookTypeDeployment  HookType = "deployment"
	HookTypeProvision   HookType = "provision"
	HookTypeDeprovision HookType = "deprovision"
	HookTypeBind        HookType = "bind"
	HookTypeUnbind      HookType = "unbind"
	HookTypeRemoval     HookType = "removal"
)

const (
	RAW_TEMPLATE_ID_FIELD = "id"

	PLAN_NAMES_ANNOTATION = "plan_names"
	EMPTY_PLAN_NAME       = ""
)

type RawTemplate map[string]interface{}

type Template struct {
	Id    string                `json:"id"`
	Body  []KubernetesComponent `json:"body"`
	Hooks map[HookType]*api.Pod `json:"hooks"`
}

type ComponentType string

const (
	ComponentTypeBroker   ComponentType = "broker"
	ComponentTypeInstance ComponentType = "instance"
	ComponentTypeBoth     ComponentType = "both"
)

type KubernetesComponent struct {
	Type                   ComponentType                `json:"componentType"`
	PersistentVolumeClaims []*api.PersistentVolumeClaim `json:"persistentVolumeClaims"`
	Deployments            []*extensions.Deployment     `json:"deployments"`
	Ingresses              []*extensions.Ingress        `json:"ingresses"`
	Services               []*api.Service               `json:"services"`
	ServiceAccounts        []*api.ServiceAccount        `json:"serviceAccounts"`
	Secrets                []*api.Secret                `json:"secrets"`
	ConfigMaps             []*api.ConfigMap             `json:"configMaps"`
}

func IsServiceBrokerTemplate(template Template) bool {
	for _, component := range template.Body {
		if component.Type == ComponentTypeBroker || component.Type == ComponentTypeBoth {
			return true
		}
	}
	return false
}
