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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	k8sAPI "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	caModels "github.com/trustedanalytics-ng/tap-ca/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/k8s"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestCreateSSLLayerForApplication(t *testing.T) {
	serviceName := "example-service"
	instanceID := "4ert534nt-435ng-dfgi5"
	idxAndShortInstanceID := "x" + instanceID

	fakeCertkeyResponse := caModels.CertKeyResponse{
		CertificateContent: "random-certificate",
		KeyContent:         "random-key",
	}

	fakeData := models.CatalogEntityWithK8sTemplate{
		Template: &templateModels.Template{
			Body: []templateModels.KubernetesComponent{{
				Deployments: []*extensions.Deployment{
					&extensions.Deployment{
						ObjectMeta: k8sAPI.ObjectMeta{
							Labels: make(map[string]string),
						},
					},
				},
				Services: []*k8sAPI.Service{
					&k8sAPI.Service{
						ObjectMeta: k8sAPI.ObjectMeta{
							Labels: make(map[string]string),
						},
					},
				},
			}},
		},
		Instance: &catalogModels.Instance{
			Id: instanceID,
		},
	}

	fakeData.Template.Body[0].Deployments[0].ObjectMeta.Labels[templateModels.PlaceholderIdxAndShortInstanceID] = idxAndShortInstanceID
	fakeData.Template.Body[0].Services[0].ObjectMeta.Name = serviceName

	secretData := make(map[string][]byte)
	secretData["ssl.crt"] = []byte(fakeCertkeyResponse.CertificateContent)
	secretData["ssl.key"] = []byte(fakeCertkeyResponse.KeyContent)
	expectedSecret := k8sAPI.Secret{
		ObjectMeta: k8sAPI.ObjectMeta{
			Name:   idxAndShortInstanceID + "-certkey",
			Labels: k8s.GetK8sTapLabels(fakeData.Instance.Id),
		},
		Data: secretData,
	}

	Convey("Given proper data (CatalogEntityWithK8sTemplate)", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		defer mockCtrl.Finish()
		mocker.caAPI.EXPECT().GetCertKey(serviceName).Return(fakeCertkeyResponse, nil)
		mocker.kubernetesAPI.EXPECT().CreateSecret(expectedSecret).Return(nil)

		Convey("When executing createSSLLayerForApplication", func() {
			err := manager.createSSLLayerForApplication(fakeData)

			Convey("returned error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given proper data but broken ca-service", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		defer mockCtrl.Finish()
		mocker.caAPI.EXPECT().GetCertKey(serviceName).Return(caModels.CertKeyResponse{}, errors.New("broken connection!"))

		Convey("When executing createSSLLayerForApplication", func() {
			err := manager.createSSLLayerForApplication(fakeData)

			Convey("returned error should resemble expected one", func() {
				So(err, ShouldResemble, errors.New("getting certkey failed: broken connection!"))
			})
		})
	})

	Convey("Given proper data but broken k8s api", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		defer mockCtrl.Finish()
		mocker.caAPI.EXPECT().GetCertKey(serviceName).Return(fakeCertkeyResponse, nil)
		mocker.kubernetesAPI.EXPECT().CreateSecret(expectedSecret).Return(errors.New("something bad happened!"))

		Convey("When executing createSSLLayerForApplication", func() {
			err := manager.createSSLLayerForApplication(fakeData)

			Convey("returned error should resemble expected one", func() {
				So(err, ShouldResemble, errors.New("creating secret failed: something bad happened!"))
			})
		})
	})

	fakeBadData := models.CatalogEntityWithK8sTemplate{}

	Convey("Given bad data", t, func() {
		_, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)
		defer mockCtrl.Finish()

		Convey("When executing createSSLLayerForApplication", func() {
			err := manager.createSSLLayerForApplication(fakeBadData)

			Convey("returned error should resemble expected one", func() {
				So(err, ShouldResemble, errors.New("data has nil pointers!"))
			})
		})
	})
}
