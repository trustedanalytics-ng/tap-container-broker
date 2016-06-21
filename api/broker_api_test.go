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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocraft/web"
	. "github.com/smartystreets/goconvey/convey"
	k8sAPI "k8s.io/kubernetes/pkg/api"

	client "github.com/trustedanalytics-ng/tap-container-broker/client"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
)

func TestGetSecret(t *testing.T) {
	Convey("Test GetSecret", t, func() {
		serviceMock, router, mockCtrl := prepareMocksAndRouter(t)
		cBroker := getContainerBroker(router, t)

		Convey("Given existing secret it should return secret with status code OK", func() {
			testSecretName := "testSecret"
			testSecretData := "dGVzdA=="
			testSecret := getTestSecret(testSecretName, testSecretData)
			serviceMock.EXPECT().GetSecret(testSecretName).Return(&testSecret, nil)

			responseSecret, status, err := cBroker.GetSecret(testSecretName)
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(responseSecret, ShouldResemble, testSecret)
		})

		Convey("Given not existing secret it should return status code NotFound", func() {
			wrongSecretName := "not-existing secret"
			serviceMock.EXPECT().GetSecret(wrongSecretName).Return(&k8sAPI.Secret{}, errors.New(k8s.NotFound))

			_, status, err := cBroker.GetSecret(wrongSecretName)
			So(err, ShouldBeNil)
			So(status, ShouldResemble, http.StatusNotFound)
		})

		Convey("In case of unknown error when getting secret from k8s it should return status code InternalServerError", func() {
			testSecretName := "testSecret"
			errorMessage := "some error"
			serviceMock.EXPECT().GetSecret(testSecretName).Return(&k8sAPI.Secret{}, errors.New(errorMessage))

			_, status, err := cBroker.GetSecret(testSecretName)
			So(err.Error(), ShouldEqual, errorMessage)
			So(status, ShouldEqual, http.StatusInternalServerError)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestGetConfigMap(t *testing.T) {
	Convey("Test GetConfigMap", t, func() {
		serviceMock, router, mockCtrl := prepareMocksAndRouter(t)
		cBroker := getContainerBroker(router, t)

		Convey("Given existing ConfigMap it should return configMap with status code OK", func() {
			testConfigMapName := "testConfigMap"
			testConfigMap := getTestConfigMap()
			serviceMock.EXPECT().GetConfigMap(testConfigMapName).Return(&testConfigMap, nil)

			responseConfigMap, status, err := cBroker.GetConfigMap(testConfigMapName)
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(responseConfigMap, ShouldResemble, testConfigMap)
		})

		Convey("Given not existing ConfigMap it should return status code NotFound", func() {
			wrongConfigMapName := "not-existing ConfigMap"
			serviceMock.EXPECT().GetConfigMap(wrongConfigMapName).Return(&k8sAPI.ConfigMap{}, errors.New(k8s.NotFound))

			_, status, err := cBroker.GetConfigMap(wrongConfigMapName)
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusNotFound)
		})

		Convey("In case of unknown error when getting ConfigMap from k8s it should return status code InternalServerError", func() {
			testConfigMapName := "testConfigMap"
			errorMessage := "some error"
			serviceMock.EXPECT().GetConfigMap(testConfigMapName).Return(&k8sAPI.ConfigMap{}, errors.New(errorMessage))

			_, status, err := cBroker.GetConfigMap(testConfigMapName)
			So(err.Error(), ShouldEqual, errorMessage)
			So(status, ShouldEqual, http.StatusInternalServerError)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func getContainerBroker(router *web.Router, t *testing.T) *client.TapContainerBrokerApiConnector {
	testServer := httptest.NewServer(router)
	cBroker, err := client.NewTapContainerBrokerApiWithBasicAuth(testServer.URL, "user", "password")
	if err != nil {
		t.Fatal("Container broker error: ", err)
	}
	return cBroker
}
