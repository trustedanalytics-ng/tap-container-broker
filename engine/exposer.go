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
	"errors"
	"fmt"

	"k8s.io/kubernetes/pkg/api"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/engine/processor"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (s *ServiceEngineManager) Expose(instanceId string, request models.ExposeRequest) ([]string, error) {
	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return nil, err
	}

	err = s.exposeOperation(data, request)
	if err != nil {
		return nil, err
	}

	return s.setHostnameInInstanceMetadata(instanceId)
}

func (s *ServiceEngineManager) Hide(instanceId string) ([]string, error) {
	data, _, _, err := s.processorService.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: instanceId})
	if err != nil {
		return nil, err
	}

	err = s.hideOperation(data)
	if err != nil {
		return nil, err
	}

	return s.setHostnameInInstanceMetadata(instanceId)
}

func (s *ServiceEngineManager) hideOperation(data models.CatalogEntityWithK8sTemplate) error {
	if err := s.kubernetesApi.DeleteIngress(data.Instance.Id); err != nil {
		return err
	}

	if processor.IsServiceBrokerOffering(data) {
		// DeleteService removes automatically assigned endpoint
		if err := s.kubernetesApi.DeleteService(commonHttp.UuidToShortDnsName(data.Instance.Id)); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceEngineManager) exposeOperation(data models.CatalogEntityWithK8sTemplate, request models.ExposeRequest) error {
	var services []api.Service
	var err error

	switch data.Instance.Type {
	case catalogModels.InstanceTypeService:
		if processor.IsServiceBrokerOffering(data) {
			services, err = s.createServiceAndEndpoints(data.Instance.Id, request)
		} else {
			services, err = s.getServicesAndValidatePorts(data.Instance.Id, request)
		}
	case catalogModels.InstanceTypeServiceBroker, catalogModels.InstanceTypeApplication:
		err = errors.New("application and service-broker Instances are not supported!")
	}

	if err != nil {
		return err
	}

	if err = s.createIngressFromServices(data.Instance.Id, request.Hostname, services); err != nil {
		return err
	}

	return nil
}

func (s *ServiceEngineManager) createIngressFromServices(instanceId, hostname string, services []api.Service) error {
	ingress := k8s.MakeIngressForServices(instanceId, hostname, services)
	return s.kubernetesApi.CreateIngress(ingress)
}

func (s *ServiceEngineManager) getServicesAndValidatePorts(instanceId string, request models.ExposeRequest) ([]api.Service, error) {
	services, err := s.kubernetesApi.GetServiceByInstanceId(instanceId)
	if err != nil {
		return []api.Service{}, err
	}

	if len(services) < 1 {
		return []api.Service{}, errors.New("can not find service with instanceID: " + instanceId)
	}

	for i, service := range services {
		if len(request.Ports) > 0 {
			servicePorts := []api.ServicePort{}
			for _, port := range request.Ports {
				found := false
				for _, servicePort := range service.Spec.Ports {
					if servicePort.Port == port {
						found = true
						servicePorts = append(servicePorts, servicePort)
						break
					}
				}
				if !found {
					return services, fmt.Errorf("can not expose instance: %s. Ports must match service ports. Port: %d does not exist in Service!",
						instanceId, port)
				}
			}
			service.Spec.Ports = servicePorts
		}
		services[i] = service
	}
	return services, nil
}

func (s *ServiceEngineManager) createServiceAndEndpoints(instanceId string, request models.ExposeRequest) ([]api.Service, error) {
	if len(request.Ports) == 0 {
		return []api.Service{}, errors.New("ports list can not be empty!")
	}

	endpoint := k8s.MakeEndpointForPorts(instanceId, request.Ip, request.Ports)
	if err := s.kubernetesApi.CreateEndpoint(endpoint); err != nil {
		return []api.Service{}, err
	}

	service := k8s.MakeServiceForPorts(instanceId, request.Ports)
	err := s.kubernetesApi.CreateService(service)
	return []api.Service{service}, err
}
