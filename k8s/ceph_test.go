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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/golang/mock/gomock"
	cephModel "github.com/trustedanalytics-ng/tap-ceph-broker/model"
	"github.com/trustedanalytics-ng/tap-container-broker/mocks"
	"github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestPrepareCephMonitor(t *testing.T) {
	Convey("Given empty string should give empty string", t, func() {
		result := prepareCephMonitor("")
		So(result, ShouldResemble, []string{})
	})
	Convey("Given no elements in array should give empty string", t, func() {
		result := prepareCephMonitor("[]")
		So(result, ShouldResemble, []string{})
	})
	Convey("Given one element should return that element", t, func() {
		result := prepareCephMonitor(`["127.0.0.1"]`)
		So(result, ShouldResemble, []string{`127.0.0.1`})
	})
	Convey("Given many elements should return string coma separated with quotes inside", t, func() {
		result := prepareCephMonitor(`["127.0.0.1","127.0.0.2","127.0.0.3"]`)
		So(result, ShouldResemble, []string{`127.0.0.1`, `127.0.0.2`, `127.0.0.3`})
	})
}

func TestProcessDeploymentVolumes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	cephClient := mocks.NewMockCephBroker(mockCtrl)

	deployment := extensions.Deployment{}
	deployment.Spec.Template.Spec.Volumes = []api.Volume{}

	deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	deployment.Spec.Template.ObjectMeta.Annotations[VolumeNameKey] = "test"
	deployment.Spec.Template.ObjectMeta.Annotations[VolumeSizeKey] = "10"
	deployment.Spec.Template.ObjectMeta.Annotations[VolumeReadOnlyKey] = "false"
	deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	deployment.Spec.Template.ObjectMeta.Labels[model.PlaceholderIdxAndShortInstanceID] = "1"

	Convey("Given proper volume in specfication and create action should create RBD ", t, func() {
		cephClient.EXPECT().CreateRBD(cephModel.RBD{
			ImageName:  "test-1",
			Size:       getCephImageSize("10"),
			FileSystem: cephFsType,
		}).Return(200, nil)

		volume, err := processDeploymentVolumes(deployment, cephClient, true)

		Convey("Should return nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should return proper volume", func() {

			So(volume.RBD.RBDImage, ShouldEqual, "test-1")
			So(volume.RBD.ReadOnly, ShouldEqual, false)
			So(volume.Name, ShouldEqual, "test")
		})
	})

	Convey("Given proper volume in specfication and no create action should create RBD ", t, func() {
		cephClient.EXPECT().DeleteRBD("test-1").Return(204, nil)

		_, err := processDeploymentVolumes(deployment, cephClient, false)

		Convey("Should return nil", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("Given wrong volume in specfication should return empty volume", t, func() {
		//overwrite deployment
		deployment := extensions.Deployment{}
		deployment.Spec.Template.Spec.Volumes = []api.Volume{}

		volume, err := processDeploymentVolumes(deployment, cephClient, false)

		Convey("Should return nil error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should return empty volume", func() {
			So(volume.Name, ShouldEqual, "")
			So(volume.RBD, ShouldBeNil)
		})
	})
}
