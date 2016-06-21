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
package builder

import (
	"encoding/json"
	"errors"

	"github.com/trustedanalytics-ng/tap-catalog/models"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

var logger, _ = commonLogger.InitLogger("builder")

// message is optional, if pass then LAST_STATE_CHANGE_REASON key will be added/updated in Instance Metadata
func MakePatchesForInstanceStateAndLastStateMetadata(message string, currentState, stateToSet models.InstanceState) ([]models.Patch, error) {
	if stateToSet == models.InstanceStateDestroyReq && currentState == models.InstanceStateFailure {
		message = models.ReasonDeleteFailure
	}

	patches, err := makePatchesForStateUpdate(currentState.String(), stateToSet.String())
	if err != nil {
		return patches, err
	}

	if message != "" {
		lastStateChangeReasonMetadata := models.Metadata{
			Id: models.LAST_STATE_CHANGE_REASON, Value: message,
		}
		metadataPatch, err := MakePatch("Metadata", lastStateChangeReasonMetadata, models.OperationAdd)
		if err != nil {
			return patches, err
		}
		patches = append(patches, metadataPatch)
	}
	return patches, nil
}

func MakePatchesForOfferingStateUpdate(currentState, stateToSet models.ServiceState) ([]models.Patch, error) {
	return makePatchesForStateUpdate(string(currentState), string(stateToSet))
}

func makePatchesForStateUpdate(currentState, stateToSet string) ([]models.Patch, error) {
	if currentState == "" || stateToSet == "" {
		return nil, errors.New("currentState and stateToSet cannot be empty!")
	}

	patches := []models.Patch{}

	statePatch, err := MakePatchWithPreviousValue("State", stateToSet, currentState, models.OperationUpdate)
	if err != nil {
		return patches, err
	}
	patches = append(patches, statePatch)
	return patches, nil
}

func MakePatch(field string, newValue interface{}, operation models.PatchOperation) (models.Patch, error) {
	newValueByte, err := json.Marshal(newValue)
	if err != nil {
		logger.Errorf("marshal field new value %s error: %v", field, err)
		return models.Patch{}, err
	}

	patch := models.Patch{
		Operation: operation,
		Field:     &field,
		Value:     (*json.RawMessage)(&newValueByte),
	}
	return patch, nil
}

func MakePatchWithPreviousValue(field string, newValue interface{}, oldValue interface{}, operation models.PatchOperation) (models.Patch, error) {
	patch, err := MakePatch(field, newValue, operation)
	if err != nil {
		return patch, err
	}

	currentValueByte, err := json.Marshal(oldValue)
	if err != nil {
		logger.Errorf("marshal field current value %s error: %v", field, err)
		return models.Patch{}, err
	}

	patch.PrevValue = currentValueByte

	return patch, nil
}
