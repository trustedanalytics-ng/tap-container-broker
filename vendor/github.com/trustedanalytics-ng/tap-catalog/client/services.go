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

func (c *TapCatalogApiConnector) GetServices() ([]models.Service, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, services))
	result := []models.Service{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) GetService(serviceId string) (models.Service, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, services, serviceId))
	result := models.Service{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) UpdateService(serviceId string, patches []models.Patch) (models.Service, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, services, serviceId))
	result := models.Service{}
	status, err := brokerHttp.PatchModel(connector, patches, http.StatusOK, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) UpdatePlan(serviceId, planId string, patches []models.Patch) (models.ServicePlan, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/plans/%s", c.Address, services, serviceId, planId))
	result := models.ServicePlan{}
	status, err := brokerHttp.PatchModel(connector, patches, http.StatusOK, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) AddService(service models.Service) (models.Service, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, services))
	result := models.Service{}
	status, err := brokerHttp.PostModel(connector, service, http.StatusCreated, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) GetServicePlan(serviceId, planId string) (models.ServicePlan, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s/%s", c.Address, services, serviceId, plans, planId))
	result := models.ServicePlan{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, status, err
}

func (c *TapCatalogApiConnector) DeleteService(serviceId string) (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, services, serviceId))
	status, err := brokerHttp.DeleteModel(connector, http.StatusNoContent)
	return status, err
}

func (c *TapCatalogApiConnector) DeleteServicePlan(serviceId, planId string) (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s/%s", c.Address, services, serviceId, plans, planId))
	status, err := brokerHttp.DeleteModel(connector, http.StatusNoContent)
	return status, err
}
