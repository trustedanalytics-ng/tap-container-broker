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
)

func TestConvertToProperEnvName(t *testing.T) {
	Convey("Test ConvertToProperEnvName should replace - with _", t, func() {
		result := ConvertToProperEnvName("my-app")
		So(result, ShouldEqual, "my_app")
	})
}

func TestGetCephImageSize(t *testing.T) {
	Convey("Test getCephImageSize should return default value if no size provided", t, func() {
		result := getCephImageSize("")
		So(result, ShouldEqual, defaultCephImageSizeMB)
	})
	Convey("Test getCephImageSize should return uint64 from string ", t, func() {
		result := getCephImageSize("")
		So(result, ShouldEqual, defaultCephImageSizeMB)
	})
}

func TestAddProtocolToHost(t *testing.T) {
	host := "example.pl"
	hostWithHttp := "http://example.pl"
	hostWithHttps := "https://example.pl"

	Convey("Test addProtocolToHost should return host with http prefix", t, func() {
		annotations := map[string]string{useExternalSslFlag: "false"}

		result := addProtocolToHost(annotations, host)
		So(result, ShouldEqual, hostWithHttp)
	})

	Convey("Test addProtocolToHost should return host with http prefix if no info in annotations", t, func() {
		annotations := make(map[string]string)

		result := addProtocolToHost(annotations, host)
		So(result, ShouldEqual, hostWithHttp)
	})

	Convey("Test addProtocolToHost should return host with http prefix if annotations can not be parsed", t, func() {
		annotations := map[string]string{useExternalSslFlag: "wrongValue"}

		result := addProtocolToHost(annotations, host)
		So(result, ShouldEqual, hostWithHttp)
	})

	Convey("Test addProtocolToHost should return host with https prefix", t, func() {
		annotations := map[string]string{useExternalSslFlag: "true"}

		result := addProtocolToHost(annotations, host)
		So(result, ShouldEqual, hostWithHttps)
	})

	Convey("Test addProtocolToHost should return unchanged host", t, func() {
		annotations := make(map[string]string)
		readyToUseHost := "ftp://HOST"

		result := addProtocolToHost(annotations, readyToUseHost)
		So(result, ShouldEqual, readyToUseHost)
	})
}

func TestAppendSourceEnvsToDestinationEnvsIfNotContained(t *testing.T) {
	Convey("Test appendSourceEnvsToDestinationEnvsIfNotContained should append only not existing EnvVars", t, func() {
		source := []api.EnvVar{
			{Name: "test", Value: "old"},
			{Name: "not-contained", Value: "not-contained"},
		}
		dst := []api.EnvVar{
			{Name: "test", Value: "updated"},
			{Name: "contained", Value: "contained"},
		}

		result := appendSourceEnvsToDestinationEnvsIfNotContained(source, dst)
		So(len(result), ShouldEqual, 3)
		So(result[0].Name, ShouldEqual, "test")
		So(result[0].Value, ShouldEqual, "updated")

		result = appendSourceEnvsToDestinationEnvsIfNotContained(dst, source)
		So(len(result), ShouldEqual, 3)
		So(result[0].Name, ShouldEqual, "test")
		So(result[0].Value, ShouldEqual, "old")
	})
}
