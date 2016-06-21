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

package processor

import (
	"fmt"

	"k8s.io/kubernetes/pkg/apis/extensions"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

func (p *ServiceProcessorManager) StartInstance(instanceId string, data models.CatalogEntityWithK8sTemplate) error {
	return p.ScaleOnKubernetes(instanceId, data, catalogModels.InstanceStateStarting)
}

func (p *ServiceProcessorManager) StopInstance(instanceId string, data models.CatalogEntityWithK8sTemplate) error {
	return p.ScaleOnKubernetes(instanceId, data, catalogModels.InstanceStateStopping)
}

func (p *ServiceProcessorManager) ScaleOnKubernetes(instanceId string, data models.CatalogEntityWithK8sTemplate, stateToUpdateBeforeScale catalogModels.InstanceState) error {
	for _, component := range data.Template.Body {
		for _, deploymentSrc := range component.Deployments {
			replicasToSet := getNumberOfReplicasToSet(data, stateToUpdateBeforeScale, deploymentSrc)

			deployment, err := p.kubernetesApi.GetDeployment(deploymentSrc.Name)
			if err != nil {
				return fmt.Errorf("GetDeployment error, instanceId: %s, err: %v", instanceId, err)
			}

			if shouldRestart(stateToUpdateBeforeScale, int(deployment.Spec.Replicas), replicasToSet) {
				err = p.kubernetesApi.ScaleDeploymentAndWait(deploymentSrc.Name, instanceId, 0)
				if err != nil {
					return fmt.Errorf("ScaleDeploymentAndWait error, instanceId: %s, err: %v", instanceId, err)
				}
			}

			// TODO DPNG-14030 consider to doing it simultaneously using channels (it can take while)
			if err := p.kubernetesApi.ScaleDeploymentAndWait(deploymentSrc.Name, instanceId, replicasToSet); err != nil {
				return fmt.Errorf("ScaleDeploymentAndWait error, instanceId: %s, err: %v", instanceId, err)
			}
		}
	}
	return nil
}

func getNumberOfReplicasToSet(data models.CatalogEntityWithK8sTemplate, stateToUpdateBeforeScale catalogModels.InstanceState, deploymentSrc *extensions.Deployment) int {
	switch stateToUpdateBeforeScale {
	case catalogModels.InstanceStateStopping:
		return 0
	case catalogModels.InstanceStateStarting:
		switch data.Instance.Type {
		case catalogModels.InstanceTypeService, catalogModels.InstanceTypeServiceBroker:
			return int(deploymentSrc.Spec.Replicas)
		case catalogModels.InstanceTypeApplication:
			return data.Application.Replication
		}
	default:
		logger.Fatalf("Flow not supported: %s", catalogModels.InstanceStateStarting.String())
	}
	return 0
}

func shouldRestart(stateToUpdateBeforeScale catalogModels.InstanceState, currentReplicas, replicasToSet int) bool {
	return stateToUpdateBeforeScale == catalogModels.InstanceStateStarting && currentReplicas == replicasToSet
}
