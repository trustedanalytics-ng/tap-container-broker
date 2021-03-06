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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: engine/engine.go

package mocks

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/trustedanalytics-ng/tap-container-broker/models"
	api "k8s.io/kubernetes/pkg/api"
)

// Mock of ServiceEngine interface
type MockServiceEngine struct {
	ctrl     *gomock.Controller
	recorder *_MockServiceEngineRecorder
}

// Recorder for MockServiceEngine (not exported)
type _MockServiceEngineRecorder struct {
	mock *MockServiceEngine
}

func NewMockServiceEngine(ctrl *gomock.Controller) *MockServiceEngine {
	mock := &MockServiceEngine{ctrl: ctrl}
	mock.recorder = &_MockServiceEngineRecorder{mock}
	return mock
}

func (_m *MockServiceEngine) EXPECT() *_MockServiceEngineRecorder {
	return _m.recorder
}

func (_m *MockServiceEngine) CreateKubernetesInstance(createInstanceRequest models.CreateInstanceRequest) error {
	ret := _m.ctrl.Call(_m, "CreateKubernetesInstance", createInstanceRequest)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockServiceEngineRecorder) CreateKubernetesInstance(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateKubernetesInstance", arg0)
}

func (_m *MockServiceEngine) DeleteKubernetesInstance(instanceId string) error {
	ret := _m.ctrl.Call(_m, "DeleteKubernetesInstance", instanceId)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockServiceEngineRecorder) DeleteKubernetesInstance(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteKubernetesInstance", arg0)
}

func (_m *MockServiceEngine) Expose(instanceId string, request models.ExposeRequest) ([]string, error) {
	ret := _m.ctrl.Call(_m, "Expose", instanceId, request)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) Expose(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Expose", arg0, arg1)
}

func (_m *MockServiceEngine) GetContainerBrokerHealth() error {
	ret := _m.ctrl.Call(_m, "GetContainerBrokerHealth")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockServiceEngineRecorder) GetContainerBrokerHealth() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetContainerBrokerHealth")
}

func (_m *MockServiceEngine) GetVersions() ([]models.VersionsResponse, error) {
	ret := _m.ctrl.Call(_m, "GetVersions")
	ret0, _ := ret[0].([]models.VersionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetVersions() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetVersions")
}

func (_m *MockServiceEngine) GetInstanceLogs(instanceId string) (map[string]string, error) {
	ret := _m.ctrl.Call(_m, "GetInstanceLogs", instanceId)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetInstanceLogs(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetInstanceLogs", arg0)
}

func (_m *MockServiceEngine) GetIngressHosts(instanceId string) ([]string, error) {
	ret := _m.ctrl.Call(_m, "GetIngressHosts", instanceId)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetIngressHosts(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetIngressHosts", arg0)
}

func (_m *MockServiceEngine) GetConfigMap(configMapName string) (*api.ConfigMap, error) {
	ret := _m.ctrl.Call(_m, "GetConfigMap", configMapName)
	ret0, _ := ret[0].(*api.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetConfigMap(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConfigMap", arg0)
}

func (_m *MockServiceEngine) GetCredentials(instanceId string) ([]models.ContainerCredenials, error) {
	ret := _m.ctrl.Call(_m, "GetCredentials", instanceId)
	ret0, _ := ret[0].([]models.ContainerCredenials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetCredentials(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetCredentials", arg0)
}

func (_m *MockServiceEngine) GetSecret(secretName string) (*api.Secret, error) {
	ret := _m.ctrl.Call(_m, "GetSecret", secretName)
	ret0, _ := ret[0].(*api.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) GetSecret(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSecret", arg0)
}

func (_m *MockServiceEngine) Hide(instanceId string) ([]string, error) {
	ret := _m.ctrl.Call(_m, "Hide", instanceId)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockServiceEngineRecorder) Hide(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Hide", arg0)
}

func (_m *MockServiceEngine) ScaleInstance(instanceId string) error {
	ret := _m.ctrl.Call(_m, "ScaleInstance", instanceId)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockServiceEngineRecorder) ScaleInstance(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ScaleInstance", arg0)
}

func (_m *MockServiceEngine) ValidateAndLinkInstance(srcInstanceId string, dstInstanceId string, isBindOperation bool) (models.MessageResponse, int, error) {
	ret := _m.ctrl.Call(_m, "ValidateAndLinkInstance", srcInstanceId, dstInstanceId, isBindOperation)
	ret0, _ := ret[0].(models.MessageResponse)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockServiceEngineRecorder) ValidateAndLinkInstance(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ValidateAndLinkInstance", arg0, arg1, arg2)
}

func (_m *MockServiceEngine) ValidateAndBindDependencyInstance(srcInstanceId string, dstData models.CatalogEntityWithK8sTemplate) error {
	ret := _m.ctrl.Call(_m, "ValidateAndBindDependencyInstance", srcInstanceId, dstData)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockServiceEngineRecorder) ValidateAndBindDependencyInstance(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ValidateAndBindDependencyInstance", arg0, arg1)
}
