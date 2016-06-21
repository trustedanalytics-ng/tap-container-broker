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

package main

import (
	"os"
	"sync"

	"github.com/gocraft/web"

	caApi "github.com/trustedanalytics-ng/tap-ca/client"
	catalogApi "github.com/trustedanalytics-ng/tap-catalog/client"
	cephBrokerApi "github.com/trustedanalytics-ng/tap-ceph-broker/client"
	"github.com/trustedanalytics-ng/tap-container-broker/adapter"
	"github.com/trustedanalytics-ng/tap-container-broker/api"
	"github.com/trustedanalytics-ng/tap-container-broker/engine"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	httpGoCommon "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-go-common/util"
	templateRepositoryApi "github.com/trustedanalytics-ng/tap-template-repository/client"
)

var logger, _ = commonLogger.InitLogger("main")
var waitGroup = &sync.WaitGroup{}

const (
	getLogsTailLinesEnvVarName = "LOGS_TAIL_LINES"
	defaultGetLogsTailLines    = uint32(500)
	k8sAPIAddressEnvVarName    = "K8S_API_ADDRESS"
	k8sAPIUsernameEnvVarName   = "K8S_API_USERNAME"
	k8sAPIPasswordEnvVarName   = "K8S_API_PASSWORD"
)

func main() {
	initServices()
	go util.TerminationObserver(waitGroup, "Container-broker")
	go adapter.StartConsumer(waitGroup)

	context := api.Context{}
	r := web.New(context)
	r.Middleware(web.LoggerMiddleware)
	r.Get("/healthz", context.GetContainerBrokerHealth)

	basicAuthRouter := r.Subrouter(context, "/api/v1")
	route(basicAuthRouter)

	// for testing purpose, where v1 is current version
	v1AliasRouter := r.Subrouter(context, "/api/v1.0")
	route(v1AliasRouter)

	httpGoCommon.StartServer(r)
}

func route(router *web.Router) {
	router.Middleware((*api.Context).CheckEngineService)
	router.Middleware((*api.Context).BasicAuthorizeMiddleware)

	router.Get("/service/:instanceId/logs", (*api.Context).GetInstanceLogs)
	router.Get("/service/:instanceId/credentials", (*api.Context).GetCredentials)
	router.Get("/service/:instanceId/hosts", (*api.Context).GetInstanceHosts)
	router.Post("/service/:instanceId/expose", (*api.Context).Expose)
	router.Delete("/service/:instanceId/expose", (*api.Context).Hide)

	router.Get("/secret/:secretName", (*api.Context).GetSecret)
	router.Get("/configmap/:configMapName", (*api.Context).GetConfigMap)

	router.Post("/bind/:srcInstanceId/:dstInstanceId", (*api.Context).Bind)
	router.Post("/unbind/:srcInstanceId/:dstInstanceId", (*api.Context).Unbind)

	router.Get("/deployment/core/versions", (*api.Context).GetVersions)
}

func initServices() {
	templateRepositoryConnector, err := getTemplateRepositoryConnector()
	if err != nil {
		logger.Fatal("Can't connect with TAP-template-repository!", err)
	}

	catalogConnector, err := getCatalogConnector()
	if err != nil {
		logger.Fatal("Can't connect with TAP-catalog!", err)
	}

	cephBrokerConnector, err := getCephBrokerConnector()
	if err != nil {
		logger.Fatal("Can't connect with TAP-ceph-broker!", err)
	}

	caConnector, err := getCAConnector()
	if err != nil {
		logger.Fatal("Can't connect with TAP-CA!", err)
	}

	logsTailLinesUint32, err := util.GetUint32EnvValueOrDefault(getLogsTailLinesEnvVarName, defaultGetLogsTailLines)
	if err != nil {
		logger.Warningf("error parsing %v var, using default one: %v, err: %v", getLogsTailLinesEnvVarName, defaultGetLogsTailLines, err)
	}

	logsTailLinesUint := uint(logsTailLinesUint32)

	kubernetesApiConnector, err := k8s.GetNewK8FabricatorInstance(
		k8s.K8sClusterCredentials{
			Server:   os.Getenv(k8sAPIAddressEnvVarName),
			Username: os.Getenv(k8sAPIUsernameEnvVarName),
			Password: os.Getenv(k8sAPIPasswordEnvVarName),
		},
		cephBrokerConnector,
		k8s.K8FabricatorOptions{GetLogsTailLines: logsTailLinesUint})
	if err != nil {
		logger.Fatal("Can't connect with Kubernetes!", err)
	}

	engine := engine.GetNewServiceEngineManager(kubernetesApiConnector, templateRepositoryConnector, catalogConnector,
		caConnector, cephBrokerConnector, nil)
	api.ServiceEngine = engine
	adapter.ServiceEngine = engine
}

func getTemplateRepositoryConnector() (*templateRepositoryApi.TemplateRepositoryConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("TEMPLATE_REPOSITORY")
	if err != nil {
		return nil, err
	}
	return templateRepositoryApi.NewTemplateRepositoryBasicAuth("https://"+address, username, password)
}

func getCatalogConnector() (catalogApi.TapCatalogApi, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("CATALOG")
	if err != nil {
		return nil, err
	}
	return catalogApi.NewTapCatalogApiWithBasicAuth("https://"+address, username, password)
}

func getCephBrokerConnector() (*cephBrokerApi.CephBrokerConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("CEPH_BROKER")
	if err != nil {
		return nil, err
	}
	return cephBrokerApi.NewCephBrokerBasicAuth("https://"+address, username, password)
}

func getCAConnector() (*caApi.TapCaApiConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("CA")
	if err != nil {
		return nil, err
	}
	return caApi.NewTapCaApiConnector("http://"+address, username, password)
}
