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

package engine

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

const (
	fakeAppName1 = "fake-app1"
	fakeAppName2 = "fake-app2"
	fakeAppName3 = "fake-app3"

	fakeImageVersion1 = "0.8.0.2061"
	fakeImageVersion2 = "0.8.0.2062"
	fakeImageVersion3 = "0.8.0.2063"

	fakeImage1 = "127.0.0.1:30000/" + fakeAppName1 + ":" + fakeImageVersion1
	fakeImage2 = "127.0.0.1:30000/" + fakeAppName2 + ":" + fakeImageVersion2
	fakeImage3 = "127.0.0.1:30000/" + fakeAppName3 + ":" + fakeImageVersion3

	fakeImageSignature1 = "290ba45ecc542d030053192d9dda0a7f4499cb1e3f140f949aa84b29afb48b1f"
	fakeImageSignature2 = "290ba45ecc542d030053192d9dda0a7f4499cb1e3f140f949aa84b29afb48b2f"
	fakeImageSignature3 = "290ba45ecc542d030053192d9dda0a7f4499cb1e3f140f949aa84b29afb48b3f"

	fakeImageID1 = "docker://sha256:" + fakeImageSignature1
	fakeImageID2 = "docker://sha256:" + fakeImageSignature2
	fakeImageID3 = "docker://sha256:" + fakeImageSignature3
)

func TestGetVersion(t *testing.T) {
	Convey("Test Versions", t, func() {
		mocker, manager, mockCtrl := prepareMockerAndMockedServiceEngineManager(t)

		Convey("Given existing deployments it should return array with valid version informations", func() {

			deploys := getFakeDeploymentList([]string{fakeAppName1, fakeAppName2, fakeAppName3})
			podsList1 := getSamplePodsList(fakeAppName1, fakeImage1, fakeImageID1)
			podsList2 := getSamplePodsList(fakeAppName2, fakeImage2, fakeImageID2)
			podsList3 := getSamplePodsList(fakeAppName3, fakeImage3, fakeImageID3)

			gomock.InOrder(
				mocker.kubernetesAPI.EXPECT().ListDeploymentsByLabel(appTypeLabelKey, appTypeLabelValue).Return(deploys, nil),
				mocker.kubernetesAPI.EXPECT().GetPodsBySpecifiedSelector(k8sAppSelectorName, fakeAppName1).Return(podsList1, nil),
				mocker.kubernetesAPI.EXPECT().GetPodsBySpecifiedSelector(k8sAppSelectorName, fakeAppName2).Return(podsList2, nil),
				mocker.kubernetesAPI.EXPECT().GetPodsBySpecifiedSelector(k8sAppSelectorName, fakeAppName3).Return(podsList3, nil),
			)

			responseVersions, err := manager.GetVersions()

			Convey("err should not be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("response should be proper", func() {
				So(responseVersions, ShouldResemble, getValidVersionResponse())
			})
		})
		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func getFakeDeploymentList(deployments []string) *extensions.DeploymentList {
	return &extensions.DeploymentList{
		Items: []extensions.Deployment{
			{
				ObjectMeta: api.ObjectMeta{Name: deployments[0]},
			},
			{
				ObjectMeta: api.ObjectMeta{Name: deployments[1]},
			},
			{
				ObjectMeta: api.ObjectMeta{Name: deployments[2]},
			},
		},
	}
}

func getSamplePodsList(containerName, image, imageId string) *api.PodList {
	return &api.PodList{
		Items: []api.Pod{
			{
				Spec:   getSamplePodSpec(containerName, image),
				Status: getSamplePodStatus(containerName, imageId),
			},
			{
				Spec:   getSamplePodSpec(containerName+"instance-2", image),
				Status: getSamplePodStatus(containerName, imageId),
			},
			{
				Spec:   getSamplePodSpec(containerName+"instance-3", image),
				Status: getSamplePodStatus(containerName, imageId),
			},
		},
	}
}

func getSamplePodSpec(containerName, image string) api.PodSpec {
	return api.PodSpec{
		Containers: []api.Container{
			{
				Name:  containerName,
				Image: image,
			},
			{
				Name:  containerName + "ssl",
				Image: image + "ssl",
			},
		},
	}
}

func getSamplePodStatus(containerName, imageId string) api.PodStatus {
	return api.PodStatus{
		ContainerStatuses: []api.ContainerStatus{
			{
				Name:    containerName,
				ImageID: imageId,
			},
			{
				Name:    containerName + "ssl",
				ImageID: imageId + "ssl",
			},
		},
	}
}

func getValidVersionResponse() []models.VersionsResponse {
	return []models.VersionsResponse{
		{
			Name:          fakeAppName1,
			Image:         fakeImage1,
			ImageVersions: fakeImageVersion1,
			Signature:     fakeImageSignature1,
		},
		{
			Name:          fakeAppName2,
			Image:         fakeImage2,
			ImageVersions: fakeImageVersion2,
			Signature:     fakeImageSignature2,
		},
		{
			Name:          fakeAppName3,
			Image:         fakeImage3,
			ImageVersions: fakeImageVersion3,
			Signature:     fakeImageSignature3,
		},
	}
}
