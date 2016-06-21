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
package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	k8sAPI "k8s.io/kubernetes/pkg/api"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	"github.com/trustedanalytics-ng/tap-go-common/util"
	templateRepositoryModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func getJavaOpts() string {
	result := ""
	if xms := os.Getenv("JAVA_XMS"); xms != "" {
		result = result + fmt.Sprintf("-Xms%s", xms)
	}
	if xmx := os.Getenv("JAVA_XMX"); xmx != "" {
		if result != "" {
			result += " "
		}
		result = result + fmt.Sprintf("-Xmx%s", xmx)
	}
	if metaspace := os.Getenv("JAVA_METASPACE"); metaspace != "" {
		if result != "" {
			result += " "
		}
		result = result + fmt.Sprintf("-XX:MetaspaceSize=%s", metaspace)
	}
	return result
}

func addAdditionalEnvsAndSetReplicasToApplicationTemplate(template *templateRepositoryModel.Template, image *catalogModels.Image, replicas int) {
	for a, _ := range template.Body {
		for i, _ := range template.Body[a].Deployments {
			for j, _ := range template.Body[a].Deployments[i].Spec.Template.Spec.Containers {
				template.Body[a].Deployments[i].Spec.Replicas = int32(replicas)
				if image == nil || image.Type != catalogModels.ImageTypeJava {
					continue
				}
				if javaOpts := getJavaOpts(); javaOpts != "" {
					env := k8sAPI.EnvVar{Name: "JAVA_OPTS", Value: javaOpts}
					template.Body[a].Deployments[i].Spec.Template.Spec.Containers[j].Env = append(template.Body[a].Deployments[i].Spec.Template.Spec.Containers[j].Env, env)
				}
			}
		}
	}
}

func (s *ServiceProcessorManager) GetCatalogEntityWithK8sTemplate(request models.CatalogEntityWithK8sTemplateRequest) (
	models.CatalogEntityWithK8sTemplate, models.DeployResponseStatus, int, error) {

	instance, httpStatus, err := s.catalogApi.GetInstance(request.InstanceId)
	if err != nil {
		var message models.DeployResponseStatus
		if commonHttp.IsNotFoundError(err) {
			err = fmt.Errorf("service instance %v is not found", request.InstanceId)
			message = models.DeployResponseStatusServiceNotFound
		} else {
			err = fmt.Errorf("GetInstance %q from Catalog error: %v", request.InstanceId, err)
			message = models.DeployResponseStatusUnknown
		}
		logger.Error(err.Error())
		return models.CatalogEntityWithK8sTemplate{}, message, httpStatus, err
	}

	return s.GetInstanceCatalogEntityWithK8STemplate(request, instance)
}

func (s *ServiceProcessorManager) GetInstanceCatalogEntityWithK8STemplate(request models.CatalogEntityWithK8sTemplateRequest, instance catalogModels.Instance) (
	models.CatalogEntityWithK8sTemplate, models.DeployResponseStatus, int, error) {

	result := models.CatalogEntityWithK8sTemplate{}
	var status int
	var err error

	result.Instance = &instance

	switch instance.Type {
	case catalogModels.InstanceTypeApplication:
		result, status, err = s.getApplicationCatalogEntityWithK8sTemplate(result, request)
	case catalogModels.InstanceTypeService:
		result, status, err = s.getServiceCatalogEntityWithK8sTemplate(result, request)
	case catalogModels.InstanceTypeServiceBroker:
		result, status, err = s.getServiceBrokerCatalogEntityWithK8sTemplate(result, request)
	default:
		return result, models.DeployResponseStatusUnknown, http.StatusInternalServerError,
			fmt.Errorf("Following Instance type is not supported: %v", result.Instance.Type)
	}

	if err != nil {
		return result, models.DeployResponseStatusUnknown, status, err
	}
	return result, models.DeployResponseStatusInvalidConfiguration, http.StatusOK, err
}

func (s *ServiceProcessorManager) getServiceBrokerCatalogEntityWithK8sTemplate(data models.CatalogEntityWithK8sTemplate,
	request models.CatalogEntityWithK8sTemplateRequest) (models.CatalogEntityWithK8sTemplate, int, error) {

	templateId := catalogModels.GetValueFromMetadata(data.Instance.Metadata, catalogModels.BROKER_TEMPLATE_ID)
	if templateId == "" {
		err := errors.New("BROKER_TEMPLATE_ID has to exist in Metadata and not be empty for service-broker instance!")
		logger.Error(err)
		return data, http.StatusBadRequest, err
	}

	additionalReplacements := prepareAdditionalReplacements(data, request)
	template, status, err := s.templateApi.GenerateParsedTemplate(templateId, request.InstanceId,
		templateRepositoryModel.EMPTY_PLAN_NAME, additionalReplacements)
	if err != nil {
		logger.Error("templateApi.GenerateParsedTemplate failed: ", err)
		return data, status, err
	}
	data.Template = &template
	return data, http.StatusOK, nil
}

func (s *ServiceProcessorManager) getServiceCatalogEntityWithK8sTemplate(data models.CatalogEntityWithK8sTemplate,
	request models.CatalogEntityWithK8sTemplateRequest) (models.CatalogEntityWithK8sTemplate, int, error) {

	service, status, err := s.catalogApi.GetService(data.Instance.ClassId)
	if err != nil {
		logger.Error("catalogApi.GetService failed: ", err)
		return data, status, err
	}
	data.Service = &service

	instancePlanId := catalogModels.GetValueFromMetadata(data.Instance.Metadata, catalogModels.OFFERING_PLAN_ID)
	for _, plan := range service.Plans {
		if plan.Id == instancePlanId {
			data.ServicePlan = &plan
			break
		}
	}
	if data.ServicePlan == nil {
		return data, http.StatusBadRequest, errors.New("Plan is missing!")
	}

	additionalReplacements := prepareAdditionalReplacements(data, request)
	template, status, err := s.templateApi.GenerateParsedTemplate(service.TemplateId, request.InstanceId,
		data.ServicePlan.Name, additionalReplacements)
	if err != nil {
		logger.Error("templateApi.GenerateParsedTemplate failed: ", err)
		return data, status, err
	}
	data.Template = &template
	return data, http.StatusOK, nil
}

func (s *ServiceProcessorManager) getApplicationCatalogEntityWithK8sTemplate(data models.CatalogEntityWithK8sTemplate,
	request models.CatalogEntityWithK8sTemplateRequest) (models.CatalogEntityWithK8sTemplate, int, error) {

	app, status, err := s.catalogApi.GetApplication(data.Instance.ClassId)
	if err != nil {
		logger.Error("catalogApi.GetApplication failed: ", err)
		return data, status, err
	}
	data.Application = &app

	img, status, err := s.catalogApi.GetImage(app.ImageId)
	if err != nil {
		logger.Error("catalogApi.GetImage failed: ", err)
		return data, status, err
	}

	additionalReplacements := prepareAdditionalReplacements(data, request)
	template, status, err := s.templateApi.GenerateParsedTemplate(app.TemplateId, request.InstanceId,
		templateRepositoryModel.EMPTY_PLAN_NAME, additionalReplacements)
	if err != nil {
		logger.Error("templateApi.GenerateParsedTemplate failed: ", err)
		return data, status, err
	}

	addAdditionalEnvsAndSetReplicasToApplicationTemplate(&template, &img, app.Replication)
	data.Template = &template
	return data, http.StatusOK, nil
}

func prepareAdditionalReplacements(data models.CatalogEntityWithK8sTemplate, request models.CatalogEntityWithK8sTemplateRequest) map[string]string {
	additionalReplacements := make(map[string]string)
	additionalReplacements[templateRepositoryModel.PlaceholderHostname] = data.Instance.Name
	additionalReplacements[templateRepositoryModel.PlaceholderInstanceName] = data.Instance.Name

	if image := catalogModels.GetValueFromMetadata(data.Instance.Metadata, catalogModels.APPLICATION_IMAGE_ADDRESS); image != "" {
		additionalReplacements[templateRepositoryModel.PlaceholderImage] = image
	}

	if request.BoundInstanceId != "" {
		additionalReplacements[templateRepositoryModel.PlaceholderBoundInstanceID] = request.BoundInstanceId
	}

	if request.ApplicationSSLCertHash != "" {
		additionalReplacements[templateRepositoryModel.PlaceholderCertificateHash] = request.ApplicationSSLCertHash
	}

	if data.Service != nil {
		sourceOfferingIdValue := catalogModels.GetValueFromMetadata(data.Service.Metadata, templateRepositoryModel.PlaceholderSourceOfferingID)
		if sourceOfferingIdValue == "" {
			additionalReplacements[templateRepositoryModel.PlaceholderOfferingID] = data.Service.Id
		} else {
			additionalReplacements[templateRepositoryModel.PlaceholderOfferingID] = sourceOfferingIdValue
		}

		sourcePlanIdValue := catalogModels.GetValueFromMetadata(data.Service.Metadata, templateRepositoryModel.GetPrefixedSourcePlanName(data.ServicePlan.Name))
		if sourcePlanIdValue == "" {
			additionalReplacements[templateRepositoryModel.PlaceholderPlanID] = data.ServicePlan.Id
		} else {
			additionalReplacements[templateRepositoryModel.PlaceholderPlanID] = sourcePlanIdValue
		}

		brokerShortInstanceIdValue := catalogModels.GetValueFromMetadata(data.Service.Metadata, templateRepositoryModel.PlaceholderBrokerShortInstanceID)
		if brokerShortInstanceIdValue != "" {
			additionalReplacements[templateRepositoryModel.PlaceholderBrokerShortInstanceID] = brokerShortInstanceIdValue
		}

		brokerInstanceIdValue := catalogModels.GetValueFromMetadata(data.Service.Metadata, templateRepositoryModel.PlaceholderBrokerInstanceID)
		if brokerInstanceIdValue != "" {
			additionalReplacements[templateRepositoryModel.PlaceholderBrokerInstanceID] = brokerInstanceIdValue
		}
	}

	if len(data.Instance.Metadata) > 0 {
		metadata := GetMetadataAsMap(data.Instance.Metadata)
		metadataByte, _ := json.Marshal(metadata)
		additionalReplacements[templateRepositoryModel.PlaceholderExtraEnvs] = string(metadataByte)
	}

	additionalReplacements[templateRepositoryModel.PlaceholderCreatedBy] = data.Instance.AuditTrail.CreatedBy

	memoryLimit := util.GetEnvValueOrDefault("MEMORY_LIMIT", "0")
	additionalReplacements[templateRepositoryModel.PlaceholderMemoryLimit] = memoryLimit

	return additionalReplacements
}
