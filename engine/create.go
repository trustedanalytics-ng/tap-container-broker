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

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/engine/processor"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (s *ServiceEngineManager) CreateKubernetesInstance(createInstanceRequest models.CreateInstanceRequest) error {
	instanceId := createInstanceRequest.InstanceId

	instance, _, err := s.catalogApi.GetInstance(instanceId)
	if err != nil {
		return fmt.Errorf("InstanceID: %s, GetInstance from Catalog error: %v", instanceId, err)
	}

	defer s.handleOperationOnInstanceError(instanceId, false, &err)

	message := "Deploying requested by user: " + instance.AuditTrail.LastUpdateBy
	_, err = s.processorService.UpdateInstanceState(instanceId, message, catalogModels.InstanceStateRequested, catalogModels.InstanceStateDeploying)
	if err != nil {
		return err
	}

	var caCertHash string
	if instance.Type == catalogModels.InstanceTypeApplication {
		caCertHash, err = s.getCACertHash()
		if err != nil {
			return err
		}
	}

	externalRequest := models.CatalogEntityWithK8sTemplateRequest{
		InstanceId:             instanceId,
		ApplicationSSLCertHash: caCertHash,
	}
	data, _, _, err := s.processorService.GetInstanceCatalogEntityWithK8STemplate(externalRequest, instance)
	if err != nil {
		logger.Errorf("InstanceId: %s, getInstanceCatalogEntityWithK8STemplate error: %v", instanceId, err)
		return err
	}

	startAction := s.getStartComponentAction(instanceId, data)

	if err = s.processorService.WaitForInstanceDependencies(data); err != nil {
		logger.Errorf("InstanceId: %s, Dependiences are not ready - error: %v", instanceId, err)
		return err
	}

	if data, err = s.runHooksAndSetReplicasAndCreateComponents(instanceId, data); err != nil {
		logger.Errorf("InstanceId: %s, Can't create components, error: %v", instanceId, err)
		return err
	}

	for _, binding := range data.Instance.Bindings {
		if err = s.ValidateAndBindDependencyInstance(binding.Id, data); err != nil {
			return fmt.Errorf("InstanceId: %s, Can't bind instance: %s, error: %v", instanceId, binding.Id, err)
		}
	}

	if _, err = s.setHostnameInInstanceMetadata(instanceId); err != nil {
		logger.Errorf("InstanceId: %s, Can't set urls in metadata, error: %v", instanceId, err)
		return err
	}

	if startAction != nil {
		if err = startAction(instanceId); err != nil {
			logger.Errorf("InstanceID: %s, Can't start component, error: %v", instanceId, err)
			return err
		}
	}
	return nil
}

func (s *ServiceEngineManager) runHooksAndSetReplicasAndCreateComponents(instanceId string, data models.CatalogEntityWithK8sTemplate) (models.CatalogEntityWithK8sTemplate, error) {
	var componentsToAdd []templateModel.KubernetesComponent
	var err error

	if data.Instance.Type == catalogModels.InstanceTypeServiceBroker {
		if _, data.Instance, err = s.processorService.ProcessHook(templateModel.HookTypeDeployment, data); err != nil {
			return data, err
		}
		componentsToAdd = adjustReplicasInDeploymentsAndReturnFilteredComponents(data, 0, serviceBrokerComponentFilter)
	} else {
		if _, data.Instance, err = s.processorService.ProcessHook(templateModel.HookTypeProvision, data); err != nil {
			return data, err
		}

		if isApplication(data) {
			if err = s.createSSLLayerForApplication(data); err != nil {
				return data, err
			}
		}
		componentsToAdd = adjustReplicasInDeploymentsAndReturnFilteredComponents(data, 0, instanceComponentFilter)
	}

	parameters := processor.GetMetadataAsMap(data.Instance.Metadata)
	if err = s.kubernetesApi.FabricateComponents(instanceId, isApplication(data), parameters, componentsToAdd); err != nil {
		return data, err
	}

	instance, _, err := s.processorService.UpdateStateAndGetInstance(instanceId, "", catalogModels.InstanceStateDeploying, catalogModels.InstanceStateStopped)
	data.Instance = &instance
	return data, err
}

func adjustReplicasInDeploymentsAndReturnFilteredComponents(data models.CatalogEntityWithK8sTemplate, replicas int32, filter componentFilter) []templateModel.KubernetesComponent {
	result := []templateModel.KubernetesComponent{}
	for _, component := range data.Template.Body {
		if filter(component) {
			for j, _ := range component.Deployments {
				component.Deployments[j].Spec.Replicas = replicas
			}
			result = append(result, component)
		}
	}
	return result
}

type componentFilter func(templateModel.KubernetesComponent) bool

func serviceBrokerComponentFilter(component templateModel.KubernetesComponent) bool {
	return component.Type == templateModel.ComponentTypeBroker || component.Type == templateModel.ComponentTypeBoth
}

func instanceComponentFilter(component templateModel.KubernetesComponent) bool {
	return component.Type == templateModel.ComponentTypeInstance || component.Type == templateModel.ComponentTypeBoth
}

type StartComponentAction func(string) error

func (s *ServiceEngineManager) UpdateInstanceStateFromStoppedToStartReq(instanceID string) error {
	_, err := s.processorService.UpdateInstanceState(instanceID, "", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq)
	return err
}

func (s *ServiceEngineManager) getStartComponentAction(instanceId string, data models.CatalogEntityWithK8sTemplate) StartComponentAction {
	if processor.IsServiceBrokerOffering(data) {
		// todo: we should validate somehow if serviceBroker offering is started successfully - https://intel-data.atlassian.net/browse/DPNG-10462
		return s.changeInstanceFromServiceBrokerStateToRunning
	} else {
		if shouldInstanceBeStarted(data) {
			return s.UpdateInstanceStateFromStoppedToStartReq
		}
	}
	return nil
}

func shouldInstanceBeStarted(data models.CatalogEntityWithK8sTemplate) bool {
	if data.Application != nil {
		if data.Application.Replication > 0 {
			return true
		}
	} else if isThereAnyDeploymentWithNonZeroReplicas(data.Template.Body) {
		return true
	}
	return false
}

func isThereAnyDeploymentWithNonZeroReplicas(components []templateModel.KubernetesComponent) bool {
	for _, component := range components {
		for _, deployment := range component.Deployments {
			if deployment.Spec.Replicas > int32(0) {
				return true
			}
		}
	}
	return false
}

func (s *ServiceEngineManager) changeInstanceFromServiceBrokerStateToRunning(instanceId string) error {
	if _, err := s.processorService.UpdateInstanceState(instanceId, "", catalogModels.InstanceStateStopped, catalogModels.InstanceStateStarting); err != nil {
		return err
	}
	_, err := s.processorService.UpdateInstanceState(instanceId, "", catalogModels.InstanceStateStarting, catalogModels.InstanceStateRunning)
	return err
}

func isApplication(data models.CatalogEntityWithK8sTemplate) bool {
	return data.Instance.Type == catalogModels.InstanceTypeApplication
}

func (s *ServiceEngineManager) updateServiceBrokerOfferings(instance *catalogModels.Instance, previousState, nextState catalogModels.ServiceState) error {
	for _, metadata := range instance.Metadata {
		if catalogModels.IsServiceBrokerOfferingMetadata(metadata) {
			service, _, err := s.catalogApi.GetService(metadata.Value)
			if err != nil {
				logger.Errorf("Service-broker instanceId: %s, Can't get Service from Catalog, error: %v", instance.Id, err)
				return err
			}
			if service.State == catalogModels.ServiceStateDeploying {
				if err := s.processorService.UpdateServiceStatus(service.Id, previousState, nextState); err != nil {
					logger.Errorf("Service-broker instanceId: %s, Can't update state to: %s, error: %v",
						instance.Id, string(nextState), err)
					return err
				}
				logger.Info("Service id: %s, name: %s state updated to: %s", service.Id, service.Name, string(nextState))
			}
		}
	}
	return nil
}

func (s *ServiceEngineManager) handleOperationOnInstanceError(instanceId string, shouldSetBrokerReadyState bool, err *error) {
	// refresh instance to get its latest State
	instance, _, getErr := s.catalogApi.GetInstance(instanceId)
	if getErr != nil {
		logger.Errorf("InstanceID: %s, Can't refresh Instance - GetInstance from Catalog error: %v", instance.Id, getErr)
		return
	}

	s.handleServiceBrokerInstanceCreation(instance, shouldSetBrokerReadyState, *err)

	if *err != nil {
		if instance.State != catalogModels.InstanceStateFailure {
			s.processorService.UpdateInstanceState(instance.Id, (*err).Error(), instance.State, catalogModels.InstanceStateFailure)
		}
	}
}

func (s *ServiceEngineManager) handleServiceBrokerInstanceCreation(instance catalogModels.Instance, shouldSetBrokerReadyState bool, err error) {
	if instance.Type == catalogModels.InstanceTypeServiceBroker {
		if err != nil {
			logger.Errorf("Service-broker deploying fail. InstanceId: %s, name: %s. All assigned offerings will be set to OFFLINE state. Cause: %v",
				instance.Id, instance.Name, err)
			s.updateServiceBrokerOfferings(&instance, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateOffline)
		} else if shouldSetBrokerReadyState {
			if instance.State == catalogModels.InstanceStateRunning {
				s.updateServiceBrokerOfferings(&instance, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateReady)
			} else {
				s.updateServiceBrokerOfferings(&instance, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateOffline)
			}
		}
	}
}
