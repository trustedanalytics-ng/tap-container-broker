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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

type TapContainerBrokerApi interface {
	BindInstance(srcInstanceId, dstInstanceId string) (models.MessageResponse, int, error)
	UnbindInstance(srcInstanceId, dstInstanceId string) (models.MessageResponse, int, error)
	ExposeInstance(instanceId string, body models.ExposeRequest) ([]string, int, error)
	UnexposeInstance(instanceId string) (int, error)
	GetInstanceLogs(instanceId string) (map[string]string, int, error)
	GetContainerBrokerHealth() (int, error)
	GetInstanceHosts(instanceId string) ([]string, error)
	GetCredentials(instanceId string) ([]models.ContainerCredenials, int, error)
	GetConfigMap(configMapName string) (api.ConfigMap, int, error)
	GetSecret(secretName string) (api.Secret, int, error)
	GetVersions() ([]models.VersionsResponse, int, error)
}

type TapContainerBrokerApiConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

func NewTapContainerBrokerApiWithBasicAuth(address, username, password string) (*TapContainerBrokerApiConnector, error) {
	client, _, err := commonHttp.GetHttpClient()
	if err != nil {
		return nil, err
	}
	return &TapContainerBrokerApiConnector{Address: address, Username: username, Password: password, Client: client}, nil
}

func (c *TapContainerBrokerApiConnector) getApiConnector(url string) commonHttp.ApiConnector {
	return commonHttp.ApiConnector{
		BasicAuth: &commonHttp.BasicAuth{User: c.Username, Password: c.Password},
		Client:    c.Client,
		Url:       url,
	}
}

func (c *TapContainerBrokerApiConnector) BindInstance(srcInstanceId, dstInstanceId string) (models.MessageResponse, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/bind/%s/%s", c.Address, srcInstanceId, dstInstanceId))
	result := &models.MessageResponse{}
	status, err := commonHttp.PostModel(connector, "", http.StatusAccepted, result)
	return *result, status, err
}

func (c *TapContainerBrokerApiConnector) ExposeInstance(instanceId string, body models.ExposeRequest) ([]string, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/service/%s/expose", c.Address, instanceId))
	result := &[]string{}
	status, err := commonHttp.PostModel(connector, body, http.StatusOK, result)
	return *result, status, err
}

func (c *TapContainerBrokerApiConnector) UnexposeInstance(instanceId string) (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/service/%s/expose", c.Address, instanceId))
	return commonHttp.DeleteModel(connector, http.StatusNoContent)
}

func (c *TapContainerBrokerApiConnector) UnbindInstance(srcInstanceId, dstInstanceId string) (models.MessageResponse, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/unbind/%s/%s", c.Address, srcInstanceId, dstInstanceId))
	result := &models.MessageResponse{}
	status, err := commonHttp.PostModel(connector, "", http.StatusAccepted, result)
	return *result, status, err
}

func (c *TapContainerBrokerApiConnector) GetInstanceLogs(instanceId string) (map[string]string, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/service/%s/logs", c.Address, instanceId))
	status, response, err := commonHttp.RestGET(connector.Url, commonHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	logs := make(map[string]string)
	if status == http.StatusOK {
		err = json.Unmarshal(response, &logs)
	} else if status == http.StatusInternalServerError {
		msg := commonHttp.MessageResponse{}
		json.Unmarshal(response, &msg)
		return logs, status, errors.New(msg.Message)
	}
	return logs, status, err
}

func (c *TapContainerBrokerApiConnector) GetCredentials(instanceId string) (returnType []models.ContainerCredenials, status int, err error) {
	endpoint := fmt.Sprintf("/api/v1/service/%s/credentials", instanceId)
	status, err = c.callRestGETAndCheck(endpoint, &returnType)
	return
}

func (c *TapContainerBrokerApiConnector) GetContainerBrokerHealth() (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/healthz", c.Address))
	status, _, err := commonHttp.RestGET(connector.Url, commonHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	if status != http.StatusOK {
		err = errors.New("Invalid health status: " + string(status))
	}
	return status, err
}

func (c *TapContainerBrokerApiConnector) GetInstanceHosts(instanceId string) ([]string, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/service/%s/hosts", c.Address, instanceId))
	_, response, err := commonHttp.RestGET(connector.Url, commonHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	var hosts []string
	err = json.Unmarshal(response, &hosts)
	return hosts, err
}

func (c *TapContainerBrokerApiConnector) GetConfigMap(configMapName string) (returnType api.ConfigMap, status int, err error) {
	endpoint := fmt.Sprintf("/api/v1/configmap/%s", configMapName)
	status, err = c.callRestGETAndCheck(endpoint, &returnType)
	return
}

func (c *TapContainerBrokerApiConnector) GetSecret(secretName string) (returnType api.Secret, status int, err error) {
	endpoint := fmt.Sprintf("/api/v1/secret/%s", secretName)
	status, err = c.callRestGETAndCheck(endpoint, &returnType)
	return
}

func (c *TapContainerBrokerApiConnector) GetVersions() (returnType []models.VersionsResponse, status int, err error) {
	endpoint := "/api/v1/deployment/core/versions"
	status, err = c.callRestGETAndCheck(endpoint, &returnType)
	return
}

func (c *TapContainerBrokerApiConnector) callRestGETAndCheck(endpoint string, returnType interface{}) (int, error) {
	connector := c.getApiConnector(c.Address + endpoint)
	status, response, err := commonHttp.RestGET(connector.Url, commonHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	switch status {
	case http.StatusOK:
		err = json.Unmarshal(response, &returnType)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return status, nil
	case http.StatusInternalServerError:
		msg := commonHttp.MessageResponse{}
		err = json.Unmarshal(response, &msg)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return status, errors.New(msg.Message)
	default:
		return status, nil
	}
}
