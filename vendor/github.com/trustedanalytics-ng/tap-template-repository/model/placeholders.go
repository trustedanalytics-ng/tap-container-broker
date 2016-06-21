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
package model

import (
	"os"

	"github.com/trustedanalytics-ng/tap-go-common/util"
)

const (
	PlaceholderOrg   = "org"
	PlaceholderSpace = "space"

	PlaceholderDomainName        = "domain_name"
	PlaceholderImage             = "image"
	PlaceholderHostname          = "hostname"
	PlaceholderExtraEnvs         = "extra_envs"
	PlaceholderMemoryLimit       = "memory_limit"
	PlaceholderNginxSSLImageName = "nginx_ssl_name"

	PlaceholderInstanceName = "instance_name"
	PlaceholderInstanceID   = "instance_id"

	//TODO this is obsolete and will be removed soon -> DPNG-12250
	PlaceholderIdxAndShortInstanceID = "idx_and_short_instance_id"
	PlaceholderShortInstanceID       = "short_instance_id"
	PlaceholderBoundInstanceID       = "bound_instance_id"

	PlaceholderBrokerShortInstanceID = "broker_short_instance_id"
	PlaceholderBrokerInstanceID      = "broker_instance_id"

	PlaceholderRandom    = "random"
	PlaceholderRandomDNS = "random_dns"

	PlaceholderOfferingID         = "offering_id"
	PlaceholderPlanID             = "plan_id"
	PlaceholderSourceOfferingID   = "source_offering_id"
	PlaceholderSourcePlanIDPrefix = "source_plan_id-"

	PlaceholderCertificateHash = "cert_hash"

	PlaceholderUseExternalSslFlag = "use_external_ssl"

	PlaceholderRepositoryUri = "repository_uri"
	PlaceholderTapVersion    = "tap_version"
	PlaceholderProtocol = "tap_protocol"

	PlaceholderCreatedBy = "created_by"

	defaultOrg             = "00000000-0000-0000-0000-000000000000"
	defaultMemoryLimit     = "1Gi"
	defaultSpace           = "defaultSpace"
	defaultRepositoryUri   = "127.0.0.1:30000"
	defaultTapVersion      = "latest"
	defaultExternalSsl     = "false"
	defaultProtocol        = "http"
)

func GetPlaceholderWithDollarPrefix(placeholder string) string {
	return "$" + placeholder
}

func GetPrefixedSourcePlanName(planName string) string {
	return PlaceholderSourcePlanIDPrefix + planName
}

func GetProtocol(sslFlag string) string {
	protocol := defaultProtocol
	if sslFlag == "true" {
		protocol = "https"
	}
	return protocol
}

func getDefaultReplacements() map[string]string {
	return map[string]string{
		GetPlaceholderWithDollarPrefix(PlaceholderDomainName):            os.Getenv("DOMAIN"),
		GetPlaceholderWithDollarPrefix(PlaceholderNginxSSLImageName):     os.Getenv("NGINX_SSL_IMAGE_NAME"),
		GetPlaceholderWithDollarPrefix(PlaceholderOrg):                   util.GetEnvValueOrDefault("CORE_ORGANIZATION_UUID", defaultOrg),
		GetPlaceholderWithDollarPrefix(PlaceholderRepositoryUri):         util.GetEnvValueOrDefault("REPOSITORY_URI",defaultRepositoryUri),
		GetPlaceholderWithDollarPrefix(PlaceholderTapVersion):            util.GetEnvValueOrDefault("TAP_VERSION", defaultTapVersion),
		GetPlaceholderWithDollarPrefix(PlaceholderUseExternalSslFlag):    util.GetEnvValueOrDefault("USE_EXTERNAL_SSL", defaultExternalSsl),
		GetPlaceholderWithDollarPrefix(PlaceholderProtocol):           	  GetProtocol(util.GetEnvValueOrDefault("USE_EXTERNAL_SSL", defaultExternalSsl)),
		GetPlaceholderWithDollarPrefix(PlaceholderSpace):                 defaultSpace,
		GetPlaceholderWithDollarPrefix(PlaceholderMemoryLimit):           defaultMemoryLimit,
	}
}

func GetMapWithDefaultReplacementsIfKeyNotExists(originalMap map[string]string) map[string]string {
	defaults := getDefaultReplacements()
	for key, value := range defaults {
		if _, ok := originalMap[key]; !ok {
			originalMap[key] = value
		}
	}
	return originalMap
}