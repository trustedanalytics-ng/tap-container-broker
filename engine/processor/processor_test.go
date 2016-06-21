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

package processor

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"

	"github.com/trustedanalytics-ng/tap-container-broker/models"
)

func TestMonitorInstanceDeployment(t *testing.T) {
	Convey("Test MonitorInstanceDeployment", t, func() {
		setProcessorParams(time.Millisecond, 2, 2)
		mocker, manager, mockCtrl := prepareMockerAndMockerServiceProcessorManager(t)

		Convey("Should start and return nil error response", func() {
			instance, _, _ := getSampleInstanceServiceAndTemplate()
			pod := getSamplePod(api.PodRunning, true)

			gomock.InOrder(
				mocker.kubernetesAPI.EXPECT().GetPodsByInstanceId(goodInstanceID).Return([]api.Pod{pod}, nil),
				mocker.kubernetesAPI.EXPECT().GetPodEvents(gomock.Any()).Return(nil, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, patchesInstanceStateStartingToRunning).Return(instance, http.StatusOK, nil),
			)

			err := manager.MonitorInstanceDeployments(instance)
			So(err, ShouldBeNil)
		})

		Convey("Should return error if there is no pods", func() {
			instance, _, _ := getSampleInstanceServiceAndTemplate()

			mocker.kubernetesAPI.EXPECT().GetPodsByInstanceId(goodInstanceID).Return([]api.Pod{}, nil).Times(int(models.ProcessorsTries))

			err := manager.MonitorInstanceDeployments(instance)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "pod can't be stared! Event message:")
		})

		Convey("Should stop instance if pod has pending containers", func() {
			instance, _, _ := getSampleInstanceServiceAndTemplate()
			pod := getSamplePod(api.PodPending, false)
			gomock.InOrder(
				mocker.kubernetesAPI.EXPECT().GetPodsByInstanceId(goodInstanceID).Return([]api.Pod{pod}, nil),
				mocker.kubernetesAPI.EXPECT().GetPodEvents(gomock.Any()).Return(nil, nil),
				mocker.kubernetesAPI.EXPECT().GetPodsByInstanceId(goodInstanceID).Return([]api.Pod{pod}, nil),
				mocker.kubernetesAPI.EXPECT().GetPodEvents(gomock.Any()).Return(nil, nil),
				mocker.catalogAPI.EXPECT().UpdateInstance(goodInstanceID, gomock.Any()).Return(instance, http.StatusOK, nil),
			)

			err := manager.MonitorInstanceDeployments(instance)
			So(err, ShouldBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
