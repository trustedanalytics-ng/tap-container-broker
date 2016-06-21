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
	"net/http"

	caModels "github.com/trustedanalytics-ng/tap-ca/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

type TapCaApi interface {
	GetCa() (caModels.CaResponse, error)
	GetCertKey(domain string) (caModels.CertKeyResponse, error)
	GetCaBundle() (caModels.CaBundleResponse, error)
}

type TapCaApiConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

const (
	apiPrefix  = "api/"
	apiVersion = "v1"
	ca         = apiPrefix + apiVersion + "/ca"
	certKey    = apiPrefix + apiVersion + "/certkey"
	caBundle   = apiPrefix + apiVersion + "/ca-bundle"
)

func NewTapCaApiConnector(address, username, password string) (*TapCaApiConnector, error) {
	client, _, err := commonHttp.GetHttpClient()
	if err != nil {
		return nil, err
	}
	return &TapCaApiConnector{address, username, password, client}, nil
}

func (c *TapCaApiConnector) getApiConnector(url string) commonHttp.ApiConnector {
	return commonHttp.ApiConnector{
		BasicAuth: &commonHttp.BasicAuth{User: c.Username, Password: c.Password},
		Client:    c.Client,
		Url:       url,
	}
}
