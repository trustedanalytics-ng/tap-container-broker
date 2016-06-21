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
	"fmt"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/engine/processor"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (s *ServiceEngineManager) ScaleInstance(instanceId string) error {
	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return err
	}

	stateToUpdateBeforeScale, stateToUpdateAfterScale, err := getStatesToUpdate(data)
	if err != nil {
		return err
	}
	logger.Infof("Scaling instance with id: %s and state: %s to state: %s", data.Instance.Id, data.Instance.State.String(), stateToUpdateAfterScale.String())

	if _, err := s.processorService.UpdateInstanceState(instanceId, "", data.Instance.State, stateToUpdateBeforeScale); err != nil {
		return fmt.Errorf("updateInstanceStatus BEFORE scale error, instanceId: %s, err: %v", instanceId, err)
	}

	defer s.handleOperationOnInstanceError(instanceId, true, &err)

	//stopping/starting for broker-service instances is not supported
	if !processor.IsServiceBrokerOffering(data) {
		if err = s.processorService.ScaleOnKubernetes(instanceId, data, stateToUpdateBeforeScale); err != nil {
			return err
		}

		if shouldStartInstance(data, stateToUpdateBeforeScale) {
			err = s.processorService.MonitorInstanceDeployments(*data.Instance)
			return err
		}
	}

	if _, err := s.processorService.UpdateInstanceState(instanceId, "", stateToUpdateBeforeScale, stateToUpdateAfterScale); err != nil {
		return fmt.Errorf("updateInstanceStatus AFTER scale error, instanceId: %s, err: %v", instanceId, err)
	}
	return nil
}

func shouldStartInstance(data models.CatalogEntityWithK8sTemplate, stateToUpdateBeforeScale catalogModels.InstanceState) bool {
	return isThereAnyDeployment(data.Template.Body) && stateToUpdateBeforeScale == catalogModels.InstanceStateStarting
}

func isThereAnyDeployment(components []templateModel.KubernetesComponent) bool {
	for _, component := range components {
		if len(component.Deployments) > 0 {
			return true
		}
	}
	return false
}

func getStatesToUpdate(data models.CatalogEntityWithK8sTemplate) (catalogModels.InstanceState, catalogModels.InstanceState, error) {
	switch data.Instance.State {
	case catalogModels.InstanceStateStopReq:
		return getStopStates()
	case catalogModels.InstanceStateStartReq:
		if isApplication(data) && data.Application.Replication == 0 {
			return catalogModels.InstanceStateFailure, catalogModels.InstanceStateFailure,
				fmt.Errorf("can not start application with %d replicas!", data.Application.Replication)
		}
		return getStartStates()
	case catalogModels.InstanceStateReconfiguration:
		if isApplication(data) {
			if data.Application.Replication == 0 {
				return getStopStates()
			} else {
				return getStartStates()
			}
		} else {
			// restart case - scale is supported only for Application Instance type
			return getStartStates()
		}
	default:
		return catalogModels.InstanceStateFailure, catalogModels.InstanceStateFailure,
			fmt.Errorf("inappropriate state: %s for scaling", data.Instance.State.String())
	}
}

func getStopStates() (catalogModels.InstanceState, catalogModels.InstanceState, error) {
	return catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped, nil
}

func getStartStates() (catalogModels.InstanceState, catalogModels.InstanceState, error) {
	return catalogModels.InstanceStateStarting, catalogModels.InstanceStateRunning, nil
}
