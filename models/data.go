/**
 * Copyright (c) 2017 Intel Corporation
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
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	templateRepositoryModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

type CatalogEntityWithK8sTemplate struct {
	Template    *templateRepositoryModel.Template
	Instance    *catalogModels.Instance
	Application *catalogModels.Application
	Service     *catalogModels.Service
	ServicePlan *catalogModels.ServicePlan
}

type CatalogEntityWithK8sTemplateRequest struct {
	InstanceId             string
	BoundInstanceId        string
	ApplicationSSLCertHash string
}
