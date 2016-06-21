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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestExpose(t *testing.T) {
	Convey("Test Expose endpoint", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)

		ports := []int32{1000, 2000}
		body := models.ExposeRequest{
			Hostname: "test-name",
			Ports:    ports,
		}

		instance, service, template := getSampleInstanceServiceAndTemplate()
		services := []api.Service{
			k8s.MakeServiceForPorts(goodInstanceID, ports),
		}

		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		Convey("Given existing instanceId from service-broker should register hosts with status code OK", func() {
			data.Template.Body[0].Type = templateModels.ComponentTypeBroker
			endpoint := k8s.MakeEndpointForPorts(goodInstanceID, body.Ip, body.Ports)
			hosts := make([]string, len(body.Ports))
			hostPatches := make([]catalogModels.Patch, 1) // we make always one patch with hosts list inside

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().CreateEndpoint(endpoint).Return(nil),
				mocker.kubernetesAPI.EXPECT().CreateService(services[0]).Return(nil),
				mocker.kubernetesAPI.EXPECT().CreateIngress(gomock.Any()).Do(func(ingress extensions.Ingress) {
					tmpHosts := k8s.FetchHostsFromIngress(ingress)
					patch, _ := builder.MakePatch("Metadata", catalogModels.Metadata{Id: "urls", Value: strings.Join(tmpHosts, ",")}, catalogModels.OperationAdd)
					Convey("getHostsFromIngress should return expected number of hosts", func() {
						So(tmpHosts, ShouldHaveLength, len(body.Ports))
					})
					// use copy instead append to keep the reference to the original slice - required for the next calls
					copy(hosts, tmpHosts)
					copy(hostPatches, []catalogModels.Patch{patch})
				}).Return(nil),
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return(hosts, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, hostPatches).Return(catalogModels.Instance{}, 200, nil),
			)

			responseHosts, err := manager.Expose(goodInstanceID, body)
			So(err, ShouldBeNil)
			So(responseHosts, ShouldHaveLength, len(body.Ports))
			So(responseHosts, ShouldResemble, hosts)
		})

		Convey("Given existing service instanceId should register hosts with status code OK", func() {
			body.Ports = []int32{1000}
			hosts := make([]string, len(body.Ports))
			hostPatches := make([]catalogModels.Patch, 1) // we make always one patch with hosts list inside

			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().GetServiceByInstanceId(goodInstanceID).Return(services, nil),
				mocker.kubernetesAPI.EXPECT().CreateIngress(gomock.Any()).Do(func(ingress extensions.Ingress) {
					tmpHosts := k8s.FetchHostsFromIngress(ingress)
					patch, _ := builder.MakePatch("Metadata", catalogModels.Metadata{Id: "urls", Value: strings.Join(tmpHosts, ",")}, catalogModels.OperationAdd)
					Convey("getHostsFromIngress should return expected number of hosts", func() {
						So(tmpHosts, ShouldHaveLength, len(body.Ports))
					})
					// use copy instead append to keep the reference to the original slice - required for the next calls
					copy(hosts, tmpHosts)
					copy(hostPatches, []catalogModels.Patch{patch})
				}).Return(nil),
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return(hosts, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, hostPatches).Return(catalogModels.Instance{}, 200, nil),
			)

			responseHosts, err := manager.Expose(goodInstanceID, body)
			So(err, ShouldBeNil)
			So(responseHosts, ShouldHaveLength, len(body.Ports))
			So(responseHosts, ShouldResemble, hosts)
		})

		Convey("Should return error if there is not assigned services", func() {
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().GetServiceByInstanceId(goodInstanceID).Return([]api.Service{}, nil),
			)

			_, err := manager.Expose(goodInstanceID, body)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "can not find service with instanceID")
		})

		Convey("Should return error if there is not specific port in service's ports", func() {
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().GetServiceByInstanceId(goodInstanceID).Return(services, nil),
			)

			var nonServicePort int32 = 3000
			body.Ports = []int32{nonServicePort}

			_, err := manager.Expose(goodInstanceID, body)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "does not exist in Service!")
		})

		Convey("Given existing instanceId from service-broker should return error when request ports list is empty", func() {
			data.Template.Body[0].Type = templateModels.ComponentTypeBroker

			mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil)

			body.Ports = []int32{}

			_, err := manager.Expose(goodInstanceID, body)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ports list can not be empty!")
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestHide(t *testing.T) {
	Convey("Test Hide endpoint", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)

		emptyHosts := []string{}
		instance, service, template := getSampleInstanceServiceAndTemplate()
		data := getCatalogEntityWithK8sTemplateForService(instance, service, template)

		patch, _ := builder.MakePatch("Metadata", catalogModels.Metadata{Id: "urls", Value: ""}, catalogModels.OperationAdd)
		hostPatches := []catalogModels.Patch{patch}

		Convey("Given existing service instanceId should unregister hosts with status code OK", func() {
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().DeleteIngress(goodInstanceID).Return(nil),
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return(emptyHosts, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, hostPatches).Return(catalogModels.Instance{}, 200, nil),
			)

			_, err := manager.Hide(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Convey("Given existing instanceId from service-broker should unregister hosts with status code OK", func() {
			data.Template.Body[0].Type = templateModels.ComponentTypeBroker
			gomock.InOrder(
				mocker.service.EXPECT().GetCatalogEntityWithK8sTemplate(gomock.Any()).Return(data, models.DeployResponseStatus(""), http.StatusOK, nil),
				mocker.kubernetesAPI.EXPECT().DeleteIngress(goodInstanceID).Return(nil),
				mocker.kubernetesAPI.EXPECT().DeleteService(commonHttp.UuidToShortDnsName(goodInstanceID)).Return(nil),
				mocker.kubernetesAPI.EXPECT().GetIngressHosts(goodInstanceID).Return(emptyHosts, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, hostPatches).Return(catalogModels.Instance{}, 200, nil),
			)

			_, err := manager.Hide(goodInstanceID)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
