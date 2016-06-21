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
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (s *ServiceEngineManager) DeleteKubernetesInstance(instanceId string) error {
	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return err
	}

	if data.Instance.Type == catalogModels.InstanceTypeServiceBroker {
		logger.Warningf("Delete on service-broker instance is not supported yet!")
		// todo validate if there are service broker offering instances and delete them
	}

	if _, err = s.processorService.UpdateInstanceState(instanceId, "", catalogModels.InstanceStateDestroyReq, catalogModels.InstanceStateDestroying); err != nil {
		return fmt.Errorf("InstanceID: %s has incorrect status: %v: %v", instanceId, data.Instance.State, err)
	}

	return s.removeKubernetesComponentsAndCatalogData(instanceId, data)
}

func (s *ServiceEngineManager) removeKubernetesComponentsAndCatalogData(instanceId string, data models.CatalogEntityWithK8sTemplate) error {
	if err := s.kubernetesApi.DeleteAllByInstanceId(instanceId); err != nil {
		logger.Error(fmt.Sprintf("InstanceID: %s, Can't remove kuberenetes components, error: %v", instanceId, err))
		s.processorService.UpdateInstanceState(instanceId, err.Error(), catalogModels.InstanceStateDestroying, catalogModels.InstanceStateFailure)
		return err
	}

	var hookType templateModel.HookType
	if data.Instance.Type == catalogModels.InstanceTypeServiceBroker {
		hookType = templateModel.HookTypeRemoval
	} else {
		hookType = templateModel.HookTypeDeprovision
	}

	if _, _, err := s.processorService.ProcessHook(hookType, data); err != nil {
		// if instance was in FAILURE state before we will remove it even if hook execution fails
		if lastMessage := catalogModels.GetValueFromMetadata(data.Instance.Metadata, catalogModels.LAST_STATE_CHANGE_REASON); lastMessage != catalogModels.ReasonDeleteFailure {
			s.processorService.UpdateInstanceState(instanceId, err.Error(), catalogModels.InstanceStateDestroying, catalogModels.InstanceStateFailure)
			return err
		} else {
			logger.Warningf("Hook proccesing failed! Continue removing instance %s", instanceId)
		}
	}

	if _, err := s.catalogApi.DeleteInstance(instanceId); err != nil {
		logger.Errorf("InstanceID: %s, Can't delete instance from Catalog, error: %v", instanceId, err)
		return err
	}

	if data.Instance.Type == catalogModels.InstanceTypeApplication {
		if _, err := s.catalogApi.DeleteApplication(data.Instance.ClassId); err != nil {
			logger.Errorf("InstanceID: %s, Can't delete application from Catalog, error: %v", instanceId, err)
			return err
		}
	}
	return nil
}
