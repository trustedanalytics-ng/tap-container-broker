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

package processor

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
)

func copyEnvs(envs []api.EnvVar) []api.EnvVar {
	res := []api.EnvVar{}
	for _, env := range envs {
		res = append(res, env)
	}
	return res
}

func TestAddBoundServicesRegistry(t *testing.T) {
	testCases := []struct {
		dstEnvs      []api.EnvVar
		instanceName string
		offeringName string
		outEnvs      []api.EnvVar
	}{
		{
			dstEnvs:      []api.EnvVar{},
			instanceName: "mydb",
			offeringName: "mysql56",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_MYSQL56", Value: "mydb"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "someenv"}},
			instanceName: "mydb",
			offeringName: "mysql56",
			outEnvs:      []api.EnvVar{{Name: "someenv"}, {Name: "SERVICES_BOUND_MYSQL56", Value: "mydb"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service,mynats"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}},
			instanceName: "logservice",
			offeringName: "logstash",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}, {Name: "SERVICES_BOUND_LOGSTASH", Value: "logservice"}},
		},
		{
			dstEnvs:      []api.EnvVar{},
			instanceName: "my-postgres",
			offeringName: "postgres-93",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_POSTGRES_93", Value: "my-postgres"}},
		},
	}

	Convey("For set of test cases addBoundServiceRegistry should return proper response", t, func() {
		for _, tc := range testCases {
			Convey(fmt.Sprintf("For input envs=%v, instanceName=%s, serviceName=%s, the response should be %v", tc.dstEnvs, tc.instanceName, tc.offeringName, tc.outEnvs), func() {
				response := copyEnvs(tc.dstEnvs)
				response = addToBoundServicesRegistry(response, tc.offeringName, tc.instanceName)

				Convey("Response should be proper", func() {
					So(response, ShouldResemble, tc.outEnvs)
				})
			})
		}
	})
}

func TestRemoveBoundServicesRegistry(t *testing.T) {
	testCases := []struct {
		dstEnvs      []api.EnvVar
		instanceName string
		offeringName string
		outEnvs      []api.EnvVar
	}{
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_MYSQL56", Value: "mydb"}},
			instanceName: "mydb",
			offeringName: "mysql56",
			outEnvs:      []api.EnvVar{},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "someenv"}, {Name: "SERVICES_BOUND_MYSQL56", Value: "mydb"}},
			instanceName: "mydb",
			offeringName: "mysql56",
			outEnvs:      []api.EnvVar{{Name: "someenv"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service,mynats"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats,service"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service1,mynats,service2"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "service1,service2"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}, {Name: "SERVICES_BOUND_LOGSTASH", Value: "logservice"}},
			instanceName: "mynats",
			offeringName: "nats",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_LOGSTASH", Value: "logservice"}},
		},
		{
			dstEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}, {Name: "SERVICES_BOUND_POSTGRES_93", Value: "my-postgres"}},
			instanceName: "my-postgres",
			offeringName: "postgres-93",
			outEnvs:      []api.EnvVar{{Name: "SERVICES_BOUND_NATS", Value: "mynats"}},
		},
	}

	Convey("For set of test cases removeBoundServiceRegistry should return proper response", t, func() {
		for _, tc := range testCases {
			Convey(fmt.Sprintf("For input envs=%v, instanceName=%s, serviceName=%s, the response should be %v", tc.dstEnvs, tc.instanceName, tc.offeringName, tc.outEnvs), func() {
				response := copyEnvs(tc.dstEnvs)
				response = removeFromBoundServicesRegistry(response, tc.offeringName, tc.instanceName)

				Convey("Response should be proper", func() {
					So(response, ShouldResemble, tc.outEnvs)
				})
			})
		}
	})
}

func TestServiceProcessorManager_LinkInstances(t *testing.T) {
	Convey("Test LinkInstances", t, func() {
		setProcessorParams(time.Millisecond, 2, 2)
		mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)

		srcData := getCatalogEntityWithK8sTemplateForService()
		dstData := getCatalogEntityWithK8sTemplateForService()

		deployment := dstData.Template.Body[0].Deployments[0]

		Convey("Should link instances and return nil error response", func() {
			gomock.InOrder(
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, patchesInstanceStateRunningToUnavailable).Return(*dstData.Instance, http.StatusOK, nil),

				// setUnavailableStateAndStopInstance
				mocker.kubernetesAPI.EXPECT().GetDeployment(deployment.Name).Return(deployment, nil),
				mocker.kubernetesAPI.EXPECT().ScaleDeploymentAndWait(deployment.Name, goodInstanceID, 0).Return(nil),

				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(*dstData.Instance, http.StatusOK, nil),

				mocker.kubernetesAPI.EXPECT().UpdateDeployment(goodInstanceID, gomock.Any()).Return(nil),

				// handleBindingOperationResult
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, patchesInstanceStateUnavailableToStopped).Return(*dstData.Instance, http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, patchesInstanceStateStoppedToStartReq).Return(*dstData.Instance, http.StatusOK, nil),
			)

			err := manager.LinkInstances(srcData, dstData, true)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
