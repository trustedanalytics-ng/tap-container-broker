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
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestScaleServices(t *testing.T) {
	setProcessorParams(time.Millisecond, 2, 2)

	Convey("Test Scale method for service instance", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, service, template := getSampleInstanceServiceAndTemplate()
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Should start and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should stop and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStopReq

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopReq, catalogModels.InstanceStateStopping).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStopping).Return(nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should restart and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateReconfiguration

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateReconfiguration, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestScaleApplication(t *testing.T) {
	setProcessorParams(time.Millisecond, 2, 2)
	Convey("Test Scale method for application instance", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, _, template := getSampleInstanceServiceAndTemplate()
		instance.Type = catalogModels.InstanceTypeApplication

		app, _ := getSampleApplicationAndImage()
		data := getCatalogEntityWithK8sTemplateForApp(instance, app, template)

		Convey("Should start and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should stop and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStopReq

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopReq, catalogModels.InstanceStateStopping).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStopping).Return(nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should restart and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateReconfiguration

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateReconfiguration, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should scale to 0 and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateReconfiguration
			data.Application.Replication = 0

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateReconfiguration, catalogModels.InstanceStateStopping).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStopping).Return(nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should return error if replicas equal 0 in start case", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq
			data.Application.Replication = 0

			mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "can not start application with 0 replicas!")
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestScaleServiceBrokerInstance(t *testing.T) {
	Convey("Test Scale method for service-broker instance", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, service, template := getSampleServiceBrokerInstanceServiceAndTemplate()
		instance.Type = catalogModels.InstanceTypeServiceBroker
		template.Body[0].Type = templateModels.ComponentTypeBroker

		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		updatedInstance := instance

		Convey("Should start and update Offering state to READY and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq
			updatedInstance.State = catalogModels.InstanceStateRunning

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(updatedInstance, http.StatusOK, nil),

				// updateServiceBrokerOfferings
				mocker.catalogAPI.EXPECT().GetService(offeringId).Return(service, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateServiceStatus(offeringId, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateReady).Return(nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should stop and update Offering state to OFFLINE and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStopReq
			updatedInstance.State = catalogModels.InstanceStateStopped

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopReq, catalogModels.InstanceStateStopping).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStopping).Return(nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(updatedInstance, http.StatusOK, nil),

				// updateServiceBrokerOfferings
				mocker.catalogAPI.EXPECT().GetService(offeringId).Return(service, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateServiceStatus(offeringId, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateOffline).Return(nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should fail on start and update Offering state to OFFLINE and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq
			updatedInstance.State = catalogModels.InstanceStateFailure

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),

				mocker.service.EXPECT().ScaleOnKubernetes(goodInstanceID, data, catalogModels.InstanceStateStarting).Return(nil),
				mocker.service.EXPECT().MonitorInstanceDeployments(*data.Instance).Return(nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(updatedInstance, http.StatusOK, nil),

				// updateServiceBrokerOfferings
				mocker.catalogAPI.EXPECT().GetService(offeringId).Return(service, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateServiceStatus(offeringId, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateOffline).Return(nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestScaleServiceBrokerOfferingInstance(t *testing.T) {
	Convey("Test Scale method for service-broker instances", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, service, template := getSampleServiceBrokerInstanceServiceAndTemplate()
		template.Body[0].Type = templateModels.ComponentTypeBroker
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Should only change state in StartReq and return nil", func() {
			data.Instance.State = catalogModels.InstanceStateStartReq

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting).Return(http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateStarting, catalogModels.InstanceStateRunning).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Should return error if service instance state has inappropriate value", func() {
			instance.State = catalogModels.InstanceStateRunning

			mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil)

			err := manager.ScaleInstance(goodInstanceID)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
