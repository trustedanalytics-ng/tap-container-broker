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
	"fmt"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	exp                       = "EXP_"
	exp_mount_config_map_name = exp + "MOUNT_CONFIG_MAP_NAME_"
	exp_mount_path            = exp + "MOUNT_PATH_"
	exp_mount_read_only       = exp + "MOUNT_READ_ONLY_"

	VCAP_PARAM_NAME          = "VCAP"
	VCAP_SERVICES_PARAM_NAME = "VCAP_SERVICES"
	METADATA_CREDENTIALS     = "CREDENTIALS"
)

type VcapServices map[string][]VcapEntity

type VcapEntity struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Tags        []string               `json:"tags"`
	Plan        string                 `json:"plan"`
	Credentials map[string]interface{} `json:"credentials"`
}

type VolumeDisk struct {
	Name          string
	Path          string
	ConfigMapName string
	IsReadOnly    bool
}

func updateEnvByBindResponse(containerEnv []api.EnvVar, data map[string]string, serviceName, instanceName string, isBinding bool) ([]api.EnvVar, error) {
	if len(data) == 0 {
		return containerEnv, nil
	}

	if vcapValue, ok := data[VCAP_PARAM_NAME]; ok {
		vcapEntity := VcapEntity{}
		if err := json.Unmarshal([]byte(vcapValue), &vcapEntity); err != nil {
			return nil, err
		}
		return updateContainerEnvVarsByVcapEntity(containerEnv, vcapEntity, serviceName, isBinding)
	}

	if isBinding {
		for name, value := range data {
			newEnvVar := api.EnvVar{Name: getPrefixedEnvName(instanceName, name), Value: value}
			containerEnv = append(containerEnv, newEnvVar)
		}
	} else {
		for _, env := range containerEnv {
			if !strings.Contains(env.Name, getPrefixedEnvName(instanceName, "")) {
				containerEnv = append(containerEnv, env)
			}
		}
	}
	return containerEnv, nil
}

func updateContainerEnvVarsByVcapEntity(containerEnv []api.EnvVar, vcapEntity VcapEntity, serviceName string, isBinding bool) ([]api.EnvVar, error) {
	var err error
	for i, env := range containerEnv {
		if env.Name == VCAP_SERVICES_PARAM_NAME {
			containerEnv[i], err = updateEnvVarByVcapEntity(env, vcapEntity, serviceName, isBinding)
			return containerEnv, err
		}
	}

	// no VCAP_SERVICES in container envs - let's add it:
	valueByte, err := json.Marshal(VcapServices{serviceName: []VcapEntity{vcapEntity}})
	if err != nil {
		logger.Error("Can't Marshal into VcapServices", err)
		return nil, err
	}
	containerEnv = append(containerEnv, api.EnvVar{Name: VCAP_SERVICES_PARAM_NAME, Value: string(valueByte)})
	return containerEnv, nil
}

func updateEnvVarByVcapEntity(env api.EnvVar, vcapEntity VcapEntity, serviceName string, isBinding bool) (api.EnvVar, error) {
	value := VcapServices{}
	if err := json.Unmarshal([]byte(env.Value), &value); err != nil {
		logger.Error("Can't Unmarshal into VcapServices", err)
		return env, err
	}

	if isBinding {
		if service, ok := value[serviceName]; ok {
			value[serviceName] = append(service, vcapEntity)
		} else {
			value[serviceName] = []VcapEntity{vcapEntity}
		}
	} else if entities, ok := value[serviceName]; ok {
		updatedEntities := []VcapEntity{}
		for _, entity := range entities {
			if entity.Name != vcapEntity.Name {
				updatedEntities = append(updatedEntities, entity)
			}
		}
		if len(updatedEntities) == 0 {
			delete(value, serviceName)
		} else {
			value[serviceName] = updatedEntities
		}
	}

	updatedValue, err := json.Marshal(value)
	if err != nil {
		logger.Error("Can't Marshal into VcapServices", err)
		return env, err
	}
	env.Value = string(updatedValue)
	return env, nil
}

func getEnvVarsAndVolumeDisksFromComponentsDeployments(components []templateModel.KubernetesComponent, instanceName string) ([]api.EnvVar, []VolumeDisk, error) {
	result := []api.EnvVar{}
	expEnvVars := []api.EnvVar{}

	for _, component := range components {
		for _, deployment := range component.Deployments {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				for _, env := range container.Env {
					if strings.HasPrefix(env.Name, exp) {
						expEnvVars = append(expEnvVars, env)
					}

					env.Name = getPrefixedEnvName(instanceName, env.Name)
					result = append(result, env)
				}
			}
		}
	}

	volumeDisks, err := getVolumeDisksFromExpEnvVars(expEnvVars)
	return result, volumeDisks, err
}

func getVolumeDisksFromExpEnvVars(expEnvVars []api.EnvVar) ([]VolumeDisk, error) {
	result := []VolumeDisk{}
	for _, env := range expEnvVars {
		if strings.HasPrefix(env.Name, exp_mount_config_map_name) {
			name := strings.TrimPrefix(env.Name, exp_mount_config_map_name)
			logger.Debugf("VolumeDisk found: %s", name)
			volumeDisk := VolumeDisk{
				Name:          name,
				ConfigMapName: env.Value,
				IsReadOnly:    true,
			}

			pathEnvVar := getEnvVarByName(exp_mount_path+name, expEnvVars)
			if pathEnvVar == nil {
				return result, fmt.Errorf("there is no required env: %s", exp_mount_path+name)
			}
			volumeDisk.Path = pathEnvVar.Value

			readOnlyEnvVar := getEnvVarByName(exp_mount_read_only+name, expEnvVars)
			if readOnlyEnvVar != nil {
				isReadOnly, err := strconv.ParseBool(readOnlyEnvVar.Value)
				if err != nil {

				}
				volumeDisk.IsReadOnly = isReadOnly
			}

			result = append(result, volumeDisk)
		}
	}
	return result, nil
}

func getVolumeFromVolumeDisk(srcInstanceName string, volumeDisk VolumeDisk) api.Volume {
	return api.Volume{
		Name: getVolumeName(srcInstanceName, volumeDisk.Name),
		VolumeSource: api.VolumeSource{
			ConfigMap: &api.ConfigMapVolumeSource{
				LocalObjectReference: api.LocalObjectReference{
					Name: volumeDisk.ConfigMapName,
				},
			},
		},
	}
}

func getVolumeMountFromVolumeDisk(srcInstanceName string, volumeDisk VolumeDisk) api.VolumeMount {
	return api.VolumeMount{
		Name:      getVolumeName(srcInstanceName, volumeDisk.Name),
		ReadOnly:  volumeDisk.IsReadOnly,
		MountPath: volumeDisk.Path,
	}
}

func getVolumeName(srcInstanceName, envVarName string) string {
	return fmt.Sprintf("%s%s-volume", getVolumeNamePrefix(srcInstanceName), strings.ToLower(envVarName))
}

func getVolumeNamePrefix(srcInstanceName string) string {
	return fmt.Sprintf("%s-configmap-", srcInstanceName)
}

func getEnvVarByName(name string, envVars []api.EnvVar) *api.EnvVar {
	for _, env := range envVars {
		if env.Name == name {
			return &env
		}
	}
	return nil
}

func getPrefixedEnvName(prefix, envName string) string {
	return strings.ToUpper(k8s.ConvertToProperEnvName(prefix) + "_" + envName)
}

func addToBoundServicesRegistry(envs []api.EnvVar, offeringName, serviceName string) []api.EnvVar {
	keyIndex := getRegistryKeyIndex(envs, offeringName)
	if keyIndex == -1 {
		registry := api.EnvVar{Name: getRegistryKey(offeringName), Value: serviceName}
		return append(envs, registry)
	}

	envs[keyIndex].Value = addToSeparatedStrings(envs[keyIndex].Value, serviceName, ",")
	return envs
}

func removeFromBoundServicesRegistry(envs []api.EnvVar, offeringName, serviceName string) []api.EnvVar {
	keyIndex := getRegistryKeyIndex(envs, offeringName)
	if keyIndex == -1 {
		logger.Warningf("Trial of removing not existing entry from bound services registry of service %s of offering %s", serviceName, offeringName)
		return envs
	}

	envs[keyIndex].Value = deleteFromSeparatedStrings(envs[keyIndex].Value, serviceName, ",")
	if envs[keyIndex].Value == "" {
		return append(envs[:keyIndex], envs[keyIndex+1:]...)
	}
	return envs
}

func getRegistryKeyIndex(envs []api.EnvVar, offering string) int {
	registryKey := getRegistryKey(offering)

	for i := range envs {
		if envs[i].Name == registryKey {
			return i
		}
	}
	return -1
}

func getRegistryKey(offeringName string) string {
	return fmt.Sprintf("SERVICES_BOUND_%s", strings.ToUpper(k8s.ConvertToProperEnvName(offeringName)))
}

func addToSeparatedStrings(separatedStrings, newString, sep string) string {
	list := strings.Split(separatedStrings, sep)

	for _, v := range list {
		if v == newString {
			return separatedStrings
		}
	}

	list = append(list, newString)
	return strings.Join(list, sep)
}

func deleteFromSeparatedStrings(separatedStrings, toDelete, sep string) string {
	list := strings.Split(separatedStrings, sep)
	for i, v := range list {
		if v == toDelete {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return strings.Join(list, sep)
}
