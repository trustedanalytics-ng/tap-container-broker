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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/trustedanalytics-ng/tap-ceph-broker/model"
	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

// CephBroker delivers an interface to access ceph-broker functionality to the client
type CephBroker interface {
	CreateRBD(device model.RBD) (int, error)
	DeleteRBD(name string) (int, error)

	ListLocks() ([]model.Lock, int, error)
	DeleteLock(lock model.Lock) (int, error)

	GetCephBrokerHealth() (int, error)
}

// CephBrokerConnector keeps data required to connect to the service
type CephBrokerConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

// NewCephBrokerBasicAuth returns initialized CephBrokerConnector structure for basic auth
func NewCephBrokerBasicAuth(address, username, password string) (*CephBrokerConnector, error) {
	client, _, err := brokerHttp.GetHttpClient()
	if err != nil {
		return nil, err
	}
	return &CephBrokerConnector{address, username, password, client}, nil
}

// NewCephBrokerCa returns initialized CephBrokerConnector structure for basic auth using certificate
func NewCephBrokerCa(address, username, password, certPemFile, keyPemFile, caPemFile string) (*CephBrokerConnector, error) {
	client, _, err := brokerHttp.GetHttpClientWithCertAndCaFromFile(certPemFile, keyPemFile, caPemFile)
	if err != nil {
		return nil, err
	}
	return &CephBrokerConnector{address, username, password, client}, nil
}

// CreateRBD calls api/v1/rbd POST method and verifies response status code
func (t *CephBrokerConnector) CreateRBD(device model.RBD) (int, error) {
	url := fmt.Sprintf("%s/api/v1/rbd", t.Address)

	b, err := json.Marshal(&device)
	if err != nil {
		return 400, err
	}

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestPOST(url, string(b), brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return status, err
	}
	if status != http.StatusOK {
		return status, errors.New("bad response status: " + strconv.Itoa(status))
	}
	return status, nil
}

// DeleteRBD calls api/v1/rbd DELETE method and verifies response status code
func (t *CephBrokerConnector) DeleteRBD(name string) (int, error) {
	url := fmt.Sprintf("%s/api/v1/rbd/%s", t.Address, name)

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestDELETE(url, "", brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return status, err
	}
	if status != http.StatusNoContent {
		return status, errors.New("bad response status: " + strconv.Itoa(status))
	}
	return status, nil
}

// GetCephBrokerHealth calls healthz and verifies response status code
func (t *CephBrokerConnector) GetCephBrokerHealth() (int, error) {
	url := fmt.Sprintf("%s/healthz", t.Address)

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestGET(url, brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if status != http.StatusOK {
		return http.StatusInternalServerError, fmt.Errorf("invalid health status: %v", err)
	}
	return http.StatusOK, nil
}

func (t *CephBrokerConnector) ListLocks() ([]model.Lock, int, error) {
	ret := []model.Lock{}

	url := fmt.Sprintf("%s/api/v1/lock", t.Address)

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, body, err := brokerHttp.RestGET(url, brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return ret, status, err
	}
	if status != http.StatusOK {
		return ret, status, errors.New("bad response status: " + strconv.Itoa(status))
	}

	if err := json.Unmarshal(body, &ret); err != nil {
		panic(err)
	}

	return ret, status, nil
}

func (t *CephBrokerConnector) DeleteLock(lock model.Lock) (int, error) {
	url := fmt.Sprintf("%s/api/v1/lock/%s/%s/%s", t.Address, lock.ImageName, lock.LockName, lock.Locker)
	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestDELETE(url, "", brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return status, err
	}
	if status != http.StatusNoContent {
		return status, errors.New("bad response status: " + strconv.Itoa(status))
	}
	return status, nil
}
