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
	"k8s.io/kubernetes/pkg/api"

	caModels "github.com/trustedanalytics-ng/tap-ca/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestCreateKubernetesInstance(t *testing.T) {
	setProcessorParams(time.Millisecond, 2, 2)

	Convey("Test CreateKubernetesInstance method for service instance", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		instance, service, template := getSampleInstanceServiceAndTemplate()
		instance.State = catalogModels.InstanceStateRequested
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Should create service instance and start and return nil", func() {
			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateRequested, catalogModels.InstanceStateDeploying).Return(http.StatusOK, nil),
				mocker.service.EXPECT().GetInstanceCatalogEntityWithK8STemplate(gomock.Any(), instance).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().WaitForInstanceDependencies(data).Return(nil),

				// runHooksAndSetReplicasAndCreateComponents
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeProvision, data).Return(nil, data.Instance, nil),
				mocker.kubernetesAPI.EXPECT().FabricateComponents(goodInstanceID, false, gomock.Any(), data.Template.Body).Return(nil),
				mocker.service.EXPECT().UpdateStateAndGetInstance(goodInstanceID, gomock.Any(), catalogModels.InstanceStateDeploying, catalogModels.InstanceStateStopped).Return(instance, http.StatusOK, nil),

				// setHostnameInInstanceMetadata
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return([]string{}, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(instance, http.StatusOK, nil),

				// startAction
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, "", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.CreateKubernetesInstance(models.CreateInstanceRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
		})

		Convey("Should create service instance and not start and return nil", func() {
			template.Body[0].Deployments[0].Spec.Replicas = 0
			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateRequested, catalogModels.InstanceStateDeploying).Return(http.StatusOK, nil),
				mocker.service.EXPECT().GetInstanceCatalogEntityWithK8STemplate(gomock.Any(), instance).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().WaitForInstanceDependencies(data).Return(nil),

				// runHooksAndSetReplicasAndCreateComponents
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeProvision, data).Return(nil, data.Instance, nil),
				mocker.kubernetesAPI.EXPECT().FabricateComponents(goodInstanceID, false, gomock.Any(), data.Template.Body).Return(nil),
				mocker.service.EXPECT().UpdateStateAndGetInstance(goodInstanceID, gomock.Any(), catalogModels.InstanceStateDeploying, catalogModels.InstanceStateStopped).Return(instance, http.StatusOK, nil),

				// setHostnameInInstanceMetadata
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return([]string{}, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(instance, http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.CreateKubernetesInstance(models.CreateInstanceRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
		})

		Convey("Should create service instance with dependency and return nil", func() {
			bindingId := "bindingTestId"
			instance.Bindings = []catalogModels.InstanceBindings{{
				Id: bindingId,
			}}
			instanceToBound := getSampleInstance()
			instanceToBound.Id = bindingId

			srcData := getCatalogEntityWithK8sTemplateForService(instanceToBound, service, template)

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateRequested, catalogModels.InstanceStateDeploying).Return(http.StatusOK, nil),
				mocker.service.EXPECT().GetInstanceCatalogEntityWithK8STemplate(gomock.Any(), instance).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().WaitForInstanceDependencies(data).Return(nil),
			)

			instance.State = catalogModels.InstanceStateStopped

			gomock.InOrder(
				// runHooksAndSetReplicasAndCreateComponents
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeProvision, data).Return(nil, data.Instance, nil),
				mocker.kubernetesAPI.EXPECT().FabricateComponents(goodInstanceID, false, gomock.Any(), data.Template.Body).Return(nil),
				mocker.service.EXPECT().UpdateStateAndGetInstance(goodInstanceID, gomock.Any(), catalogModels.InstanceStateDeploying, catalogModels.InstanceStateStopped).Return(instance, http.StatusOK, nil),

				// ValidateAndBindDependencyInstance
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(srcData, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().LinkInstances(srcData, gomock.Any(), true).Return(nil),

				// setHostnameInInstanceMetadata
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return([]string{}, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(instance, http.StatusOK, nil),

				// startAction
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, "", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.CreateKubernetesInstance(models.CreateInstanceRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
		})

		Convey("Should create application instance and return nil", func() {
			data.Instance.Type = catalogModels.InstanceTypeApplication
			data.Template.Body[0].Services = append(template.Body[0].Services, &api.Service{})

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(*data.Instance, http.StatusOK, nil),
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, gomock.Any(), catalogModels.InstanceStateRequested, catalogModels.InstanceStateDeploying).Return(http.StatusOK, nil),
				mocker.caAPI.EXPECT().GetCa().Return(caModels.CaResponse{}, nil),
				mocker.service.EXPECT().GetInstanceCatalogEntityWithK8STemplate(gomock.Any(), *data.Instance).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.service.EXPECT().WaitForInstanceDependencies(data).Return(nil),

				// runHooksAndSetReplicasAndCreateComponents
				mocker.service.EXPECT().ProcessHook(templateModels.HookTypeProvision, data).Return(nil, data.Instance, nil),
				mocker.caAPI.EXPECT().GetCertKey(gomock.Any()).Return(caModels.CertKeyResponse{}, nil),
				mocker.kubernetesAPI.EXPECT().CreateSecret(gomock.Any()).Return(nil),
				mocker.kubernetesAPI.EXPECT().FabricateComponents(goodInstanceID, true, gomock.Any(), data.Template.Body).Return(nil),
				mocker.service.EXPECT().UpdateStateAndGetInstance(goodInstanceID, gomock.Any(), catalogModels.InstanceStateDeploying, catalogModels.InstanceStateStopped).Return(instance, http.StatusOK, nil),

				// setHostnameInInstanceMetadata
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return([]string{}, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(instance, http.StatusOK, nil),

				// startAction
				mocker.service.EXPECT().UpdateInstanceState(goodInstanceID, "", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq).Return(http.StatusOK, nil),

				// handleOperationOnInstanceError
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
			)

			err := manager.CreateKubernetesInstance(models.CreateInstanceRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
