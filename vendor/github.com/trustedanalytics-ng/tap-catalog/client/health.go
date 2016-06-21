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
	"errors"
	"fmt"
	"net/http"

	"strconv"

	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *TapCatalogApiConnector) GetCatalogHealth() (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, healthz))
	status, _, err := brokerHttp.RestGET(connector.Url, brokerHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	if status != http.StatusOK {
		err = errors.New("Invalid health status: " + strconv.Itoa(status))
	}
	return status, err
}
