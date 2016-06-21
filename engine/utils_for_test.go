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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/mocks"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	goodInstanceID  = "44"
	wrongInstanceID = "13"
	offeringId      = "test-offering-id"
	planId          = "test-plan-id"
	planName        = "test-plan-name"
	templateId      = "test-template-id"
	imageId         = "imageId"
	deploymentName  = "test-deployment"
	podName         = "test-pod"
	containerName   = "test-container"
)

type Mocker struct {
	service       *mocks.MockServiceProcessor
	catalogAPI    *mocks.MockTapCatalogApi
	templateAPI   *mocks.MockTemplateRepository
	caAPI         *mocks.MockTapCaApi
	cephAPI       *mocks.MockCephBroker
	kubernetesAPI *mocks.MockKubernetesApi
}

func prepareMockerAndMockedServiceEngineManager(t *testing.T) (Mocker, *ServiceEngineManager, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)

	mocker := Mocker{
		service:       mocks.NewMockServiceProcessor(mockCtrl),
		catalogAPI:    mocks.NewMockTapCatalogApi(mockCtrl),
		templateAPI:   mocks.NewMockTemplateRepository(mockCtrl),
		caAPI:         mocks.NewMockTapCaApi(mockCtrl),
		cephAPI:       mocks.NewMockCephBroker(mockCtrl),
		kubernetesAPI: mocks.NewMockKubernetesApi(mockCtrl),
	}

	return mocker, getMockedServiceEngineManager(mocker), mockCtrl
}

func getMockedServiceEngineManager(mocker Mocker) *ServiceEngineManager {
	return GetNewServiceEngineManager(mocker.kubernetesAPI, mocker.templateAPI, mocker.catalogAPI, mocker.caAPI, mocker.cephAPI, mocker.service)
}

func getSampleInstance() catalogModels.Instance {
	return catalogModels.Instance{
		Id:      goodInstanceID,
		Type:    catalogModels.InstanceTypeService,
		State:   catalogModels.InstanceStateRunning,
		ClassId: offeringId,
		Metadata: []catalogModels.Metadata{
			{Id: catalogModels.OFFERING_PLAN_ID, Value: planId},
		},
	}
}

func getSampleServiceBrokerInstanceServiceAndTemplate() (catalogModels.Instance, catalogModels.Service, templateModels.Template) {
	instance, service, template := getSampleInstanceServiceAndTemplate()

	brokerTemplateIdMetadata := catalogModels.Metadata{
		Id:    catalogModels.BROKER_TEMPLATE_ID,
		Value: templateId,
	}
	brokerOfferingIdMetadata := catalogModels.Metadata{
		Id:    catalogModels.GetPrefixedOfferingName(service.Name),
		Value: offeringId,
	}
	instance.Metadata = append(instance.Metadata, brokerTemplateIdMetadata, brokerOfferingIdMetadata)

	service.State = catalogModels.ServiceStateDeploying

	return instance, service, template
}

func getCatalogEntityWithK8sTemplateForService(instance catalogModels.Instance, service catalogModels.Service, template templateModels.Template) models.CatalogEntityWithK8sTemplate {
	return models.CatalogEntityWithK8sTemplate{
		Instance: &instance,
		Service:  &service,
		Template: &template,
	}
}

func getCatalogEntityWithK8sTemplateForApp(instance catalogModels.Instance, app catalogModels.Application, template templateModels.Template) models.CatalogEntityWithK8sTemplate {
	return models.CatalogEntityWithK8sTemplate{
		Instance:    &instance,
		Application: &app,
		Template:    &template,
	}
}

func getSampleInstanceServiceAndTemplate() (catalogModels.Instance, catalogModels.Service, templateModels.Template) {
	servicePlan := catalogModels.ServicePlan{
		Id:   planId,
		Name: planName,
	}
	service := catalogModels.Service{
		Id:         offeringId,
		TemplateId: templateId,
		Plans:      []catalogModels.ServicePlan{servicePlan},
	}

	deployment := &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name: deploymentName,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
		},
	}
	template := templateModels.Template{
		Id: templateId,
		Body: []templateModels.KubernetesComponent{
			{
				Type:        templateModels.ComponentTypeInstance,
				Deployments: []*extensions.Deployment{deployment},
			},
		},
	}
	return getSampleInstance(), service, template
}

func getSampleApplicationAndImage() (catalogModels.Application, catalogModels.Image) {
	image := catalogModels.Image{
		Id: imageId,
	}
	app := catalogModels.Application{
		ImageId:     image.Id,
		TemplateId:  templateId,
		Replication: 1,
	}

	return app, image
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

func getTestCredentials() []models.ContainerCredenials {
	const containersNumber = 2
	const containerEnvsNumber = 2

	result := []models.ContainerCredenials{}
	for j := 0; j < containersNumber; j++ {
		container := models.ContainerCredenials{}
		container.Name = fmt.Sprintf("sample container no. %d", j)
		container.Envs = make(map[string]interface{})
		for k := 0; k < containerEnvsNumber; k++ {
			container.Envs[strconv.Itoa(k)] = fmt.Sprintf("sample env no. %d", k)
		}
		result = append(result, container)
	}
	return result
}

func getSampleDeploymentWithReplicas(replicas int) *extensions.Deployment {
	return &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name: deploymentName,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32(replicas),
		},
	}
}

func getSamplePod(phase api.PodPhase, containerReady bool) api.Pod {
	return api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: podName,
		},
		Spec: api.PodSpec{},
		Status: api.PodStatus{
			Phase: phase,
			ContainerStatuses: []api.ContainerStatus{
				{
					Name:  containerName,
					Ready: containerReady,
				},
			},
		},
	}
}
func setProcessorParams(intervalSec time.Duration, tries, waitForDependenciesTries int32) {
	models.ProcessorsIntervalSec = intervalSec
	models.ProcessorsTries = tries
	models.WaitForDepsTries = waitForDependenciesTries
}
