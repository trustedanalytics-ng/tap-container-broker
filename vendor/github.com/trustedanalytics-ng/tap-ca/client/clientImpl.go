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

	caModels "github.com/trustedanalytics-ng/tap-ca/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *TapCaApiConnector) GetCa() (caModels.CaResponse, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, ca))
	result := caModels.CaResponse{}
	_, err := commonHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}

func (c *TapCaApiConnector) GetCertKey(domain string) (caModels.CertKeyResponse, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, certKey, domain))
	result := caModels.CertKeyResponse{}
	_, err := commonHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}

func (c *TapCaApiConnector) GetCaBundle() (caModels.CaBundleResponse, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, caBundle))
	result := caModels.CaBundleResponse{}
	_, err := commonHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}
