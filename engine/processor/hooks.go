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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/kubernetes/pkg/api"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateRepositoryModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (p *ServiceProcessorManager) ProcessHook(hookType templateRepositoryModel.HookType, data models.CatalogEntityWithK8sTemplate) (map[string]string, *catalogModels.Instance, error) {
	instanceId := data.Instance.Id
	instance := data.Instance

	if hook := data.Template.Hooks[hookType]; hook != nil {
		output, err := p.processPod(hook, instanceId)
		if err != nil {
			logger.Errorf("InstanceID: %s, Can't create hook, error: %v", instanceId, err)
			return output, instance, err
		}

		switch hookType {
		case templateRepositoryModel.HookTypeProvision, templateRepositoryModel.HookTypeDeployment:
			instance, err = p.parseHookOutputToMapAndSaveInInstanceMetadata(*instance, output, hookType)
		case templateRepositoryModel.HookTypeBind:
			output, err = parseHookOutputToMap(instanceId, output, hookType)
		default:
			err = nil
		}
		if err != nil {
			logger.Errorf("InstanceID: %s, Can't process Hook output, error: %v", instanceId, err)
		}
		return output, instance, nil
	}
	return make(map[string]string), instance, nil
}

func (p *ServiceProcessorManager) processPod(pod *api.Pod, instanceId string) (map[string]string, error) {
	logger.Debug("Processing pod:", pod.Name)
	result := make(map[string]string)
	isPodProcessed := false

	if err := p.kubernetesApi.CreatePod(*pod); err != nil {
		logger.Error("Create pod error:", err)
		return result, err
	}

	for i := int32(0); i < models.ProcessorsTries; i++ {
		if i > 0 {
			time.Sleep(models.ProcessorsIntervalSec)
		}

		updatedPod, err := p.kubernetesApi.GetPod(pod.Name)
		if err != nil {
			logger.Error("Get pod error:", err)
			return result, err
		}

		if updatedPod.Status.Phase == api.PodRunning || updatedPod.Status.Phase == api.PodPending {
			logger.Info(fmt.Sprintf("Pod with name: %s and instanceId: %s is still running. Results will be collected on next attempt", pod.Name, instanceId))
			continue
		}

		if result, err = p.kubernetesApi.GetSpecificPodLogs(*updatedPod); err != nil {
			logger.Error(fmt.Sprintf("Can't get Job logs from pod! Job name: %s, instanceId: %s", updatedPod.Name, instanceId), err)
			continue
		}

		if updatedPod.Status.Phase == api.PodFailed || updatedPod.Status.Phase == api.PodUnknown {
			err = errors.New(fmt.Sprintf("Pod with name: %s and instanceId: %s FAILED!", pod.Name, instanceId))
			logger.Error(err.Error(), result)
			return result, err
		}
		isPodProcessed = true
		break
	}

	err := p.kubernetesApi.DeletePod(pod.Name)
	if err != nil {
		logger.Error(fmt.Sprintf("Delete Pod with name: %s and instanceId :%s error: %v", pod.Name, instanceId, err))
	} else {
		logger.Info(fmt.Sprintf("Pod with name: %s and instanceId: %s DELETED!", pod.Name, instanceId))
	}

	if !isPodProcessed {
		return result, errors.New("Pod couldn't be processed! Pod name: " + pod.Name)
	}
	return result, err
}

func (p *ServiceProcessorManager) parseHookOutputToMapAndSaveInInstanceMetadata(instance catalogModels.Instance, output map[string]string, hookType templateRepositoryModel.HookType) (*catalogModels.Instance, error) {
	resultMap, err := parseHookOutputToMap(instance.Id, output, hookType)
	if err != nil {
		return &instance, err
	}

	for k, v := range resultMap {
		metadata := catalogModels.Metadata{Id: k, Value: v}
		if instance, _, err = p.UpdateInstance(instance.Id, "Metadata", metadata, catalogModels.OperationAdd); err != nil {
			logger.Errorf("Can't update Metadata in Catalog. InstanceId: %s, error: %v", instance.Id, err)
			return &instance, err
		}
	}
	logger.Debugf("Processing %s Hook type SUCCESS for InstanceId: %s", hookType, instance.Id)
	return &instance, nil
}

func parseHookOutputToMap(instanceId string, output map[string]string, hookType templateRepositoryModel.HookType) (map[string]string, error) {
	resultMap := make(map[string]string)

	for _, containerLog := range output {
		if err := json.Unmarshal([]byte(containerLog), &resultMap); err != nil {
			logger.Debugf("Can't parse hook output into map[string]string, hook type: %s, InstanceId: %s", string(hookType), instanceId)
			resultMap = make(map[string]string)
			continue
		}
		break
	}

	if len(resultMap) == 0 {
		for _, containerLog := range output {
			vcapEntity := VcapEntity{}
			if err := json.Unmarshal([]byte(containerLog), &vcapEntity); err != nil {
				logger.Debugf("Can't parse hook output into VcapEntity, hook type: %s, InstanceId: %s", string(hookType), instanceId)
				continue
			}
			logger.Debugf("PARSED output into VcapEntity, hook type: %s, InstanceId: %s", string(hookType), vcapEntity)
			resultMap[VCAP_PARAM_NAME] = containerLog
			break
		}
	}

	if len(output) != 0 && len(resultMap) == 0 {
		logger.Errorf("Processing Hook type: %s - ERROR, InstanceId: %s, Can't parse Output! Output value: %v", string(hookType), instanceId, resultMap)
	} else {
		logger.Infof("Processing Hook type: %s - SUCESSS, InstanceId: %s", string(hookType), instanceId)
	}
	return resultMap, nil
}
