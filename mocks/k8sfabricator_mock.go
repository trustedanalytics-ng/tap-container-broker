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
// Source: k8s/k8sfabricator.go

package mocks

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/trustedanalytics-ng/tap-container-broker/models"
	model "github.com/trustedanalytics-ng/tap-template-repository/model"
	api "k8s.io/kubernetes/pkg/api"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	extensions "k8s.io/kubernetes/pkg/apis/extensions"
)

// Mock of KubernetesApi interface
type MockKubernetesApi struct {
	ctrl     *gomock.Controller
	recorder *_MockKubernetesApiRecorder
}

// Recorder for MockKubernetesApi (not exported)
type _MockKubernetesApiRecorder struct {
	mock *MockKubernetesApi
}

func NewMockKubernetesApi(ctrl *gomock.Controller) *MockKubernetesApi {
	mock := &MockKubernetesApi{ctrl: ctrl}
	mock.recorder = &_MockKubernetesApiRecorder{mock}
	return mock
}

func (_m *MockKubernetesApi) EXPECT() *_MockKubernetesApiRecorder {
	return _m.recorder
}

func (_m *MockKubernetesApi) FabricateComponents(instanceId string, shouldOverwriteEnvs bool, parameters map[string]string, components []model.KubernetesComponent) error {
	ret := _m.ctrl.Call(_m, "FabricateComponents", instanceId, shouldOverwriteEnvs, parameters, components)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) FabricateComponents(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "FabricateComponents", arg0, arg1, arg2, arg3)
}

func (_m *MockKubernetesApi) GetFabricatedComponentsForAllOrgs() ([]*model.KubernetesComponent, error) {
	ret := _m.ctrl.Call(_m, "GetFabricatedComponentsForAllOrgs")
	ret0, _ := ret[0].([]*model.KubernetesComponent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetFabricatedComponentsForAllOrgs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFabricatedComponentsForAllOrgs")
}

func (_m *MockKubernetesApi) GetFabricatedComponents(organization string) ([]*model.KubernetesComponent, error) {
	ret := _m.ctrl.Call(_m, "GetFabricatedComponents", organization)
	ret0, _ := ret[0].([]*model.KubernetesComponent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetFabricatedComponents(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFabricatedComponents", arg0)
}

func (_m *MockKubernetesApi) DeleteAllByInstanceId(instanceId string) error {
	ret := _m.ctrl.Call(_m, "DeleteAllByInstanceId", instanceId)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteAllByInstanceId(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteAllByInstanceId", arg0)
}

func (_m *MockKubernetesApi) DeleteAllPersistentVolumeClaims() error {
	ret := _m.ctrl.Call(_m, "DeleteAllPersistentVolumeClaims")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteAllPersistentVolumeClaims() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteAllPersistentVolumeClaims")
}

func (_m *MockKubernetesApi) GetAllPersistentVolumes() ([]api.PersistentVolume, error) {
	ret := _m.ctrl.Call(_m, "GetAllPersistentVolumes")
	ret0, _ := ret[0].([]api.PersistentVolume)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetAllPersistentVolumes() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetAllPersistentVolumes")
}

func (_m *MockKubernetesApi) GetDeploymentsEnvsByInstanceId(instanceId string) ([]models.ContainerCredenials, error) {
	ret := _m.ctrl.Call(_m, "GetDeploymentsEnvsByInstanceId", instanceId)
	ret0, _ := ret[0].([]models.ContainerCredenials)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetDeploymentsEnvsByInstanceId(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetDeploymentsEnvsByInstanceId", arg0)
}

func (_m *MockKubernetesApi) GetService(name string) (*api.Service, error) {
	ret := _m.ctrl.Call(_m, "GetService", name)
	ret0, _ := ret[0].(*api.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetService(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetService", arg0)
}

func (_m *MockKubernetesApi) CreateService(service api.Service) error {
	ret := _m.ctrl.Call(_m, "CreateService", service)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateService(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateService", arg0)
}

func (_m *MockKubernetesApi) DeleteService(name string) error {
	ret := _m.ctrl.Call(_m, "DeleteService", name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteService(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteService", arg0)
}

func (_m *MockKubernetesApi) GetServiceByInstanceId(instanceId string) ([]api.Service, error) {
	ret := _m.ctrl.Call(_m, "GetServiceByInstanceId", instanceId)
	ret0, _ := ret[0].([]api.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetServiceByInstanceId(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetServiceByInstanceId", arg0)
}

func (_m *MockKubernetesApi) GetServices() ([]api.Service, error) {
	ret := _m.ctrl.Call(_m, "GetServices")
	ret0, _ := ret[0].([]api.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetServices() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetServices")
}

func (_m *MockKubernetesApi) GetEndpoint(name string) (*api.Endpoints, error) {
	ret := _m.ctrl.Call(_m, "GetEndpoint", name)
	ret0, _ := ret[0].(*api.Endpoints)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetEndpoint", arg0)
}

func (_m *MockKubernetesApi) CreateEndpoint(endpoint api.Endpoints) error {
	ret := _m.ctrl.Call(_m, "CreateEndpoint", endpoint)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateEndpoint", arg0)
}

func (_m *MockKubernetesApi) DeleteEndpoint(name string) error {
	ret := _m.ctrl.Call(_m, "DeleteEndpoint", name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteEndpoint", arg0)
}

func (_m *MockKubernetesApi) GetPodsByInstanceId(instanceId string) ([]api.Pod, error) {
	ret := _m.ctrl.Call(_m, "GetPodsByInstanceId", instanceId)
	ret0, _ := ret[0].([]api.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetPodsByInstanceId(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPodsByInstanceId", arg0)
}

func (_m *MockKubernetesApi) ListDeployments() (*extensions.DeploymentList, error) {
	ret := _m.ctrl.Call(_m, "ListDeployments")
	ret0, _ := ret[0].(*extensions.DeploymentList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) ListDeployments() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListDeployments")
}

func (_m *MockKubernetesApi) ListDeploymentsByLabel(labelKey string, labelValue string) (*extensions.DeploymentList, error) {
	ret := _m.ctrl.Call(_m, "ListDeploymentsByLabel", labelKey, labelValue)
	ret0, _ := ret[0].(*extensions.DeploymentList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) ListDeploymentsByLabel(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListDeploymentsByLabel", arg0, arg1)
}

func (_m *MockKubernetesApi) CreateConfigMap(configMap *api.ConfigMap) error {
	ret := _m.ctrl.Call(_m, "CreateConfigMap", configMap)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateConfigMap(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateConfigMap", arg0)
}

func (_m *MockKubernetesApi) GetConfigMap(name string) (*api.ConfigMap, error) {
	ret := _m.ctrl.Call(_m, "GetConfigMap", name)
	ret0, _ := ret[0].(*api.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetConfigMap(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConfigMap", arg0)
}

func (_m *MockKubernetesApi) GetSecret(name string) (*api.Secret, error) {
	ret := _m.ctrl.Call(_m, "GetSecret", name)
	ret0, _ := ret[0].(*api.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetSecret(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSecret", arg0)
}

func (_m *MockKubernetesApi) CreateSecret(secret api.Secret) error {
	ret := _m.ctrl.Call(_m, "CreateSecret", secret)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateSecret(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateSecret", arg0)
}

func (_m *MockKubernetesApi) DeleteSecret(name string) error {
	ret := _m.ctrl.Call(_m, "DeleteSecret", name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteSecret(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteSecret", arg0)
}

func (_m *MockKubernetesApi) GetIngress(name string) (*extensions.Ingress, error) {
	ret := _m.ctrl.Call(_m, "GetIngress", name)
	ret0, _ := ret[0].(*extensions.Ingress)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetIngress(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetIngress", arg0)
}

func (_m *MockKubernetesApi) CreateIngress(ingress extensions.Ingress) error {
	ret := _m.ctrl.Call(_m, "CreateIngress", ingress)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateIngress(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateIngress", arg0)
}

func (_m *MockKubernetesApi) DeleteIngress(name string) error {
	ret := _m.ctrl.Call(_m, "DeleteIngress", name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteIngress(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteIngress", arg0)
}

func (_m *MockKubernetesApi) GetJobs() (*batch.JobList, error) {
	ret := _m.ctrl.Call(_m, "GetJobs")
	ret0, _ := ret[0].(*batch.JobList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetJobs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetJobs")
}

func (_m *MockKubernetesApi) GetJobsByInstanceId(instanceId string) (*batch.JobList, error) {
	ret := _m.ctrl.Call(_m, "GetJobsByInstanceId", instanceId)
	ret0, _ := ret[0].(*batch.JobList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetJobsByInstanceId(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetJobsByInstanceId", arg0)
}

func (_m *MockKubernetesApi) DeleteJob(jobName string) error {
	ret := _m.ctrl.Call(_m, "DeleteJob", jobName)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeleteJob(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteJob", arg0)
}

func (_m *MockKubernetesApi) GetPod(name string) (*api.Pod, error) {
	ret := _m.ctrl.Call(_m, "GetPod", name)
	ret0, _ := ret[0].(*api.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetPod(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPod", arg0)
}

func (_m *MockKubernetesApi) GetPodsBySpecifiedSelector(labelKey string, labelValue string) (*api.PodList, error) {
	ret := _m.ctrl.Call(_m, "GetPodsBySpecifiedSelector", labelKey, labelValue)
	ret0, _ := ret[0].(*api.PodList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetPodsBySpecifiedSelector(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPodsBySpecifiedSelector", arg0, arg1)
}

func (_m *MockKubernetesApi) CreatePod(pod api.Pod) error {
	ret := _m.ctrl.Call(_m, "CreatePod", pod)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreatePod(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreatePod", arg0)
}

func (_m *MockKubernetesApi) DeletePod(podName string) error {
	ret := _m.ctrl.Call(_m, "DeletePod", podName)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) DeletePod(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeletePod", arg0)
}

func (_m *MockKubernetesApi) GetSpecificPodLogs(pod api.Pod) (map[string]string, error) {
	ret := _m.ctrl.Call(_m, "GetSpecificPodLogs", pod)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetSpecificPodLogs(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSpecificPodLogs", arg0)
}

func (_m *MockKubernetesApi) GetPodsLogs(instanceId string) (map[string]string, error) {
	ret := _m.ctrl.Call(_m, "GetPodsLogs", instanceId)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetPodsLogs(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPodsLogs", arg0)
}

func (_m *MockKubernetesApi) GetJobLogs(job batch.Job) (map[string]string, error) {
	ret := _m.ctrl.Call(_m, "GetJobLogs", job)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetJobLogs(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetJobLogs", arg0)
}

func (_m *MockKubernetesApi) CreateJob(job *batch.Job, instanceId string) error {
	ret := _m.ctrl.Call(_m, "CreateJob", job, instanceId)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) CreateJob(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateJob", arg0, arg1)
}

func (_m *MockKubernetesApi) ScaleDeploymentAndWait(deploymentName string, instanceId string, replicas int) error {
	ret := _m.ctrl.Call(_m, "ScaleDeploymentAndWait", deploymentName, instanceId, replicas)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) ScaleDeploymentAndWait(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ScaleDeploymentAndWait", arg0, arg1, arg2)
}

func (_m *MockKubernetesApi) UpdateDeployment(instanceId string, prepareFunction models.PrepareDeployment) error {
	ret := _m.ctrl.Call(_m, "UpdateDeployment", instanceId, prepareFunction)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockKubernetesApiRecorder) UpdateDeployment(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "UpdateDeployment", arg0, arg1)
}

func (_m *MockKubernetesApi) GetDeployment(name string) (*extensions.Deployment, error) {
	ret := _m.ctrl.Call(_m, "GetDeployment", name)
	ret0, _ := ret[0].(*extensions.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetDeployment(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetDeployment", arg0)
}

func (_m *MockKubernetesApi) GetIngressHosts(instanceId string) ([]string, error) {
	ret := _m.ctrl.Call(_m, "GetIngressHosts", instanceId)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetIngressHosts(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetIngressHosts", arg0)
}

func (_m *MockKubernetesApi) GetPodEvents(pod api.Pod) ([]api.Event, error) {
	ret := _m.ctrl.Call(_m, "GetPodEvents", pod)
	ret0, _ := ret[0].([]api.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockKubernetesApiRecorder) GetPodEvents(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPodEvents", arg0)
}
