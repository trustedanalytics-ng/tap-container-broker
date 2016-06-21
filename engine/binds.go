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
	"errors"
	"fmt"
	"net/http"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

func (s *ServiceEngineManager) ValidateAndLinkInstance(srcInstanceId, dstInstanceId string, isBindOperation bool) (models.MessageResponse, int, error) {
	if srcInstanceId == dstInstanceId {
		err := errors.New(string(models.DeployResponseStatusRecursiveBinding))
		return models.MessageResponse{Message: models.DeployResponseStatusRecursiveBinding}, http.StatusBadRequest, err
	}

	srcData, deployStatus, httpStatus, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: srcInstanceId, BoundInstanceId: dstInstanceId})
	if err != nil {
		return models.MessageResponse{Message: deployStatus}, httpStatus, err
	}

	dstData, deployStatus, httpStatus, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: dstInstanceId})
	if err != nil {
		return models.MessageResponse{Message: deployStatus}, httpStatus, err
	}

	if err := validateSourceInstance(srcData); err != nil {
		return models.MessageResponse{Message: models.DeployResponseStatusUnrecoverableError}, http.StatusForbidden, err
	}

	if status, err := validateDestinationInstanceBindings(srcInstanceId, dstData.Instance.Bindings, isBindOperation); err != nil {
		return models.MessageResponse{Message: models.DeployResponseStatusUnrecoverableError}, status, err
	}

	if status, err := validateDestinationInstance(dstData); err != nil {
		return models.MessageResponse{Message: models.DeployResponseStatusUnrecoverableError}, status, err
	}

	go s.processorService.LinkInstances(srcData, dstData, isBindOperation)
	return models.MessageResponse{}, http.StatusAccepted, nil
}

func (s *ServiceEngineManager) ValidateAndBindDependencyInstance(srcInstanceId string, dstData models.CatalogEntityWithK8sTemplate) error {
	if srcInstanceId == dstData.Instance.Id {
		err := errors.New(string(models.DeployResponseStatusRecursiveBinding))
		return err
	}

	srcData, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: srcInstanceId, BoundInstanceId: dstData.Instance.Id})
	if err != nil {
		return err
	}

	if err = validateSourceInstance(srcData); err != nil {
		return err
	}

	if _, err = validateDestinationInstance(dstData); err != nil {
		return err
	}

	return s.processorService.LinkInstances(srcData, dstData, true)
}

func validateSourceInstance(srcData models.CatalogEntityWithK8sTemplate) error {
	if srcData.Instance.Type == catalogModels.InstanceTypeServiceBroker {
		return errors.New("source instance can not be a ServiceBroker!")
	}

	if srcData.Instance.State != catalogModels.InstanceStateRunning && srcData.Instance.State != catalogModels.InstanceStateStopped {
		return errors.New("source instance has to be in RUNNING/STOPPED state!")
	}
	return nil
}

func validateDestinationInstance(dstData models.CatalogEntityWithK8sTemplate) (int, error) {
	switch dstData.Instance.State {
	case catalogModels.InstanceStateRunning, catalogModels.InstanceStateStopped:
		return http.StatusOK, nil
	default:
		return http.StatusForbidden, fmt.Errorf("Destination instance has inappropriate state: %s, id: %s", dstData.Instance.State.String(), dstData.Instance.Id)

	}
}

func validateDestinationInstanceBindings(srcInstanceId string, bindings []catalogModels.InstanceBindings, isBindOperation bool) (int, error) {
	isSrcBindOnList := false
	for _, binding := range bindings {
		if binding.Id == srcInstanceId {
			isSrcBindOnList = true
		}
	}

	if isBindOperation && isSrcBindOnList {
		return http.StatusConflict, errors.New("Binding already exist!")
	} else if !isBindOperation && !isSrcBindOnList {
		return http.StatusNotFound, errors.New("Binding doesn't exist!")
	}
	return http.StatusOK, nil
}
