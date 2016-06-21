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
package k8s

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

// Max execution time = scale time * number of deployments
func (k *K8Fabricator) DeleteAllDeployments(instanceId string, selector labels.Selector) error {
	logger.Debug("Delete deployment selector:", selector)
	deployments, err := k.listDeployments(selector)
	if err != nil {
		logger.Error("List deployment failed:", err)
		return err
	}

	for _, deployment := range deployments.Items {
		name := deployment.ObjectMeta.Name

		if err := k.ScaleDeploymentAndWait(name, instanceId, 0); err != nil {
			logger.Errorf("ScaleDeploymentAndWait for deployment %s failed: %v", name, err)
			return err
		}

		if _, err := processDeploymentVolumes(deployment, k.cephClient, false); err != nil {
			return err
		}

		logger.Debug("Deleting deployment:", name)
		err = k.client.Extensions().Deployments(api.NamespaceDefault).Delete(name, &api.DeleteOptions{})
		if err != nil {
			logger.Error("Delete deployment failed:", err)
			return err
		}
	}
	return nil
}

func (k *K8Fabricator) GetDeployment(name string) (*extensions.Deployment, error) {
	return k.client.Extensions().Deployments(api.NamespaceDefault).Get(name)
}

func (k *K8Fabricator) CreateDeployment(deployment *extensions.Deployment) (*extensions.Deployment, error) {
	return k.client.Extensions().Deployments(api.NamespaceDefault).Create(deployment)
}

func (k *K8Fabricator) ListDeployments() (*extensions.DeploymentList, error) {
	selector, err := getSelectorForManagedByLabel(managedByLabel, managedByValue)
	if err != nil {
		return nil, err
	}

	return k.listDeployments(selector)
}

func (k *K8Fabricator) ListDeploymentsByLabel(labelKey, labelValue string) (*extensions.DeploymentList, error) {
	selector, err := getSelectorForManagedByLabel(labelKey, labelValue)
	if err != nil {
		return nil, err
	}

	return k.listDeployments(selector)
}

func (k *K8Fabricator) listDeployments(selector labels.Selector) (*extensions.DeploymentList, error) {
	logger.Debug("List DeploymentList selector:", selector)
	return k.client.Extensions().Deployments(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
}

// Max execution time = models.UpdateDeploymentTimeout
func (k *K8Fabricator) UpdateDeployment(instanceId string, prepareFunction models.PrepareDeployment) error {
	condition := updateDeploymentCondition(k, prepareFunction, instanceId)
	return wait.Poll(time.Second*1, models.UpdateDeploymentTimeout, condition)
}

// Max execution time = models.UpdateDeploymentTimeout + models.WaitForScaledReplicasTimeout
func (k *K8Fabricator) ScaleDeploymentAndWait(deploymentName, instanceId string, replicas int) error {
	var prepareDeployment models.PrepareDeployment = func() (*extensions.Deployment, error) {
		deployment, err := k.client.Extensions().Deployments(api.NamespaceDefault).Get(deploymentName)
		if err != nil {
			return deployment, err
		}
		deployment.Spec.Replicas = int32(replicas)
		return deployment, nil
	}

	if err := k.UpdateDeployment(instanceId, prepareDeployment); err != nil {
		return err
	}

	deployment, err := k.client.Extensions().Deployments(api.NamespaceDefault).Get(deploymentName)
	if err != nil {
		return err
	}

	isDeploymentReadyCondition := k.areAllPodScaledCondition(deployment, instanceId, replicas)
	return wait.Poll(models.WaitForScaledReplicasInterval, models.WaitForScaledReplicasTimeout, isDeploymentReadyCondition)
}

func (k *K8Fabricator) areAllPodScaledCondition(deployment *extensions.Deployment, instanceId string, replicas int) wait.ConditionFunc {
	return func() (bool, error) {
		podSelector, err := unversioned.LabelSelectorAsSelector(deployment.Spec.Selector)
		if err != nil {
			return false, err
		}

		pods, err := k.client.Pods(api.NamespaceDefault).List(api.ListOptions{LabelSelector: podSelector})
		if err != nil {
			return false, err
		} else if len(pods.Items) == replicas {
			return true, nil
		} else {
			logger.Warningf("Number of Pods incorrect, instanceId: %s! Expecting: %d - Actual: %d, waiting...", instanceId, replicas, len(pods.Items))
			return false, nil
		}
	}
}

func updateDeploymentCondition(k *K8Fabricator, prepareFunction models.PrepareDeployment, instanceId string) wait.ConditionFunc {
	return func() (bool, error) {
		deployment, err := prepareFunction()
		if err != nil {
			return false, err
		}

		_, err = k.client.Extensions().Deployments(api.NamespaceDefault).Update(deployment)
		if err == nil {
			return true, nil
		} else if errors.IsConflict(err) {
			logger.Warningf("Update deployment Conflict error", instanceId)
			return false, nil
		}
		return false, err
	}
}
