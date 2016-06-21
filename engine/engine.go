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
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	caApi "github.com/trustedanalytics-ng/tap-ca/client"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogApi "github.com/trustedanalytics-ng/tap-catalog/client"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	cephBrokerApi "github.com/trustedanalytics-ng/tap-ceph-broker/client"
	"github.com/trustedanalytics-ng/tap-container-broker/engine/processor"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	templateApi "github.com/trustedanalytics-ng/tap-template-repository/client"
)

var logger, _ = commonLogger.InitLogger("service")

type ServiceEngine interface {
	CreateKubernetesInstance(createInstanceRequest models.CreateInstanceRequest) error
	DeleteKubernetesInstance(instanceId string) error
	Expose(instanceId string, request models.ExposeRequest) ([]string, error)
	GetContainerBrokerHealth() error
	GetVersions() ([]models.VersionsResponse, error)
	GetInstanceLogs(instanceId string) (map[string]string, error)
	GetIngressHosts(instanceId string) ([]string, error)
	GetConfigMap(configMapName string) (*api.ConfigMap, error)
	GetCredentials(instanceId string) ([]models.ContainerCredenials, error)
	GetSecret(secretName string) (*api.Secret, error)
	Hide(instanceId string) ([]string, error)
	ScaleInstance(instanceId string) error
	ValidateAndLinkInstance(srcInstanceId, dstInstanceId string, isBindOperation bool) (models.MessageResponse, int, error)
	ValidateAndBindDependencyInstance(srcInstanceId string, dstData models.CatalogEntityWithK8sTemplate) error
}

type ServiceEngineManager struct {
	kubernetesApi         k8s.KubernetesApi
	templateRepositoryApi templateApi.TemplateRepository
	catalogApi            catalogApi.TapCatalogApi
	cAApi                 caApi.TapCaApi
	cephBrokerApi         cephBrokerApi.CephBroker
	processorService      processor.ServiceProcessor
}

func GetNewServiceEngineManager(kubernetesApiConnector k8s.KubernetesApi, templateRepositoryConnector templateApi.TemplateRepository,
	catalogConnector catalogApi.TapCatalogApi, caConnector caApi.TapCaApi, cephBrokerConnector cephBrokerApi.CephBroker,
	processorService processor.ServiceProcessor) *ServiceEngineManager {

	if processorService == nil {
		processorService = processor.GetNewProcessorServiceManager(kubernetesApiConnector, templateRepositoryConnector, catalogConnector)
	}

	return &ServiceEngineManager{
		kubernetesApi:         kubernetesApiConnector,
		templateRepositoryApi: templateRepositoryConnector,
		catalogApi:            catalogConnector,
		cAApi:                 caConnector,
		cephBrokerApi:         cephBrokerConnector,
		processorService:      processorService,
	}
}

func (s *ServiceEngineManager) GetContainerBrokerHealth() error {
	if err := s.templateRepositoryApi.GetTemplateRepositoryHealth(); err != nil {
		return fmt.Errorf("template repository problem: %v", err)
	}
	if _, err := s.catalogApi.GetCatalogHealth(); err != nil {
		return fmt.Errorf("catalog problem: %v", err)
	}
	if _, err := s.cephBrokerApi.GetCephBrokerHealth(); err != nil {
		return fmt.Errorf("ceph broker problem: %v", err)
	}
	return nil
}

func (s *ServiceEngineManager) setHostnameInInstanceMetadata(instanceId string) ([]string, error) {
	hosts, err := s.kubernetesApi.GetIngressHosts(instanceId)
	if err != nil {
		return hosts, err
	}

	hostsString := strings.Join(hosts, ",")
	instanceUrlsPatch, err := builder.MakePatch("Metadata", catalogModels.Metadata{Id: "urls", Value: hostsString}, catalogModels.OperationAdd)
	if err != nil {
		return hosts, err
	}

	_, _, err = s.catalogApi.UpdateInstance(instanceId, []catalogModels.Patch{instanceUrlsPatch})
	return hosts, err
}

func (s *ServiceEngineManager) GetSecret(secretName string) (*api.Secret, error) {
	return s.kubernetesApi.GetSecret(secretName)
}

func (s *ServiceEngineManager) GetConfigMap(configMapName string) (*api.ConfigMap, error) {
	return s.kubernetesApi.GetConfigMap(configMapName)
}

func (s *ServiceEngineManager) GetIngressHosts(instanceId string) ([]string, error) {
	return s.kubernetesApi.GetIngressHosts(instanceId)
}

func (s *ServiceEngineManager) GetInstanceLogs(instanceId string) (map[string]string, error) {
	result := make(map[string]string)

	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return result, err
	}

	if err = validateInstanceState(data.Instance); err != nil {
		return result, err
	}

	if !processor.IsServiceBrokerOffering(data) {
		result, err = s.kubernetesApi.GetPodsLogs(instanceId)
		if err != nil {
			return result, err
		}
		if len(result) == 0 {
			return result, fmt.Errorf("no found matching containers, instanceId: %s", instanceId)
		}
	} else {
		logger.Warningf("service-broker offering are not supported for getting logs - instance: %s", instanceId)
	}
	return result, nil
}

func (s *ServiceEngineManager) GetCredentials(instanceId string) ([]models.ContainerCredenials, error) {
	result := []models.ContainerCredenials{}

	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return result, err
	}

	if err = validateInstanceState(data.Instance); err != nil {
		return result, err
	}

	if !processor.IsServiceBrokerOffering(data) {
		result, err = s.kubernetesApi.GetDeploymentsEnvsByInstanceId(instanceId)
		if err != nil {
			return result, err
		}
	}

	if len(data.Instance.Metadata) > 0 {
		metadataCredentials, err := convertInstanceMetadataToContainerCredenials(*data.Instance)
		if err != nil {
			return result, err
		}
		result = append(result, metadataCredentials)
	}
	return result, nil
}

func convertInstanceMetadataToContainerCredenials(instance catalogModels.Instance) (models.ContainerCredenials, error) {
	metadataCredentials := models.ContainerCredenials{}
	metadataCredentials.Name = processor.METADATA_CREDENTIALS
	metadataCredentials.Envs = make(map[string]interface{})
	for _, metadata := range instance.Metadata {
		if metadata.Id == processor.VCAP_PARAM_NAME {
			vcapEntity := processor.VcapEntity{}
			if err := json.Unmarshal([]byte(metadata.Value), &vcapEntity); err != nil {
				return metadataCredentials, err
			}

			for k, v := range vcapEntity.Credentials {
				metadataCredentials.Envs[k] = v
			}
		} else {
			metadataCredentials.Envs[metadata.Id] = metadata.Value
		}
	}
	return metadataCredentials, nil
}

func validateInstanceState(instance *catalogModels.Instance) error {
	switch instance.State {
	case catalogModels.InstanceStateStartReq, catalogModels.InstanceStateStarting, catalogModels.InstanceStateRunning,
		catalogModels.InstanceStateStopReq, catalogModels.InstanceStateStopping, catalogModels.InstanceStateStopped:
		return nil
	default:
		return fmt.Errorf("instance state is inappropriate: %s", instance.State.String())
	}
}
