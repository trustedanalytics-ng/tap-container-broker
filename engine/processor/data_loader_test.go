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
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	k8sAPI "k8s.io/kubernetes/pkg/api"
	k8sEXT "k8s.io/kubernetes/pkg/apis/extensions"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	templateModel "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func getTemplateWithImage() (*templateModel.Template, *catalogModels.Image) {
	image := &catalogModels.Image{}
	template := &templateModel.Template{}
	template.Body = []templateModel.KubernetesComponent{{}}
	template.Body[0].Deployments = make([]*k8sEXT.Deployment, 1)
	template.Body[0].Deployments[0] = &k8sEXT.Deployment{}
	template.Body[0].Deployments[0].Spec.Template.Spec.Containers = make([]k8sAPI.Container, 1)

	return template, image
}

func TestGetJavaOpts(t *testing.T) {
	testCases := []struct {
		xms       string
		xmx       string
		metaspace string
		result    string
	}{
		{"", "", "", ""},
		{"228k", "", "128m", "-Xms228k -XX:MetaspaceSize=128m"},
		{"", "512m", "128m", "-Xmx512m -XX:MetaspaceSize=128m"},
		{"", "", "128m", "-XX:MetaspaceSize=128m"},
		{"1m", "1024m", "128m", "-Xms1m -Xmx1024m -XX:MetaspaceSize=128m"},
	}

	Convey("For set of test cases", t, func() {
		for _, tc := range testCases {
			Convey(fmt.Sprintf("Given xms %q xmx %q and metaspace %q getJavaOpts should return %q", tc.xms, tc.xmx, tc.metaspace, tc.result), func() {
				os.Setenv("JAVA_XMS", tc.xms)
				os.Setenv("JAVA_XMX", tc.xmx)
				os.Setenv("JAVA_METASPACE", tc.metaspace)

				res := getJavaOpts()

				So(res, ShouldEqual, tc.result)
			})
		}
	})
}

func TestAddAdditionalEnvsAndSetRaplicasToApplicationTemplate(t *testing.T) {
	Convey("Given Java application", t, func() {
		template, image := getTemplateWithImage()
		image.Type = catalogModels.ImageTypeJava

		replicas := 3
		xms := "228k"
		xmx := "512m"
		metaspace := "128m"
		Convey(fmt.Sprintf("Given envs xms=%q xmx=%q metaspace=%q", xms, xmx, metaspace), func() {
			os.Setenv("JAVA_XMS", xms)
			os.Setenv("JAVA_XMX", xmx)
			os.Setenv("JAVA_METASPACE", metaspace)
			Convey("addAdditionalEnvsAndSetReplicasToApplicationTemplate should add proper env", func() {
				javaOpts := k8sAPI.EnvVar{Name: "JAVA_OPTS", Value: getJavaOpts()}

				addAdditionalEnvsAndSetReplicasToApplicationTemplate(template, image, replicas)

				So(template.Body[0].Deployments[0].Spec.Template.Spec.Containers[0].Env, ShouldContain, javaOpts)
				So(template.Body[0].Deployments[0].Spec.Replicas, ShouldEqual, int32(replicas))
			})
		})

		Convey("Given envs without xms, xmx, metaspace", func() {
			template, image := getTemplateWithImage()
			image.Type = catalogModels.ImageTypeJava
			os.Unsetenv("JAVA_XMS")
			os.Unsetenv("JAVA_XMX")
			os.Unsetenv("JAVA_METASPACE")
			Convey("addAdditionalEnvsAndSetReplicasToApplicationTemplate should not add JAVA_OPTS", func() {
				tmp := template.Body[0].Deployments[0].Spec.Template.Spec.Containers[0].Env
				addAdditionalEnvsAndSetReplicasToApplicationTemplate(template, image, replicas)

				So(template.Body[0].Deployments[0].Spec.Template.Spec.Containers[0].Env, ShouldResemble, tmp)
				So(template.Body[0].Deployments[0].Spec.Replicas, ShouldEqual, int32(replicas))
			})
		})
	})

	template, image := getTemplateWithImage()
	Convey("Given Python27 application", t, func() {
		image.Type = catalogModels.ImageTypePython27
		replicas := 3
		Convey("addAdditionalEnvsAndSetReplicasToApplicationTemplate should not change envs", func() {
			tmp := template.Body[0].Deployments[0].Spec.Template.Spec.Containers[0].Env
			addAdditionalEnvsAndSetReplicasToApplicationTemplate(template, image, replicas)

			So(template.Body[0].Deployments[0].Spec.Template.Spec.Containers[0].Env, ShouldResemble, tmp)
			So(template.Body[0].Deployments[0].Spec.Replicas, ShouldEqual, int32(replicas))
		})
	})
}

func TestGetCatalogEntityWithK8sTemplate(t *testing.T) {
	Convey("Test TestGetCatalogEntityWithK8sTemplate method", t, func() {
		Convey("Should return proper response (not empty ServicePlan, not empty Template) with status 200 - Service Instance type", func() {
			mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)
			defer mockCtrl.Finish()

			instance, service, template := getSampleInstanceServiceAndTemplate()

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().GetService(offeringId).Return(service, http.StatusOK, nil),
				mocker.templateAPI.EXPECT().GenerateParsedTemplate(templateId, goodInstanceID, planName, gomock.Any()).Return(template, http.StatusOK, nil),
			)

			data, _, status, err := manager.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(data.Template, ShouldNotBeNil)
			So(data.Template.Id, ShouldEqual, templateId)
			So(data.ServicePlan, ShouldNotBeNil)
			So(data.ServicePlan.Name, ShouldEqual, planName)
		})

		Convey("Should return error if no assigned servicePlan - Service Instance type", func() {
			mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)
			defer mockCtrl.Finish()

			instance, service, _ := getSampleInstanceServiceAndTemplate()
			instance.Metadata = []catalogModels.Metadata{
				{
					Id:    catalogModels.OFFERING_PLAN_ID,
					Value: "wrong-plan-id",
				},
			}

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().GetService(offeringId).Return(service, http.StatusOK, nil),
			)

			_, _, status, err := manager.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: goodInstanceID})
			So(err, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, "Plan is missing!")
		})

		Convey("Should return proper response (empty ServicePlan, not empty Template) with status 200 - ServiceBroker Instance type", func() {
			mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)
			defer mockCtrl.Finish()

			instance, _, template := getSampleInstanceServiceAndTemplate()
			instance.Type = catalogModels.InstanceTypeServiceBroker
			instance.Metadata = append(instance.Metadata, catalogModels.Metadata{Id: catalogModels.BROKER_TEMPLATE_ID, Value: templateId})

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.templateAPI.EXPECT().GenerateParsedTemplate(templateId, goodInstanceID, templateModel.EMPTY_PLAN_NAME, gomock.Any()).Return(template, http.StatusOK, nil),
			)

			data, _, status, err := manager.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{
				InstanceId: goodInstanceID,
			})
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(data.Template, ShouldNotBeNil)
			So(data.Template.Id, ShouldEqual, templateId)
			So(data.ServicePlan, ShouldBeNil)
		})

		Convey("Should return error if no TemplateId in request - ServiceBroker Instance type", func() {
			mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)
			defer mockCtrl.Finish()

			instance, _, _ := getSampleInstanceServiceAndTemplate()
			instance.Type = catalogModels.InstanceTypeServiceBroker

			mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil)

			_, _, status, err := manager.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: goodInstanceID})
			So(err, ShouldNotBeNil)
			So(status, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, "BROKER_TEMPLATE_ID has to exist in Metadata and not be empty for service-broker instance!")
		})

		Convey("Should return proper response (empty ServicePlan, not empty Template) with status 200 - Application Instance type", func() {
			mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)
			defer mockCtrl.Finish()

			instance, _, template := getSampleInstanceServiceAndTemplate()
			instance.Type = catalogModels.InstanceTypeApplication
			app, image := getSampleApplicationAndImage()

			gomock.InOrder(
				mocker.catalogAPI.EXPECT().GetInstance(goodInstanceID).Return(instance, http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().GetApplication(offeringId).Return(app, http.StatusOK, nil),
				mocker.catalogAPI.EXPECT().GetImage(imageId).Return(image, http.StatusOK, nil),
				mocker.templateAPI.EXPECT().GenerateParsedTemplate(templateId, goodInstanceID, templateModel.EMPTY_PLAN_NAME, gomock.Any()).Return(template, http.StatusOK, nil),
			)

			data, _, status, err := manager.GetCatalogEntityWithK8sTemplate(models.CatalogEntityWithK8sTemplateRequest{InstanceId: goodInstanceID})
			So(err, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)
			So(data.Template, ShouldNotBeNil)
			So(data.Template.Id, ShouldEqual, templateId)
			So(data.ServicePlan, ShouldBeNil)
		})
	})

}
