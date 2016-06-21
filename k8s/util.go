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
	"net/url"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/trustedanalytics-ng/tap-go-common/util"
)

func ConvertToProperEnvName(key string) string {
	return strings.Replace(key, "-", "_", -1)
}

func getCephImageSize(sizeValue string) uint64 {
	value, errDefault := util.GetUint64EnvValueOrDefault("CEPH_IMAGE_SIZE_MB", defaultCephImageSizeMB)
	if sizeValue != "" {
		size, err := strconv.ParseUint(sizeValue, 0, 64)
		if err != nil {
			logger.Errorf("Can't parse value: %s ", sizeValue)
		}

		if errDefault != nil {
			logger.Errorf("Can't parse value CEPH_IMAGE_SIZE_MB env - default limit: %d will be used", value)
			return value
		}
		return size
	}
	return value
}

func addProtocolToHost(annotations map[string]string, host string) string {
	protocol := "http"

	if val, ok := annotations[useExternalSslFlag]; ok {
		useSsl, err := strconv.ParseBool(val)
		if err != nil {
			logger.Warningf("%s Ingress annotation can not be parsed! Default http protocol will be used! Cause: %v", useExternalSslFlag, err)
		} else if useSsl {
			protocol = "https"
		}
	}

	parsedHost, err := url.Parse(host)
	if err != nil {
		logger.Errorf("Cannot pars host: %s - unchanged value will be returned", host)
		return host
	}

	if parsedHost.Scheme == "" {
		parsedHost.Scheme = protocol
	}
	return parsedHost.String()
}

func appendSourceEnvsToDestinationEnvsIfNotContained(sourceEnvVars, destinationEnvVars []api.EnvVar) []api.EnvVar {
	result := destinationEnvVars
	for _, sourceEnvVar := range sourceEnvVars {
		if !isEnvContained(sourceEnvVar.Name, destinationEnvVars) {
			result = append(result, sourceEnvVar)
		}
	}
	return result
}

func isEnvContained(keyName string, envs []api.EnvVar) bool {
	for _, envVar := range envs {
		if keyName == envVar.Name {
			return true
		}
	}
	return false
}

func FetchHostsFromIngress(ingress extensions.Ingress) []string {
	result := make([]string, len(ingress.Spec.Rules))
	for i, rule := range ingress.Spec.Rules {
		host := addProtocolToHost(ingress.ObjectMeta.Annotations, rule.Host)
		result[i] = host
	}
	return result
}
