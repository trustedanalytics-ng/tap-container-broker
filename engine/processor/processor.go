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
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogApi "github.com/trustedanalytics-ng/tap-catalog/client"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	templateApi "github.com/trustedanalytics-ng/tap-template-repository/client"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

var logger, _ = commonLogger.InitLogger("processor")

type ServiceProcessor interface {
	GetCatalogEntityWithK8sTemplate(request models.CatalogEntityWithK8sTemplateRequest) (models.CatalogEntityWithK8sTemplate, models.DeployResponseStatus, int, error)
	GetInstanceCatalogEntityWithK8STemplate(request models.CatalogEntityWithK8sTemplateRequest, instance catalogModels.Instance) (models.CatalogEntityWithK8sTemplate, models.DeployResponseStatus, int, error)
	LinkInstances(srcData, dstData models.CatalogEntityWithK8sTemplate, isBindOperation bool) error
	MonitorInstanceDeployments(instance catalogModels.Instance) error
	ProcessHook(hookType templateModel.HookType, data models.CatalogEntityWithK8sTemplate) (map[string]string, *catalogModels.Instance, error)
	StartInstance(instanceId string, data models.CatalogEntityWithK8sTemplate) error
	StopInstance(instanceId string, data models.CatalogEntityWithK8sTemplate) error
	ScaleOnKubernetes(instanceId string, data models.CatalogEntityWithK8sTemplate, stateToUpdateBeforeScale catalogModels.InstanceState) error
	UpdateInstanceState(instanceId, message string, currentState, stateToSet catalogModels.InstanceState) (int, error)
	UpdateStateAndGetInstance(instanceId, message string, currentState, stateToSet catalogModels.InstanceState) (catalogModels.Instance, int, error)
	UpdateInstance(instanceId, field string, valueToUpdate interface{}, operation catalogModels.PatchOperation) (catalogModels.Instance, int, error)
	UpdateServiceStatus(serviceId string, currentState, nextState catalogModels.ServiceState) error
	WaitForInstanceDependencies(data models.CatalogEntityWithK8sTemplate) error
}

type ServiceProcessorManager struct {
	kubernetesApi k8s.KubernetesApi
	templateApi   templateApi.TemplateRepository
	catalogApi    catalogApi.TapCatalogApi
}

func GetNewProcessorServiceManager(kubernetesApiConnector k8s.KubernetesApi, templateRepositoryConnector templateApi.TemplateRepository,
	catalogConnector catalogApi.TapCatalogApi) *ServiceProcessorManager {
	return &ServiceProcessorManager{
		kubernetesApi: kubernetesApiConnector,
		templateApi:   templateRepositoryConnector,
		catalogApi:    catalogConnector,
	}
}

func IsServiceBrokerOffering(data models.CatalogEntityWithK8sTemplate) bool {
	return data.Instance.Type == catalogModels.InstanceTypeService && templateModel.IsServiceBrokerTemplate(*data.Template)
}

func GetMetadataAsMap(metadata []catalogModels.Metadata) map[string]string {
	result := make(map[string]string)
	for _, meta := range metadata {
		result[meta.Id] = meta.Value
	}
	return result
}

func (p *ServiceProcessorManager) MonitorInstanceDeployments(instance catalogModels.Instance) error {
	instanceId := instance.Id
	lastMessage := ""
	tries := models.ProcessorsTries

	if instance.Type == catalogModels.InstanceTypeServiceBroker {
		tries = models.ProcessorsServiceBrokerTries
	}

	for i := int32(0); i < tries; i++ {
		if i > 0 {
			time.Sleep(models.ProcessorsIntervalSec)
		}

		pods, err := p.kubernetesApi.GetPodsByInstanceId(instanceId)
		if err != nil {
			logger.Errorf("GetPodsStateByInstanceId error! InstanceId: %s, err: %v", instanceId, err)
			return err
		}
		if len(pods) == 0 {
			lastMessage = fmt.Sprintf("There is no pods yet. Waiting... InstanceId: %s", instanceId)
			logger.Warning(lastMessage)
			continue
		}

		stateToUpdate, podResponsibleForInstanceStateValue, instancesState := processPodsState(instanceId, pods)

		if lastMessage, err = p.getLastMessageFromEvent(instanceId, podResponsibleForInstanceStateValue); err != nil {
			logger.Errorf("getLastMessageFromEvent error! InstanceId: %s, err: %v", instanceId, err)
			return err
		}

		if _, stateExist := instancesState[api.PodPending]; stateExist && !isLastIteration(i, tries) {
			logger.Debugf("There are still pending pods. Waiting... InstanceId: %s", instanceId)
			continue
		}

		logger.Infof("Updating Instance state to: %v. InstanceId: %s. Last message: %s", stateToUpdate, instanceId, lastMessage)
		lastMessage = formatMessage(stateToUpdate, lastMessage)
		_, err = p.UpdateInstanceState(instanceId, lastMessage, catalogModels.InstanceStateStarting, stateToUpdate)
		return err
	}
	return fmt.Errorf("pod can't be stared! Event message: %s ", lastMessage)
}

func formatMessage(stateToUpdate catalogModels.InstanceState, message string) string {
	switch stateToUpdate {
	case catalogModels.InstanceStateStopReq:
		return fmt.Sprintf("Timeout exeeded, last message: %v", message)
	case catalogModels.InstanceStateFailure:
		return fmt.Sprintf("FAILURE, reason: %v", message)
	default:
		return message
	}
}

func (p *ServiceProcessorManager) getLastMessageFromEvent(instanceId string, pod *api.Pod) (string, error) {
	lastMessage := ""
	if pod != nil {
		events, err := p.kubernetesApi.GetPodEvents(*pod)
		if err != nil {
			return lastMessage, err
		}
		if len(events) > 0 {
			lastMessage = events[len(events)-1].Message
		} else {
			logger.Warningf("[STRANGE] Pod has no messages yet! InstanceId: %s", instanceId)
		}
	} else {
		return lastMessage, fmt.Errorf("there is no assigned pod to its state! InstanceId: %s", instanceId)
	}
	return lastMessage, nil
}

func isLastIteration(counter, elements int32) bool {
	return counter == elements-1
}

func (p *ServiceProcessorManager) WaitForInstanceDependencies(data models.CatalogEntityWithK8sTemplate) error {
	bindings := data.Instance.Bindings
	for i := 0; int32(i) < models.WaitForDepsTries; i++ {
		notReadyBindings := []catalogModels.InstanceBindings{}
		for _, binding := range bindings {
			instance, _, err := p.catalogApi.GetInstance(binding.Id)
			if err != nil {
				return err
			}

			switch instance.State {
			case catalogModels.InstanceStateRunning:
				logger.Debugf("Instance %s dependency: %s READY", data.Instance.Id, instance.Name)
			case catalogModels.InstanceStateFailure, catalogModels.InstanceStateUnavailable:
				return fmt.Errorf("Instance dependency has inappropriate state: %s. Dependency name: %s",
					instance.State.String(), instance.Name)
			default:
				notReadyBindings = append(notReadyBindings, binding)
			}
		}
		bindings = notReadyBindings

		if len(bindings) == 0 {
			return nil
		}
		time.Sleep(models.ProcessorsIntervalSec)
	}
	return fmt.Errorf("timeout! %d of instance dependencies still no ready! Dependencies ids: %v", len(bindings), getBindingsIds(bindings))
}

func getBindingsIds(bindings []catalogModels.InstanceBindings) []string {
	result := []string{}
	for _, bind := range bindings {
		result = append(result, bind.Id)
	}
	return result
}

func processPodsState(instanceId string, pods []api.Pod) (catalogModels.InstanceState, *api.Pod, map[api.PodPhase]api.Pod) {
	instancesState := make(map[api.PodPhase]api.Pod)
	for _, pod := range pods {
		status := pod.Status.Phase
		// Pod statues are confusing: if container fail, pod still has status RUNNING - https://github.com/kubernetes/kubernetes/issues/17460
		if status == api.PodRunning {
			containersReady, err := areContainersReady(pod.Status.ContainerStatuses)
			if err != nil {
				status = api.PodFailed
				logger.Errorf("InstanceId: %s, pod can't be started - %s", instanceId, err.Error())
			}
			if !containersReady {
				status = api.PodPending
			}
		}
		instancesState[status] = pod
	}

	var podResponsibleForInstanceStateValue *api.Pod
	var stateToUpdate catalogModels.InstanceState
	for _, phase := range podPhaseCheckOrder {
		if pod, ok := instancesState[phase]; ok {
			stateToUpdate = mapPodPhaseToInstanceState(phase)
			podResponsibleForInstanceStateValue = &pod
		}
	}
	return stateToUpdate, podResponsibleForInstanceStateValue, instancesState
}

func areContainersReady(containerStatuses []api.ContainerStatus) (bool, error) {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Ready == false {
			// There is a moment when containers are Terminated with empty error before they get Running for the first time
			if containerStatus.State.Terminated != nil && strings.ToLower(containerStatus.State.Terminated.Reason) == "error" {
				logger.Debugf("detected first Terminated state for container: %v", containerStatus.Name)
				return false, nil
			}

			return false, fmt.Errorf("container: %v failed!", containerStatus.Name)
		}
	}
	return true, nil
}

var podPhaseCheckOrder = []api.PodPhase{
	api.PodSucceeded,
	api.PodRunning,
	api.PodPending,
	api.PodFailed,
	api.PodUnknown,
}

var podPhaseToInstanceStateMapper = map[api.PodPhase]catalogModels.InstanceState{
	api.PodSucceeded: catalogModels.InstanceStateStopReq,
	api.PodRunning:   catalogModels.InstanceStateRunning,
	api.PodPending:   catalogModels.InstanceStateStopReq,
	api.PodFailed:    catalogModels.InstanceStateFailure,
	api.PodUnknown:   catalogModels.InstanceStateFailure, // Should unknown state be treated as failure?
}

func mapPodPhaseToInstanceState(phase api.PodPhase) catalogModels.InstanceState {
	if state, ok := podPhaseToInstanceStateMapper[phase]; ok {
		return state
	}
	return catalogModels.InstanceStateFailure
}

func (p *ServiceProcessorManager) UpdateServiceStatus(serviceId string, currentState, nextState catalogModels.ServiceState) error {
	patches, err := builder.MakePatchesForOfferingStateUpdate(currentState, nextState)
	if err != nil {
		return err
	}

	_, _, err = p.catalogApi.UpdateService(serviceId, patches)
	return err
}

func (p *ServiceProcessorManager) UpdateInstanceState(instanceId, message string, currentState, stateToSet catalogModels.InstanceState) (int, error) {
	_, status, err := p.UpdateStateAndGetInstance(instanceId, message, currentState, stateToSet)
	if err != nil {
		logger.Errorf("InstanceID: %s, Can't update instance status: %v", instanceId, err)
	}
	return status, err
}

func (p *ServiceProcessorManager) UpdateStateAndGetInstance(instanceId, message string, currentState, stateToSet catalogModels.InstanceState) (catalogModels.Instance, int, error) {
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata(message, currentState, stateToSet)
	if err != nil {
		logger.Error("MakePatchesForInstanceStateAndLastStateMetadata error:", err)
		return catalogModels.Instance{}, http.StatusBadRequest, err
	}
	return p.catalogApi.UpdateInstance(instanceId, patches)
}

func (p *ServiceProcessorManager) UpdateInstance(instanceId, field string, valueToUpdate interface{}, operation catalogModels.PatchOperation) (catalogModels.Instance, int, error) {
	patch, err := builder.MakePatch(field, valueToUpdate, operation)
	if err != nil {
		logger.Error("MakePatch error:", err)
		return catalogModels.Instance{}, http.StatusBadRequest, err
	}
	return p.catalogApi.UpdateInstance(instanceId, []catalogModels.Patch{patch})
}
