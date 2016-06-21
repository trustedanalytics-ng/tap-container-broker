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

package engine

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestDeleteKubernetesInstance(t *testing.T) {
	setProcessorParams(time.Millisecond, 2, 2)

	Convey("Test DeleteKubernetesInstance method for service instance", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, service, template := getSampleInstanceServiceAndTemplate()
		instance.State = catalogModels.InstanceStateRequested
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Should delete service instance and return nil", func() {
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, "", catalogModels.InstanceStateDestroyReq, catalogModels.InstanceStateDestroying).Return(http.StatusOK, nil),

				// removeKubernetesComponentsAndCatalogData
				mocker.kubernetesAPI.EXPECT().DeleteAllByInstanceId(goodInstanceID).Return(nil),
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeDeprovision, data).Return(nil, data.Instance, nil),
				mocker.catalogAPI.EXPECT().DeleteInstance(goodInstanceID).Return(http.StatusOK, nil),
			)

			err := manager.DeleteKubernetesInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should delete application instance and return nil", func() {
			data.Instance.Type = catalogModels.InstanceTypeApplication

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, "", catalogModels.InstanceStateDestroyReq, catalogModels.InstanceStateDestroying).Return(http.StatusOK, nil),

				// removeKubernetesComponentsAndCatalogData
				mocker.kubernetesAPI.EXPECT().DeleteAllByInstanceId(goodInstanceID).Return(nil),
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeDeprovision, data).Return(nil, data.Instance, nil),
				mocker.catalogAPI.EXPECT().DeleteInstance(goodInstanceID).Return(http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().DeleteApplication(offeringId).Return(http.StatusOK, nil),
			)

			err := manager.DeleteKubernetesInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
