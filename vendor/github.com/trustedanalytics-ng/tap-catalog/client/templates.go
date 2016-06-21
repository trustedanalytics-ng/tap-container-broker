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
package client

import (
	"fmt"
	"net/http"

	"github.com/trustedanalytics-ng/tap-catalog/models"
	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *TapCatalogApiConnector) AddTemplate(template models.Template) (models.Template, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, templates))
	result := &models.Template{}
	status, err := brokerHttp.PostModel(connector, template, http.StatusCreated, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) UpdateTemplate(templateId string, patches []models.Patch) (models.Template, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, templates, templateId))
	result := &models.Template{}
	status, err := brokerHttp.PatchModel(connector, patches, http.StatusOK, result)
	return *result, status, err
}
