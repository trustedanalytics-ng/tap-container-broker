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
package models

import "strings"

const USER_DEFINED_APPLICATION_IMAGE_PREFIX = "app_"
const USER_DEFINED_OFFERING_IMAGE_PREFIX = "svc_"

type Image struct {
	Id         string     `json:"id"`
	Type       ImageType  `json:"type"`
	BlobType   BlobType   `json:"blobType"`
	State      ImageState `json:"state"`
	AuditTrail AuditTrail `json:"auditTrail"`
}

type ImageRefsResponse struct {
	IsAnyRefExist         bool          `json:"refsExists"`
	ApplicationReferences []Application `json:"applicationReferences"`
	ServiceReferences     []Service     `json:"serviceReferences"`
}

type ImageType string

const (
	ImageTypeJava     ImageType = "JAVA"
	ImageTypeGo       ImageType = "GO"
	ImageTypeNodeJs   ImageType = "NODEJS"
	ImageTypePython27 ImageType = "PYTHON2.7"
	ImageTypePython34 ImageType = "PYTHON3.4"
)

type BlobType string

const (
	BlobTypeTarGz BlobType = "TARGZ"
	BlobTypeJar   BlobType = "JAR"
	BlobTypeExec  BlobType = "EXEC"
)

type ImageState string

const (
	ImageStateBuilding  ImageState = "BUILDING"
	ImageStateError     ImageState = "ERROR"
	ImageStatePending   ImageState = "PENDING"
	ImageStateReady     ImageState = "READY"
	ImageStateRequested ImageState = "REQUESTED"
	ImageStateRemoving  ImageState = "REMOVING"
)

func IsApplicationInstance(imageId string) bool {
	return strings.HasPrefix(imageId, USER_DEFINED_APPLICATION_IMAGE_PREFIX)
}

func GetApplicationId(imageId string) string {
	return strings.TrimPrefix(imageId, USER_DEFINED_APPLICATION_IMAGE_PREFIX)
}

func GenerateImageId(applicationId string) string {
	return USER_DEFINED_APPLICATION_IMAGE_PREFIX + applicationId
}

func IsUserDefinedOffering(imageId string) bool {
	return strings.HasPrefix(imageId, USER_DEFINED_OFFERING_IMAGE_PREFIX)
}

func GetOfferingId(imageId string) string {
	return strings.TrimPrefix(imageId, USER_DEFINED_OFFERING_IMAGE_PREFIX)
}

func ConstructImageIdForUserOffering(offeringId string) string {
	return USER_DEFINED_OFFERING_IMAGE_PREFIX + offeringId
}
