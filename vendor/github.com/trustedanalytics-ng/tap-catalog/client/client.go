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
	"time"

	"github.com/trustedanalytics-ng/tap-catalog/models"
	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

type TapCatalogApi interface {
	AddApplication(application models.Application) (models.Application, int, error)
	AddImage(image models.Image) (models.Image, int, error)
	AddService(service models.Service) (models.Service, int, error)
	AddServiceInstance(serviceId string, instance models.Instance) (models.Instance, int, error)
	AddServiceBrokerInstance(serviceId string, instance models.Instance) (models.Instance, int, error)
	AddApplicationInstance(applicationId string, instance models.Instance) (models.Instance, int, error)
	AddTemplate(template models.Template) (models.Template, int, error)
	GetApplication(applicationId string) (models.Application, int, error)
	GetCatalogHealth() (int, error)
	GetImage(imageId string) (models.Image, int, error)
	GetImageRefs(imageId string) (models.ImageRefsResponse, int, error)
	GetInstance(instanceId string) (models.Instance, int, error)
	GetInstanceBindings(instanceId string) ([]models.Instance, int, error)
	GetServicePlan(serviceId, planId string) (models.ServicePlan, int, error)
	GetService(serviceId string) (models.Service, int, error)
	GetServices() ([]models.Service, int, error)
	GetLatestIndex() (models.Index, int, error)
	ListApplications() ([]models.Application, int, error)
	ListApplicationsInstances() ([]models.Instance, int, error)
	ListInstances() ([]models.Instance, int, error)
	ListImages() ([]models.Image, int, error)
	ListServicesInstances() ([]models.Instance, int, error)
	ListApplicationInstances(applicationId string) ([]models.Instance, int, error)
	ListServiceInstances(serviceId string) ([]models.Instance, int, error)
	UpdateApplication(applicationId string, patches []models.Patch) (models.Application, int, error)
	UpdateImage(imageId string, patches []models.Patch) (models.Image, int, error)
	UpdateInstance(instanceId string, patches []models.Patch) (models.Instance, int, error)
	UpdatePlan(serviceId, planId string, patches []models.Patch) (models.ServicePlan, int, error)
	UpdateService(serviceId string, patches []models.Patch) (models.Service, int, error)
	UpdateTemplate(templateId string, patches []models.Patch) (models.Template, int, error)
	DeleteApplication(applicationId string) (int, error)
	DeleteService(serviceId string) (int, error)
	DeleteServicePlan(serviceId, planId string) (int, error)
	DeleteImage(imageId string) (int, error)
	DeleteInstance(instanceId string) (int, error)
	WatchInstances(afterIndex uint64) (models.StateChange, int, error)
	WatchInstance(instanceId string, afterIndex uint64) (models.StateChange, int, error)
	WatchImages(afterIndex uint64) (models.StateChange, int, error)
	WatchImage(imageId string, afterIndex uint64) (models.StateChange, int, error)
	CheckStateStability() (models.StateStability, int, error)
}

type TapCatalogApiConnector struct {
	Address     string
	Username    string
	Password    string
	Client      *http.Client
	WatchClient *http.Client
}

const (
	apiPrefix    = "api/"
	apiVersion   = "v1"
	instances    = apiPrefix + apiVersion + "/instances"
	services     = apiPrefix + apiVersion + "/services"
	applications = apiPrefix + apiVersion + "/applications"
	templates    = apiPrefix + apiVersion + "/templates"
	images       = apiPrefix + apiVersion + "/images"
	latestIndex  = apiPrefix + apiVersion + "/latest-index"
	stableState  = apiPrefix + apiVersion + "/stable-state"
	checkRefs    = "check-refs"
	nextState    = "next-state"
	healthz      = "healthz"
	bindings     = "bindings"
	plans        = "plans"

	maxIdleConnectionPerHost = 100
	watchClientTimeout       = time.Duration(30 * time.Minute)
)

func NewTapCatalogApiWithBasicAuth(address, username, password string) (TapCatalogApi, error) {
	client, _, err := brokerHttp.GetHttpClientWithCustomConnectionLimit(maxIdleConnectionPerHost)
	if err != nil {
		return nil, err
	}

	watchClient, _, err := brokerHttp.GetHttpClientWithCustomConnectionLimit(maxIdleConnectionPerHost)
	if err != nil {
		return nil, err
	}
	watchClient.Timeout = watchClientTimeout

	return &TapCatalogApiConnector{
		Address:     address,
		Username:    username,
		Password:    password,
		Client:      client,
		WatchClient: watchClient}, nil
}

func (c *TapCatalogApiConnector) getApiConnector(url string) brokerHttp.ApiConnector {
	return brokerHttp.ApiConnector{
		BasicAuth: &brokerHttp.BasicAuth{User: c.Username, Password: c.Password},
		Client:    c.Client,
		Url:       url,
	}
}

func (c *TapCatalogApiConnector) getWatchApiConnector(url string) brokerHttp.ApiConnector {
	return brokerHttp.ApiConnector{
		BasicAuth: &brokerHttp.BasicAuth{User: c.Username, Password: c.Password},
		Client:    c.WatchClient,
		Url:       url,
	}
}

func (c *TapCatalogApiConnector) GetLatestIndex() (models.Index, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, latestIndex))
	result := &models.Index{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}
