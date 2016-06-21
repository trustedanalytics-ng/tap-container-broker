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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"github.com/trustedanalytics-ng/tap-ceph-broker/client"
	cephModel "github.com/trustedanalytics-ng/tap-ceph-broker/model"
	"github.com/trustedanalytics-ng/tap-go-common/util"
	"github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	cephUser       string = "admin"
	cephSecretName string = "ceph-secret"
	cephPool       string = "rbd"
	cephFsType     string = "ext4"
	cephMonitors   string = "127.0.0.1"
)

func prepareRBDVolume(volumeName, idx_and_short_instance_id string, readOnly bool) api.Volume {
	volume := api.Volume{
		Name: volumeName,
	}

	rbdImageName := volumeName + "-" + idx_and_short_instance_id

	volume.VolumeSource.RBD = &api.RBDVolumeSource{
		RBDImage:     rbdImageName,
		ReadOnly:     readOnly,
		FSType:       util.GetEnvValueOrDefault("CEPH_FS_TYPE", cephFsType),
		RBDPool:      util.GetEnvValueOrDefault("CEPH_POOL", cephPool),
		CephMonitors: prepareCephMonitor(util.GetEnvValueOrDefault("CEPH_MONITORS", cephMonitors)),
		SecretRef: &api.LocalObjectReference{
			Name: util.GetEnvValueOrDefault("CEPH_SECRET_NAME", cephSecretName),
		},
		RadosUser: util.GetEnvValueOrDefault("CEPH_USER", cephUser),
	}
	return volume
}

const (
	VolumeNameKey     = "volume_name"
	VolumeSizeKey     = "volume_size_mb"
	VolumeReadOnlyKey = "volume_read_only"
)

func prepareCephMonitor(value string) []string {
	monitors := []string{}
	if value != "" {

		err := json.Unmarshal([]byte(value), &monitors)
		if err == nil {
			if len(monitors) <= 0 {
				logger.Errorf("CEPH monitors are empty! - value: %s", value)
			}
			return monitors
		}
		logger.Error("Cant unmarshall CEPH monitor value!", err)
	} else {
		logger.Error("CEPH_MONITORS env not set!")
	}
	return monitors
}

func isProperVolumeAnnotations(deployment extensions.Deployment) bool {
	return deployment.Spec.Template.ObjectMeta.Annotations[VolumeSizeKey] != "" &&
		deployment.Spec.Template.ObjectMeta.Annotations[VolumeNameKey] != "" &&
		deployment.Spec.Template.ObjectMeta.Labels[model.PlaceholderIdxAndShortInstanceID] != ""

}

func processDeploymentVolumes(deployment extensions.Deployment, cephClient client.CephBroker, isCreateAction bool) (api.Volume, error) {
	cephVolume := api.Volume{}

	if !isProperVolumeAnnotations(deployment) {
		return cephVolume, nil
	}

	idx_and_short_instance_id := deployment.Spec.Template.ObjectMeta.Labels[model.PlaceholderIdxAndShortInstanceID]
	volumeSize := deployment.Spec.Template.ObjectMeta.Annotations[VolumeSizeKey]
	volumeName := deployment.Spec.Template.ObjectMeta.Annotations[VolumeNameKey]
	volumeReadOnly, err := strconv.ParseBool(deployment.Spec.Template.ObjectMeta.Annotations[VolumeReadOnlyKey])

	if err != nil {
		volumeReadOnly = true
		logger.Notice("Cant parse read only key, setting to true %v", err)
	}

	//posibile to add another method if more volumes will be handled
	cephVolume = prepareRBDVolume(volumeName, idx_and_short_instance_id, volumeReadOnly)
	if isCreateAction {
		if err := addVolumeToCeph(cephVolume, cephClient, volumeSize); err != nil {
			return cephVolume, err
		}
	} else {
		if err := removeVolumeFromCeph(cephVolume, cephClient); err != nil {
			return cephVolume, err
		}
	}

	return cephVolume, nil
}

func addVolumeToCeph(volume api.Volume, cephClient client.CephBroker, size string) error {
	logger.Debug("CEPH volume creation: ", volume.RBD.RBDImage)
	status, err := cephClient.CreateRBD(cephModel.RBD{
		ImageName:  volume.RBD.RBDImage,
		Size:       getCephImageSize(size),
		FileSystem: volume.RBD.FSType,
	})
	if err != nil || status != 200 {
		message := fmt.Sprintf("CEPH volume creation error, image name: %s, statusCode: %d, error: %v", volume.RBD.RBDImage, status, err)
		logger.Errorf(message)
		return errors.New(message)
	}
	return nil
}

func removeVolumeFromCeph(volume api.Volume, cephClient client.CephBroker) error {
	logger.Debug("CEPH volume deletion: ", volume.RBD.RBDImage)
	status, err := cephClient.DeleteRBD(volume.RBD.RBDImage)
	if err != nil || status != 204 {
		message := fmt.Sprintf("CEPH volume deletion error, image name: %s, statusCode: %d, error: %v", volume.RBD.RBDImage, status, err)
		logger.Errorf(message)
		return errors.New(message)
	}
	return nil
}
