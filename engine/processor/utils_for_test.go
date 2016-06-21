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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/trustedanalytics-ng/tap-catalog/builder"
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

var (
	patchesInstanceStateStartingToRunning    []catalogModels.Patch
	patchesInstanceStateRunningToUnavailable []catalogModels.Patch
	patchesInstanceStateUnavailableToStopped []catalogModels.Patch
	patchesInstanceStateStoppedToStartReq    []catalogModels.Patch
)

func init() {
	patchesInstanceStateStartingToRunning, _ = builder.MakePatchesForInstanceStateAndLastStateMetadata("", catalogModels.InstanceStateStarting, catalogModels.InstanceStateRunning)
	patchesInstanceStateRunningToUnavailable, _ = builder.MakePatchesForInstanceStateAndLastStateMetadata("", catalogModels.InstanceStateRunning, catalogModels.InstanceStateUnavailable)
	patchesInstanceStateUnavailableToStopped, _ = builder.MakePatchesForInstanceStateAndLastStateMetadata("", catalogModels.InstanceStateUnavailable, catalogModels.InstanceStateStopped)
	patchesInstanceStateStoppedToStartReq, _ = builder.MakePatchesForInstanceStateAndLastStateMetadata("", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq)
}

type Mocker struct {
	kubernetesAPI *mocks.MockKubernetesApi
	catalogAPI    *mocks.MockTapCatalogApi
	templateAPI   *mocks.MockTemplateRepository
}

func prepareMockerAndMockerServiceProcessorManager(t *testing.T) (Mocker, *ServiceProcessorManager, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)

	mocker := Mocker{
		kubernetesAPI: mocks.NewMockKubernetesApi(mockCtrl),
		catalogAPI:    mocks.NewMockTapCatalogApi(mockCtrl),
		templateAPI:   mocks.NewMockTemplateRepository(mockCtrl),
	}

	return mocker, getMockedServiceProcessorManager(mocker), mockCtrl
}

func getMockedServiceProcessorManager(mocker Mocker) *ServiceProcessorManager {
	return GetNewProcessorServiceManager(mocker.kubernetesAPI, mocker.templateAPI, mocker.catalogAPI)
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

func getCatalogEntityWithK8sTemplateForService() models.CatalogEntityWithK8sTemplate {
	instance, service, template := getSampleInstanceServiceAndTemplate()
	return models.CatalogEntityWithK8sTemplate{
		Instance: &instance,
		Service:  &service,
		Template: &template,
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
