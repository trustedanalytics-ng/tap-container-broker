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
	"net/http"
	"os"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/client/restclient"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

type KubernetesTestCreator struct {
	testClient          *testclient.Fake
	testExtensionClient *testclient.FakeExperimental
}

var logger, _ = commonLogger.InitLogger("k8s")

type K8sClusterCredentials struct {
	CLusterName    string `json:"cluster_name"`
	Server         string `json:"api_server"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	CaCert         string `json:"ca_cert"`
	AdminKey       string `json:"admin_key"`
	AdminCert      string `json:"admin_cert"`
	ConsulEndpoint string `json:"consul_http_api"`
}

func GetNewClient(creds K8sClusterCredentials) (k8sClient.Interface, error) {
	if creds.Server == "" {
		// get default K8s api client from same cluster as pod's
		return k8sClient.NewInCluster()
	} else {
		config, err := getKubernetesConfig(creds)
		if err != nil {
			return nil, err
		}
		return k8sClient.New(config)
	}
}

func getKubernetesConfig(creds K8sClusterCredentials) (*restclient.Config, error) {
	sslActive, parseError := strconv.ParseBool(os.Getenv("KUBE_SSL_ACTIVE"))
	if parseError != nil {
		logger.Error("KUBE_SSL_ACTIVE env probably not set!")
		return nil, parseError
	}

	var transport *http.Transport
	var err error

	if sslActive {
		_, transport, err = brokerHttp.GetHttpClientWithCertAndCa(creds.AdminCert, creds.AdminKey, creds.CaCert)
	} else {
		_, transport, err = brokerHttp.GetHttpClient()
	}

	if err != nil {
		return nil, err
	}

	config := &restclient.Config{
		Host:      creds.Server,
		Username:  creds.Username,
		Password:  creds.Password,
		Transport: transport,
	}
	return config, nil
}

func (k *KubernetesTestCreator) GetNewClient(creds K8sClusterCredentials) (k8sClient.Interface, error) {
	return k.testClient, nil
}

func (k *KubernetesTestCreator) GetNewExtensionsClient(creds K8sClusterCredentials) (k8sClient.Interface, error) {
	return k.testExtensionClient, nil
}

/*
	Objects will be returned in provided order
	All objects should do same action e.g. list/update/create
*/
func (k *KubernetesTestCreator) LoadSimpleResponsesWithSameAction(responeObjects ...runtime.Object) {
	k.testClient = testclient.NewSimpleFake(responeObjects...)
}

func (k *KubernetesTestCreator) LoadSimpleResponsesWithSameActionForExtensionsClient(responeObjects ...runtime.Object) {
	k.testExtensionClient = testclient.NewSimpleFakeExp(responeObjects...)
}

type KubernetesTestAdvancedParams struct {
	Verb            string
	Resource        string
	ResponceObjects []runtime.Object
}

/*
	This method allow to inject response object dependly of their action
*/
func (k *KubernetesTestCreator) LoadAdvancedResponses(params []KubernetesTestAdvancedParams) {
	fakeClient := &testclient.Fake{}

	for _, param := range params {
		o := testclient.NewObjects(api.Scheme, api.Codecs.UniversalDecoder())
		for _, obj := range param.ResponceObjects {
			if err := o.Add(obj); err != nil {
				panic(err)
			}
		}
		verb := param.Verb
		if param.Verb == "" {
			verb = "*"
		}

		resource := param.Resource
		if param.Resource == "" {
			resource = "*"
		}
		fakeClient.AddReactor(verb, resource, testclient.ObjectReaction(o, registered.RESTMapper()))
	}

	fakeClient.AddWatchReactor("*", testclient.DefaultWatchReactor(watch.NewFake(), nil))
	k.testClient = fakeClient
}
