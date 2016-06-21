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

package engine

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

func TestGetCredentials(t *testing.T) {
	Convey("Test GetAllPodsEnvsByInstanceId", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)

		instance, service, template := getSampleInstanceServiceAndTemplate()
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Given existing instanceId it should return envs with status code OK", func() {
			testCredentials := getTestCredentials()
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().GetDeploymentsEnvsByInstanceId(goodInstanceID).Return(testCredentials, nil),
			)

			responseEnvs, err := manager.GetCredentials(goodInstanceID)

			metadataCredentials, err := convertInstanceMetadataToContainerCredenials(instance)
			testCredentialsWithMetadata := append(testCredentials, metadataCredentials)

			So(err, ShouldBeNil)
			So(responseEnvs, ShouldResemble, testCredentialsWithMetadata)
		})

		Convey("Given not existing instanceId it should return status code NotFound", func() {
			mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusNotFound, errors.New(k8s.NotFound))

			var err error
			_, err = manager.GetCredentials(wrongInstanceID)
			So(err, ShouldNotBeNil)
		})

		Convey("Inncorrect Instance state case validation error", func() {
			data.Instance.State = catalogModels.InstanceStateFailure

			mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil)

			_, err := manager.GetCredentials(goodInstanceID)
			So(err.Error(), ShouldEqual, "instance state is inappropriate: "+string(catalogModels.InstanceStateFailure))
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
