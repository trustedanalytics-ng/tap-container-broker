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

package api

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gocraft/web"
	"github.com/golang/mock/gomock"
	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/mocks"
)

const (
	urlPrefix             = "/api/v1"
	secretNameWildcard    = ":secretName"
	urlGetSecret          = urlPrefix + "/secret/" + secretNameWildcard
	instanceIDWildcard    = ":instanceId"
	urlGetEnvs            = urlPrefix + "/service/" + instanceIDWildcard + "/credentials"
	urlGetVersion         = urlPrefix + "/deployment/core/versions"
	configMapNameWildcard = ":configMapName"
	urlGetConfigMap       = urlPrefix + "/configmap/" + configMapNameWildcard
	urlExpose             = urlPrefix + "/service/" + instanceIDWildcard + "/expose"
)

func prepareMocksAndRouter(t *testing.T) (*mocks.MockServiceEngine, *web.Router, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	serviceEngineMock := mocks.NewMockServiceEngine(mockCtrl)

	ServiceEngine = serviceEngineMock

	router := web.New(Context{})
	router.Get(urlGetSecret, (*Context).GetSecret)
	router.Get(urlGetConfigMap, (*Context).GetConfigMap)
	router.Get(urlGetEnvs, (*Context).GetCredentials)
	router.Get(urlGetVersion, (*Context).GetVersions)
	router.Post(urlExpose, (*Context).Expose)
	router.Delete(urlExpose, (*Context).Hide)

	return serviceEngineMock, router, mockCtrl
}

func getTestSecret(secretName, secretData string) api.Secret {
	secret := api.Secret{}
	secret.Name = secretName
	data := make(map[string][]byte)
	data[secretName] = []byte(secretData)
	secret.Data = data
	return secret
}

func getTestConfigMap() api.ConfigMap {
	const dataNumber = 3
	result := api.ConfigMap{}
	result.Data = make(map[string]string)
	for i := 0; i < dataNumber; i++ {
		result.Data[strconv.Itoa(i)] = fmt.Sprintf("sample data no. %d", i)
	}
	return result
}
