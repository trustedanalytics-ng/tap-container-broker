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
	"regexp"

	k8sAPI "k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

const (
	appTypeLabelKey    = "tap_app_type"
	appTypeLabelValue  = "CORE"
	k8sAppSelectorName = "app"
	imageAdressRegexp  = ".+:.+/.+:(.+)$"
	imageIDRegexp      = "docker://sha256:(.+)$"
)

func (s *ServiceEngineManager) GetVersions() ([]models.VersionsResponse, error) {
	deployments, err := s.kubernetesApi.ListDeploymentsByLabel(appTypeLabelKey, appTypeLabelValue)
	if err != nil {
		return nil, err
	}

	result := []models.VersionsResponse{}
	for _, deployment := range deployments.Items {
		appName := deployment.Name
		pods, err := s.kubernetesApi.GetPodsBySpecifiedSelector(k8sAppSelectorName, appName)
		if err != nil {
			return result, err
		}
		versionInfo, err := fetchVersionInfoFromPods(pods, appName)
		if err != nil {
			logger.Error(err)
		}
		result = append(result, versionInfo)
	}
	return result, nil
}

func fetchVersionInfoFromPods(pods *k8sAPI.PodList, appName string) (models.VersionsResponse, error) {
	if len(pods.Items) > 0 {
		pod := pods.Items[0]
		container, err := getContainer(appName, pod)
		if err != nil {
			return models.VersionsResponse{Name: appName}, err
		}

		containerStatus, err := getContainerStatus(appName, pod)
		if err != nil {
			return models.VersionsResponse{Name: appName}, err
		}

		imageVersion, err := parseImageVersion(container)
		if err != nil {
			return models.VersionsResponse{Name: appName}, err
		}
		imageSignature, err := parseImageSignature(containerStatus)
		if err != nil {
			return models.VersionsResponse{Name: appName}, err
		}

		return models.VersionsResponse{
			Name:          appName,
			Image:         container.Image,
			ImageVersions: imageVersion,
			Signature:     imageSignature,
		}, nil
	} else {
		return models.VersionsResponse{Name: appName}, fmt.Errorf("no pods found with label: %s:%s", k8sAppSelectorName, appName)
	}
}

func getContainer(appName string, pod k8sAPI.Pod) (k8sAPI.Container, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == appName {
			return container, nil
		}
	}
	return k8sAPI.Container{}, fmt.Errorf("cannot find container with name: %s", appName)
}

func getContainerStatus(appName string, pod k8sAPI.Pod) (k8sAPI.ContainerStatus, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == appName {
			return containerStatus, nil
		}
	}
	return k8sAPI.ContainerStatus{}, fmt.Errorf("cannot find containerStatus with name: %s", appName)
}

func parseImageVersion(container k8sAPI.Container) (string, error) {
	result := regexp.MustCompile(imageAdressRegexp).FindStringSubmatch(container.Image)
	if len(result) == 2 {
		return result[1], nil
	} else {
		return "", fmt.Errorf("cannot split image value: %s, from container: %s", container.Image, container.Name)
	}
}

func parseImageSignature(status k8sAPI.ContainerStatus) (string, error) {
	result := regexp.MustCompile(imageIDRegexp).FindStringSubmatch(status.ImageID)
	if len(result) == 2 {
		return result[1], nil
	} else {
		return "", fmt.Errorf("cannot split imageID value: %s, from containerStatus: %s", status.ImageID, status.Name)
	}
}
