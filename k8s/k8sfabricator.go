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
	"errors"
	"fmt"
	"sort"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	clientK8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/trustedanalytics-ng/tap-ceph-broker/client"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	"github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	NotFound      string = "not found"
	AlreadyExists string = "already exists"

	InstanceIdLabel string = "instance_id"
	managedByLabel  string = "managed_by"
	managedByValue  string = "TAP"

	jobName            string = "job-name"
	useExternalSslFlag string = "useExternalSsl"

	defaultCephImageSizeMB = 200
)

type KubernetesApi interface {
	FabricateComponents(instanceId string, shouldOverwriteEnvs bool, parameters map[string]string, components []model.KubernetesComponent) error
	GetFabricatedComponentsForAllOrgs() ([]*model.KubernetesComponent, error)
	GetFabricatedComponents(organization string) ([]*model.KubernetesComponent, error)
	DeleteAllByInstanceId(instanceId string) error
	DeleteAllPersistentVolumeClaims() error
	GetAllPersistentVolumes() ([]api.PersistentVolume, error)
	GetDeploymentsEnvsByInstanceId(instanceId string) ([]models.ContainerCredenials, error)
	GetService(name string) (*api.Service, error)
	CreateService(service api.Service) error
	DeleteService(name string) error
	GetServiceByInstanceId(instanceId string) ([]api.Service, error)
	GetServices() ([]api.Service, error)
	GetEndpoint(name string) (*api.Endpoints, error)
	CreateEndpoint(endpoint api.Endpoints) error
	DeleteEndpoint(name string) error
	GetPodsByInstanceId(instanceId string) ([]api.Pod, error)
	ListDeployments() (*extensions.DeploymentList, error)
	ListDeploymentsByLabel(labelKey, labelValue string) (*extensions.DeploymentList, error)
	CreateConfigMap(configMap *api.ConfigMap) error
	GetConfigMap(name string) (*api.ConfigMap, error)
	GetSecret(name string) (*api.Secret, error)
	CreateSecret(secret api.Secret) error
	DeleteSecret(name string) error
	GetIngress(name string) (*extensions.Ingress, error)
	CreateIngress(ingress extensions.Ingress) error
	DeleteIngress(name string) error
	GetJobs() (*batch.JobList, error)
	GetJobsByInstanceId(instanceId string) (*batch.JobList, error)
	DeleteJob(jobName string) error
	GetPod(name string) (*api.Pod, error)
	GetPodsBySpecifiedSelector(labelKey, labelValue string) (*api.PodList, error)
	CreatePod(pod api.Pod) error
	DeletePod(podName string) error
	GetSpecificPodLogs(pod api.Pod) (map[string]string, error)
	GetPodsLogs(instanceId string) (map[string]string, error)
	GetJobLogs(job batch.Job) (map[string]string, error)
	CreateJob(job *batch.Job, instanceId string) error
	ScaleDeploymentAndWait(deploymentName, instanceId string, replicas int) error
	UpdateDeployment(instanceId string, prepareFunction models.PrepareDeployment) error
	GetDeployment(name string) (*extensions.Deployment, error)
	GetIngressHosts(instanceId string) ([]string, error)
	GetPodEvents(pod api.Pod) ([]api.Event, error)
}

type K8Fabricator struct {
	client     clientK8s.Interface
	cephClient client.CephBroker
	options    K8FabricatorOptions
}

type K8FabricatorOptions struct {
	GetLogsTailLines uint
}

func GetNewK8FabricatorInstance(creds K8sClusterCredentials, cephClient client.CephBroker, options K8FabricatorOptions) (*K8Fabricator, error) {
	k8sClient, err := GetNewClient(creds)
	if err != nil {
		return nil, err
	}

	result := K8Fabricator{
		client:     k8sClient,
		cephClient: cephClient,
		options:    options,
	}
	return &result, err
}

func (k *K8Fabricator) FabricateComponents(instanceId string, shouldOverwriteEnvs bool, parameters map[string]string, components []model.KubernetesComponent) error {
	extraEnvironments := []api.EnvVar{{Name: "TAP_K8S", Value: "true"}}
	for key, value := range parameters {
		extraUserParam := api.EnvVar{
			Name:  ConvertToProperEnvName(key),
			Value: value,
		}
		extraEnvironments = append(extraEnvironments, extraUserParam)
	}
	logger.Debugf("Instance: %s extra parameters value: %v", instanceId, extraEnvironments)

	for _, component := range components {
		for _, sc := range component.ConfigMaps {
			if _, err := k.client.ConfigMaps(api.NamespaceDefault).Create(sc); err != nil {
				return err
			}
		}

		for _, sc := range component.Secrets {
			if _, err := k.client.Secrets(api.NamespaceDefault).Create(sc); err != nil {
				return err
			}
		}

		for _, claim := range component.PersistentVolumeClaims {
			if _, err := k.client.PersistentVolumeClaims(api.NamespaceDefault).Create(claim); err != nil {
				return err
			}
		}

		for _, svc := range component.Services {
			if _, err := k.client.Services(api.NamespaceDefault).Create(svc); err != nil {
				return err
			}
		}

		for _, acc := range component.ServiceAccounts {
			if _, err := k.client.ServiceAccounts(api.NamespaceDefault).Create(acc); err != nil {
				return err
			}
		}

		for _, deployment := range component.Deployments {
			for i, container := range deployment.Spec.Template.Spec.Containers {
				var updatedEnvVars []api.EnvVar
				if shouldOverwriteEnvs {
					updatedEnvVars = appendSourceEnvsToDestinationEnvsIfNotContained(container.Env, extraEnvironments)
				} else {
					updatedEnvVars = appendSourceEnvsToDestinationEnvsIfNotContained(extraEnvironments, container.Env)
				}
				deployment.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
			}

			cephVolume, err := processDeploymentVolumes(*deployment, k.cephClient, true)

			if err != nil {
				return err
			}

			// append only if volume not empty
			if cephVolume.Name != "" {
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, cephVolume)
			}

			if _, err := k.client.Extensions().Deployments(api.NamespaceDefault).Create(deployment); err != nil {
				return err
			}
		}

		for _, ing := range component.Ingresses {
			if _, err := k.client.Extensions().Ingress(api.NamespaceDefault).Create(ing); err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *K8Fabricator) GetFabricatedComponentsForAllOrgs() ([]*model.KubernetesComponent, error) {
	// TODO: iterate over all organizations here
	return k.GetFabricatedComponents(api.NamespaceDefault)
}

func (k *K8Fabricator) GetFabricatedComponents(organization string) ([]*model.KubernetesComponent, error) {
	selector, err := getSelectorForManagedByLabel(managedByLabel, managedByValue)
	if err != nil {
		return nil, err
	}
	listOptions := api.ListOptions{
		LabelSelector: selector,
	}

	pvcs, err := k.client.PersistentVolumeClaims(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	deployments, err := k.client.Extensions().Deployments(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	ings, err := k.client.Extensions().Ingress(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	svcs, err := k.client.Services(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	sAccounts, err := k.client.ServiceAccounts(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	secrets, err := k.client.Secrets(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	configMaps, err := k.client.ConfigMaps(organization).List(listOptions)
	if err != nil {
		return nil, err
	}
	return groupIntoComponents(
		pvcs.Items,
		deployments.Items,
		ings.Items,
		svcs.Items,
		sAccounts.Items,
		secrets.Items,
		configMaps.Items), nil
}

func groupIntoComponents(
	pvcs []api.PersistentVolumeClaim,
	deployments []extensions.Deployment,
	ings []extensions.Ingress,
	svcs []api.Service,
	sAccounts []api.ServiceAccount,
	secrets []api.Secret,
	configMaps []api.ConfigMap,
) []*model.KubernetesComponent {

	components := make(map[string]*model.KubernetesComponent)

	for i := range pvcs {
		pvc := pvcs[i]
		component := ensureAndGetKubernetesComponent(components, pvc.GetLabels())
		component.PersistentVolumeClaims = append(component.PersistentVolumeClaims, &pvc)
	}

	for i := range deployments {
		deployment := deployments[i]
		component := ensureAndGetKubernetesComponent(components, deployment.GetLabels())
		component.Deployments = append(component.Deployments, &deployment)
	}

	for i := range ings {
		ing := ings[i]
		component := ensureAndGetKubernetesComponent(components, ing.GetLabels())
		component.Ingresses = append(component.Ingresses, &ing)
	}

	for i := range svcs {
		svc := svcs[i]
		component := ensureAndGetKubernetesComponent(components, svc.GetLabels())
		component.Services = append(component.Services, &svc)
	}

	for i := range sAccounts {
		account := sAccounts[i]
		component := ensureAndGetKubernetesComponent(components, account.GetLabels())
		component.ServiceAccounts = append(component.ServiceAccounts, &account)
	}

	for i := range secrets {
		secret := secrets[i]
		component := ensureAndGetKubernetesComponent(components, secret.GetLabels())
		component.Secrets = append(component.Secrets, &secret)
	}

	for i := range configMaps {
		configMap := configMaps[i]
		component := ensureAndGetKubernetesComponent(components, configMap.GetLabels())
		component.ConfigMaps = append(component.ConfigMaps, &configMap)
	}

	var list []*model.KubernetesComponent
	for _, component := range components {
		list = append(list, component)
	}
	return list
}

func ensureAndGetKubernetesComponent(components map[string]*model.KubernetesComponent,
	labels map[string]string) *model.KubernetesComponent {

	id := labels[InstanceIdLabel]
	component, found := components[id]
	if !found {
		component = &model.KubernetesComponent{}
		components[id] = component
	}
	return component
}

func (k *K8Fabricator) CreateJob(job *batch.Job, instanceId string) error {
	logger.Debug("Creating Job. InstanceId:", instanceId)
	_, err := k.client.Extensions().Jobs(api.NamespaceDefault).Create(job)
	return err
}

func (k *K8Fabricator) GetJobs() (*batch.JobList, error) {
	selector, err := getSelectorForManagedByLabel(managedByLabel, managedByValue)
	if err != nil {
		return nil, err
	}

	return k.client.Extensions().Jobs(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
}

func (k *K8Fabricator) GetJobsByInstanceId(instanceId string) (*batch.JobList, error) {
	selector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return nil, err
	}

	return k.client.Extensions().Jobs(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
}

func (k *K8Fabricator) DeleteJob(jobName string) error {
	return k.client.Extensions().Jobs(api.NamespaceDefault).Delete(jobName, &api.DeleteOptions{})
}

func (k *K8Fabricator) GetJobLogs(job batch.Job) (map[string]string, error) {
	jobPodsSelector, err := getSelectorBySpecificJob(job)
	if err != nil {
		return nil, err
	}
	return k.getLogs(jobPodsSelector)
}

func (k *K8Fabricator) GetSpecificPodLogs(pod api.Pod) (map[string]string, error) {
	result := make(map[string]string)
	for _, container := range pod.Spec.Containers {
		byteBody, err := k.client.Pods(api.NamespaceDefault).GetLogs(pod.Name, &api.PodLogOptions{Container: container.Name}).Do().Raw()
		if err != nil {
			logger.Error(fmt.Sprintf("Can't get logs for pod: %s and container: %s", pod.Name, container.Name))
			return result, err
		}
		result[pod.Name+"-"+container.Name] = string(byteBody)
	}
	return result, nil
}

func (k *K8Fabricator) GetPodsLogs(instanceId string) (map[string]string, error) {
	podsSelector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return nil, err
	}
	return k.getLogs(podsSelector)
}

func (k *K8Fabricator) getLogs(selector labels.Selector) (map[string]string, error) {
	result := make(map[string]string)

	pods, err := k.client.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return result, nil
	}

	isThereAtLeastOneSuccessGetLogTry := false
	linesLimit := int64(k.options.GetLogsTailLines)
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			byteBody, err := k.client.Pods(api.NamespaceDefault).GetLogs(pod.Name, &api.PodLogOptions{
				Container: container.Name, TailLines: &linesLimit}).Do().Raw()
			if err != nil {
				logger.Error(fmt.Sprintf("Can't get logs for pod: %s and container: %s", pod.Name, container.Name))
			} else {
				isThereAtLeastOneSuccessGetLogTry = true
			}
			result[pod.Name+"-"+container.Name] = string(byteBody)
		}
	}
	if isThereAtLeastOneSuccessGetLogTry {
		return result, nil
	} else {
		return result, errors.New("Can't fetch logs, err:" + err.Error())
	}
}

func (k *K8Fabricator) GetIngressHosts(instanceId string) ([]string, error) {
	result := []string{}
	ingresSelector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return nil, err
	}

	ingresses, err := k.client.Extensions().Ingress(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: ingresSelector,
	})
	if err != nil {
		return nil, err
	}

	for _, ingress := range ingresses.Items {
		result = append(result, FetchHostsFromIngress(ingress)...)
	}
	return result, err
}

func (k *K8Fabricator) DeleteAllByInstanceId(instanceId string) error {
	selector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return err
	}

	accs, err := k.client.ServiceAccounts(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List service accounts failed:", err)
		return err
	}
	var name string
	for _, i := range accs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete service account:", name)
		err = k.client.ServiceAccounts(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete service account failed:", err)
			return err
		}
	}

	svcs, err := k.client.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List services failed:", err)
		return err
	}
	for _, i := range svcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete service:", name)
		err = k.client.Services(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete service failed:", err)
			return err
		}
	}

	ings, err := k.client.Extensions().Ingress(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List ingresses failed:", err)
		return err
	}
	for _, i := range ings.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete ingress:", name)
		err = k.client.Extensions().Ingress(api.NamespaceDefault).Delete(name, &api.DeleteOptions{})
		if err != nil {
			logger.Error("Delete ingress failed:", err)
			return err
		}
	}

	if err = k.DeleteAllDeployments(instanceId, selector); err != nil {
		logger.Error("Delete deployment failed:", err)
		return err
	}

	secrets, err := k.client.Secrets(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List secrets failed:", err)
		return err
	}
	for _, i := range secrets.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete secret:", name)
		err = k.client.Secrets(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete secret failed:", err)
			return err
		}
	}

	configMaps, err := k.client.ConfigMaps(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List configMaps failed:", err)
		return err
	}
	for _, i := range configMaps.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete configMap:", name)
		err = k.client.ConfigMaps(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete configMap failed:", err)
			return err
		}
	}

	pvcs, err := k.client.PersistentVolumeClaims(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List PersistentVolumeClaims failed:", err)
		return err
	}
	for _, i := range pvcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete PersistentVolumeClaims:", name)
		err = k.client.PersistentVolumeClaims(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete PersistentVolumeClaims failed:", err)
			return err
		}
	}

	endpoints, err := k.client.Endpoints(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("List endpoints failed:", err)
		return err
	}
	for _, i := range endpoints.Items {
		name = i.ObjectMeta.Name
		logger.Debug("Delete endpoint:", name)
		err = k.client.Endpoints(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete endpoint failed:", err)
			return err
		}
	}
	return nil
}

func (k *K8Fabricator) DeleteAllPersistentVolumeClaims() error {
	pvList, err := k.client.PersistentVolumeClaims(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: labels.NewSelector(),
	})
	if err != nil {
		logger.Error("List PersistentVolumeClaims failed:", err)
		return err
	}

	var errorFound bool = false
	for _, i := range pvList.Items {
		name := i.ObjectMeta.Name
		logger.Debug("Delete PersistentVolumeClaims:", name)
		err = k.client.PersistentVolumeClaims(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("Delete PersistentVolumeClaims: "+name+" failed!", err)
			errorFound = true
		}
	}

	if errorFound {
		return errors.New("Error on deleting PersistentVolumeClaims!")
	} else {
		return nil
	}
}

func (k *K8Fabricator) GetAllPersistentVolumes() ([]api.PersistentVolume, error) {
	pvList, err := k.client.PersistentVolumes().List(api.ListOptions{
		LabelSelector: labels.NewSelector(),
	})
	if err != nil {
		logger.Error("List PersistentVolume failed:", err)
		return nil, err
	}
	return pvList.Items, nil
}

func (k *K8Fabricator) GetServiceByInstanceId(instanceId string) ([]api.Service, error) {
	response := []api.Service{}

	selector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return response, err
	}

	serviceList, err := k.client.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("ListServices failed:", err)
		return response, err
	}

	return serviceList.Items, nil
}

func (k *K8Fabricator) GetService(name string) (*api.Service, error) {
	service, err := k.client.Services(api.NamespaceDefault).Get(name)
	if err != nil {
		return &api.Service{}, err
	}
	return service, nil
}

func (k *K8Fabricator) CreateService(service api.Service) error {
	_, err := k.client.Services(api.NamespaceDefault).Create(&service)
	return err
}

func (k *K8Fabricator) DeleteService(name string) error {
	err := k.client.Services(api.NamespaceDefault).Delete(name)
	return err
}

func (k *K8Fabricator) GetServices() ([]api.Service, error) {
	response := []api.Service{}

	selector, err := getSelectorForManagedByLabel(managedByLabel, managedByValue)
	if err != nil {
		logger.Error("GetSelectorForManagedByLabel error", err)
		return response, err
	}

	serviceList, err := k.client.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("ListServices failed:", err)
		return response, err
	}
	return serviceList.Items, nil
}

func (k *K8Fabricator) GetEndpoint(name string) (*api.Endpoints, error) {
	result, err := k.client.Endpoints(api.NamespaceDefault).Get(name)
	if err != nil {
		return &api.Endpoints{}, err
	}
	return result, nil
}

func (k *K8Fabricator) CreateEndpoint(endpoint api.Endpoints) error {
	_, err := k.client.Endpoints(api.NamespaceDefault).Create(&endpoint)
	return err
}

func (k *K8Fabricator) DeleteEndpoint(name string) error {
	err := k.client.Endpoints(api.NamespaceDefault).Delete(name)
	return err
}

func (k *K8Fabricator) GetPodsByInstanceId(instanceId string) ([]api.Pod, error) {
	result := []api.Pod{}
	selector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return result, err
	}

	pods, err := k.client.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	return pods.Items, err
}

type SortableEvents []api.Event

func (list SortableEvents) Len() int {
	return len(list)
}

func (list SortableEvents) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableEvents) Less(i, j int) bool {
	return list[i].LastTimestamp.Time.Before(list[j].LastTimestamp.Time)
}

func (k *K8Fabricator) GetPodEvents(pod api.Pod) ([]api.Event, error) {

	eventsList, err := k.client.Events(api.NamespaceDefault).Search(&pod)
	if err != nil {
		return []api.Event{}, err
	}

	var result SortableEvents = eventsList.Items
	sort.Sort(result)
	return result, nil
}

func (k *K8Fabricator) GetPod(name string) (*api.Pod, error) {
	result, err := k.client.Pods(api.NamespaceDefault).Get(name)
	if err != nil {
		return &api.Pod{}, err
	}
	return result, nil
}

func (k *K8Fabricator) GetPodsBySpecifiedSelector(labelKey, labelValue string) (*api.PodList, error) {
	selector, err := getCustomSelector(labelKey, labelValue)
	result, err := k.client.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return &api.PodList{}, err
	}
	return result, nil
}

func getCustomSelector(labelKey, labelValue string) (labels.Selector, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(labelKey, labels.EqualsOperator, sets.NewString(labelValue))
	if err != nil {
		return selector, err
	}
	return selector.Add(*managedByReq), nil
}

func (k *K8Fabricator) CreatePod(pod api.Pod) error {
	_, err := k.client.Pods(api.NamespaceDefault).Create(&pod)
	return err
}

func (k *K8Fabricator) DeletePod(podName string) error {
	return k.client.Pods(api.NamespaceDefault).Delete(podName, &api.DeleteOptions{})
}

type ServiceCredential struct {
	Name  string
	Host  string
	Ports []api.ServicePort
}

func (k *K8Fabricator) CreateConfigMap(configMap *api.ConfigMap) error {
	_, err := k.client.ConfigMaps(api.NamespaceDefault).Create(configMap)
	return err
}

func (k *K8Fabricator) GetConfigMap(name string) (*api.ConfigMap, error) {
	configMap, err := k.client.ConfigMaps(api.NamespaceDefault).Get(name)
	if err != nil {
		return &api.ConfigMap{}, err
	}
	return configMap, nil
}

func (k *K8Fabricator) GetSecret(name string) (*api.Secret, error) {
	result, err := k.client.Secrets(api.NamespaceDefault).Get(name)
	if err != nil {
		return &api.Secret{}, err
	}
	return result, nil
}

func (k *K8Fabricator) CreateSecret(secret api.Secret) error {
	_, err := k.client.Secrets(api.NamespaceDefault).Create(&secret)
	return err
}

func (k *K8Fabricator) DeleteSecret(name string) error {
	err := k.client.Secrets(api.NamespaceDefault).Delete(name)
	return err
}

func (k *K8Fabricator) GetIngress(name string) (*extensions.Ingress, error) {
	result, err := k.client.Extensions().Ingress(api.NamespaceDefault).Get(name)
	if err != nil {
		return &extensions.Ingress{}, err
	}
	return result, nil
}

func (k *K8Fabricator) CreateIngress(ingress extensions.Ingress) error {
	_, err := k.client.Extensions().Ingress(api.NamespaceDefault).Create(&ingress)
	return err
}

func (k *K8Fabricator) DeleteIngress(name string) error {
	err := k.client.Extensions().Ingress(api.NamespaceDefault).Delete(name, nil)
	return err
}

func (k *K8Fabricator) GetDeploymentsEnvsByInstanceId(instanceId string) ([]models.ContainerCredenials, error) {
	result := []models.ContainerCredenials{}

	selector, err := getSelectorForInstanceIdLabel(instanceId)
	if err != nil {
		return result, err
	}

	// this selector will be use to fetch all secrets/configMaps beacuse we need to find bound envs values
	selectorAll, err := getSelectorForManagedByLabel(managedByLabel, managedByValue)
	if err != nil {
		return result, err
	}

	deployments, err := k.listDeployments(selector)
	if err != nil {
		return result, err
	}

	if len(deployments.Items) < 1 {
		return result, errors.New(fmt.Sprintf("Deployments associated with the service %s are not found", instanceId))
	}

	secrets, err := k.client.Secrets(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selectorAll,
	})
	if err != nil {
		logger.Error("List secrets failed:", err)
		return result, err
	}

	configMaps, err := k.client.ConfigMaps(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selectorAll,
	})
	if err != nil {
		logger.Error("List configMaps failed:", err)
		return result, err
	}

	for _, deployment := range deployments.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerEnvs := models.ContainerCredenials{}
			containerEnvs.Name = deployment.Name + "_" + container.Name
			containerEnvs.Envs = make(map[string]interface{})

			for _, env := range container.Env {
				if env.Value != "" {
					containerEnvs.Envs[env.Name] = env.Value
				} else if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					containerEnvs.Envs[env.Name] = findSecretValue(secrets, env.ValueFrom.SecretKeyRef)
				} else if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
					containerEnvs.Envs[env.Name] = findConfigMapValue(configMaps, env.ValueFrom.ConfigMapKeyRef)
				}
			}
			result = append(result, containerEnvs)
		}
	}
	return result, nil
}

func findSecretValue(secrets *api.SecretList, selector *api.SecretKeySelector) string {
	for _, secret := range secrets.Items {
		if secret.Name != selector.Name {
			continue
		}

		for key, value := range secret.Data {
			if key == selector.Key {
				return string((value))
			}
		}
	}
	logger.Warningf("Key %s not found in Secret", selector.Key)
	return ""
}

func findConfigMapValue(configMaps *api.ConfigMapList, selector *api.ConfigMapKeySelector) string {
	for _, configMap := range configMaps.Items {
		if configMap.Name != selector.Name {
			continue
		}

		for key, value := range configMap.Data {
			if key == selector.Key {
				return string(value)
			}
		}
	}
	logger.Warningf("Key %s not found in ConfigMap", selector.Key)
	return ""
}

func getSelectorForInstanceIdLabel(instanceId string) (labels.Selector, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(managedByLabel, labels.EqualsOperator, sets.NewString(managedByValue))
	if err != nil {
		return selector, err
	}
	instanceIdReq, err := labels.NewRequirement(InstanceIdLabel, labels.EqualsOperator, sets.NewString(instanceId))
	if err != nil {
		return selector, err
	}
	return selector.Add(*managedByReq, *instanceIdReq), nil
}

func getSelectorForManagedByLabel(labelKey, labelValue string) (labels.Selector, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(labelKey, labels.EqualsOperator, sets.NewString(labelValue))
	if err != nil {
		return selector, err
	}
	return selector.Add(*managedByReq), nil
}

func getSelectorBySpecificJob(job batch.Job) (labels.Selector, error) {
	selector := labels.NewSelector()
	for k, v := range job.Labels {
		requirement, err := labels.NewRequirement(k, labels.EqualsOperator, sets.NewString(v))
		if err != nil {
			return selector, err
		}
		selector.Add(*requirement)
	}
	jobNameReq, err := labels.NewRequirement(jobName, labels.EqualsOperator, sets.NewString(job.Name))
	if err != nil {
		return selector, err
	}
	return selector.Add(*jobNameReq), nil
}
