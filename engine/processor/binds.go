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
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (p *ServiceProcessorManager) LinkInstances(srcData, dstData models.CatalogEntityWithK8sTemplate, isBindOperation bool) error {
	var err error
	defer p.handleBindingOperationResult(dstData.Instance.Id, dstData.Instance.State, &err)

	err = p.setUnavailableStateAndStopInstance(dstData)
	if err != nil {
		return err
	}

	srcEnvVars, volumeDisks, err := getEnvVarsAndVolumeDisksFromComponentsDeployments(srcData.Template.Body, srcData.Instance.Name)
	if err != nil {
		return err
	}

	hookType, operationType := getHookTypeAndPatchOperation(isBindOperation)

	output, _, err := p.ProcessHook(hookType, srcData)
	if err != nil {
		logger.Error("Can't ProcessHook:", err)
		return err
	}

	_, _, err = p.UpdateInstance(dstData.Instance.Id, "Bindings", prepareInstanceBinding(srcData, srcEnvVars, output), operationType)
	if err != nil {
		logger.Error("Can't UpdateInstance:", err)
		return err
	}

	for _, component := range dstData.Template.Body {
		for _, deploymentDst := range component.Deployments {
			var podSpecOperation podSpecProcessorFunc
			if isBindOperation {
				podSpecOperation = bindDataToPodSpec(srcData, srcEnvVars, output, volumeDisks)
			} else {
				podSpecOperation = unbindDataFromPodSpec(srcData, dstData)
			}

			err = p.updateDeploymentByPodSpecOperation(podSpecOperation, dstData, deploymentDst.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *ServiceProcessorManager) setUnavailableStateAndStopInstance(dstData models.CatalogEntityWithK8sTemplate) error {
	_, err := p.UpdateInstanceState(dstData.Instance.Id, "", dstData.Instance.State, catalogModels.InstanceStateUnavailable)
	if err != nil {
		return err
	}
	if dstData.Instance.State == catalogModels.InstanceStateRunning {
		if err := p.StopInstance(dstData.Instance.Id, dstData); err != nil {
			logger.Errorf("Cannot stop instance %s, error: %v", dstData.Instance.Id, err)
			return err
		}
	}
	return nil
}

func (p *ServiceProcessorManager) handleBindingOperationResult(instanceId string, instanceStateBeforeBind catalogModels.InstanceState, err *error) {
	message := ""
	if err != nil && *err != nil {
		message = fmt.Sprintf("Bind operation FAILED: %v", *err)
		logger.Errorf("InstanceID: %s, error: %v", instanceId, message)
	}
	p.UpdateInstanceState(instanceId, message, catalogModels.InstanceStateUnavailable, catalogModels.InstanceStateStopped)

	//Run instance if it was running before
	if (err == nil || *err == nil) && instanceStateBeforeBind == catalogModels.InstanceStateRunning {
		p.UpdateInstanceState(instanceId, message, catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq)
	}
}

func getHookTypeAndPatchOperation(isBindOperation bool) (templateModel.HookType, catalogModels.PatchOperation) {
	var hookType templateModel.HookType
	var operationType catalogModels.PatchOperation

	if isBindOperation {
		hookType = templateModel.HookTypeBind
		operationType = catalogModels.OperationAdd
	} else {
		hookType = templateModel.HookTypeUnbind
		operationType = catalogModels.OperationDelete
	}
	return hookType, operationType
}

func (p *ServiceProcessorManager) updateDeploymentByPodSpecOperation(podSpecOperation podSpecProcessorFunc, dstData models.CatalogEntityWithK8sTemplate, deploymentName string) error {
	var prepareDeployment models.PrepareDeployment = func() (*extensions.Deployment, error) {
		deployment, err := p.kubernetesApi.GetDeployment(deploymentName)
		if err != nil {
			logger.Error("Can't GetDeployment:", err)
			return deployment, err
		}

		deployment.Spec.Template.Spec, err = podSpecOperation(deployment.Spec.Template.Spec)
		if err != nil {
			logger.Error("Can't execute podSpecOperation:", err)
			return deployment, err
		}
		return deployment, nil
	}
	return p.kubernetesApi.UpdateDeployment(dstData.Instance.Id, prepareDeployment)
}

type podSpecProcessorFunc func(api.PodSpec) (api.PodSpec, error)

func bindDataToPodSpec(srcData models.CatalogEntityWithK8sTemplate, srcEnvs []api.EnvVar, output map[string]string, volumeDisks []VolumeDisk) podSpecProcessorFunc {
	return func(podSpec api.PodSpec) (api.PodSpec, error) {
		var err error

		for i, container := range podSpec.Containers {
			container, err = bindEnvVarsToContainer(container, srcData, srcEnvs, output)
			if err != nil {
				logger.Error("Can't bindEnvVarsToContainer:", err)
				return podSpec, err
			}

			for _, volumeDisk := range volumeDisks {
				container.VolumeMounts = append(container.VolumeMounts, getVolumeMountFromVolumeDisk(srcData.Instance.Name, volumeDisk))
			}
			podSpec.Containers[i] = container
		}

		for _, volumeDisk := range volumeDisks {
			podSpec.Volumes = append(podSpec.Volumes, getVolumeFromVolumeDisk(srcData.Instance.Name, volumeDisk))
		}

		return podSpec, nil
	}
}

func unbindDataFromPodSpec(srcData, dstData models.CatalogEntityWithK8sTemplate) podSpecProcessorFunc {
	return func(podSpec api.PodSpec) (api.PodSpec, error) {
		var err error
		for i, container := range podSpec.Containers {
			container, err = unbindEnvVarsFromContainer(container, srcData, dstData)
			if err != nil {
				logger.Error("Can't unbindEnvVarsFromContainer:", err)
				return podSpec, err
			}

			volumeMounts := []api.VolumeMount{}
			for _, volumeMount := range container.VolumeMounts {
				if !strings.HasPrefix(volumeMount.Name, getVolumeNamePrefix(srcData.Instance.Name)) {
					volumeMounts = append(volumeMounts, volumeMount)
				}
			}
			container.VolumeMounts = volumeMounts

			podSpec.Containers[i] = container
		}

		volumes := []api.Volume{}
		for _, volume := range podSpec.Volumes {
			if !strings.HasPrefix(volume.Name, getVolumeNamePrefix(srcData.Instance.Name)) {
				volumes = append(volumes, volume)
			}
			podSpec.Volumes = volumes
		}

		return podSpec, nil
	}
}

func bindEnvVarsToContainer(container api.Container, srcData models.CatalogEntityWithK8sTemplate, srcEnvs []api.EnvVar, output map[string]string) (api.Container, error) {
	var err error
	if IsServiceBrokerOffering(srcData) {
		container.Env, err = updateEnvByBindResponse(container.Env, output, srcData.Service.Name, srcData.Instance.Name, true)
		if err != nil {
			logger.Error("Can't updateEnvByBindResponse:", err)
			return container, err
		}
	} else {
		container.Env = append(container.Env, srcEnvs...)
		if srcData.Service != nil {
			container.Env = addToBoundServicesRegistry(container.Env, srcData.Service.Name, srcData.Instance.Name)
		}
	}
	return container, nil
}

func unbindEnvVarsFromContainer(container api.Container, srcData, dstData models.CatalogEntityWithK8sTemplate) (api.Container, error) {
	var err error
	if IsServiceBrokerOffering(srcData) {
		for _, binding := range dstData.Instance.Bindings {
			if binding.Id == srcData.Instance.Id {
				container.Env, err = updateEnvByBindResponse(container.Env, binding.Data, srcData.Service.Name, srcData.Instance.Name, false)
				if err != nil {
					logger.Error("Can't updateEnvByBindResponse:", err)
					return container, err
				}
				break
			}
		}
	} else {
		updatedEnv := []api.EnvVar{}
		for _, env := range container.Env {
			if !strings.Contains(env.Name, getPrefixedEnvName(srcData.Instance.Name, "")) {
				updatedEnv = append(updatedEnv, env)
			}
		}
		container.Env = updatedEnv
		if srcData.Service != nil {
			container.Env = removeFromBoundServicesRegistry(container.Env, srcData.Service.Name, srcData.Instance.Name)
		}
	}
	return container, nil
}

func prepareInstanceBinding(srcData models.CatalogEntityWithK8sTemplate, srcEnvs []api.EnvVar, output map[string]string) catalogModels.InstanceBindings {
	bindingsData := map[string]string{}
	if IsServiceBrokerOffering(srcData) {
		bindingsData = output
	} else {
		for _, env := range srcEnvs {
			bindingsData[env.Name] = env.Value
		}
	}
	return catalogModels.InstanceBindings{Id: srcData.Instance.Id, Data: bindingsData}
}
