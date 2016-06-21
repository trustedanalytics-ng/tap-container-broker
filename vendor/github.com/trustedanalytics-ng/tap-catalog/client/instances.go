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

func (c *TapCatalogApiConnector) ListApplicationInstances(applicationId string) ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s", c.Address, applications, applicationId, "instances"))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) ListServiceInstances(serviceId string) ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s", c.Address, services, serviceId, "instances"))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) ListApplicationsInstances() ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, applications, "instances"))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) ListServicesInstances() ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, services, "instances"))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) ListInstances() ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, instances))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) GetInstance(instanceId string) (models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, instances, instanceId))
	result := &models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) GetInstanceBindings(instanceId string) ([]models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s", c.Address, instances, instanceId, bindings))
	result := &[]models.Instance{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) UpdateInstance(instanceId string, patches []models.Patch) (models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, instances, instanceId))
	result := &models.Instance{}
	status, err := brokerHttp.PatchModel(connector, patches, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) AddServiceBrokerInstance(serviceId string, instance models.Instance) (models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/instances?isServiceBroker=true", c.Address, services, serviceId))
	result := &models.Instance{}
	status, err := brokerHttp.PostModel(connector, instance, http.StatusCreated, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) AddServiceInstance(serviceId string, instance models.Instance) (models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/instances?isServiceBroker=false", c.Address, services, serviceId))
	result := &models.Instance{}
	status, err := brokerHttp.PostModel(connector, instance, http.StatusCreated, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) AddApplicationInstance(applicationId string, instance models.Instance) (models.Instance, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/instances", c.Address, applications, applicationId))
	result := &models.Instance{}
	status, err := brokerHttp.PostModel(connector, instance, http.StatusCreated, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) DeleteInstance(instanceId string) (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, instances, instanceId))
	status, err := brokerHttp.DeleteModel(connector, http.StatusNoContent)
	return status, err
}

func (c *TapCatalogApiConnector) WatchInstances(afterIndex uint64) (models.StateChange, int, error) {
	connector := c.getWatchApiConnector(fmt.Sprintf("%s/%s/%s?afterIndex=%d", c.Address, instances, nextState, afterIndex))
	result := &models.StateChange{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) WatchInstance(instanceId string, afterIndex uint64) (models.StateChange, int, error) {
	connector := c.getWatchApiConnector(fmt.Sprintf("%s/%s/%s/%s?afterIndex=%d", c.Address, instances, instanceId, nextState, afterIndex))
	result := &models.StateChange{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}
