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
	"errors"
	"fmt"

	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func (s *ServiceEngineManager) createSSLLayerForApplication(data models.CatalogEntityWithK8sTemplate) error {
	if data.Template == nil || !isThereAnyDeployment(data.Template.Body) ||
		len(data.Template.Body[0].Deployments) == 0 ||
		len(data.Template.Body[0].Services) == 0 ||
		data.Instance == nil {
		return errors.New("data has nil pointers!")
	}

	idxAndShortInstanceID := data.Template.Body[0].Deployments[0].ObjectMeta.Labels[templateModel.PlaceholderIdxAndShortInstanceID]
	serviceName := data.Template.Body[0].Services[0].ObjectMeta.Name

	certkeySecretName := idxAndShortInstanceID + "-certkey"

	certkey, err := s.cAApi.GetCertKey(serviceName)
	if err != nil {
		return errors.New("getting certkey failed: " + err.Error())
	}

	secretData := make(map[string][]byte)
	secretData["ssl.crt"] = []byte(certkey.CertificateContent)
	secretData["ssl.key"] = []byte(certkey.KeyContent)
	certKeySecret := api.Secret{
		ObjectMeta: api.ObjectMeta{
			Name:   certkeySecretName,
			Labels: k8s.GetK8sTapLabels(data.Instance.Id),
		},
		Data: secretData,
	}

	err = s.kubernetesApi.CreateSecret(certKeySecret)
	if err != nil {
		return errors.New("creating secret failed: " + err.Error())
	}

	return nil
}

func (s *ServiceEngineManager) getCACertHash() (string, error) {
	caCertificate, err := s.cAApi.GetCa()
	if err != nil {
		return "", errors.New(fmt.Sprint("failed to get CA cert from ca-service! error: ", err.Error()))
	}
	return caCertificate.Hash, nil
}
