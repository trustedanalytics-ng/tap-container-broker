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

package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

var imageAdressRegexp = regexp.MustCompile(`(?P<host>[^:]+:?.+)/(?P<name>[^:]+):?(?P<tag>.*)$`)

func ParseImageAddress(imageAddress string) (hostWithPort, imageName, imageTag string, err error) {
	match := imageAdressRegexp.FindStringSubmatch(imageAddress)
	result := make(map[string]string)
	if len(match) == 4 {
		for i, name := range imageAdressRegexp.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		hostWithPort = result["host"]
		imageName = result["name"]
		imageTag = result["tag"]
	} else {
		err = fmt.Errorf("cannot split image value from address: %s using regexp: %s", imageAddress, imageAdressRegexp)
	}
	return
}

func RandomString(length int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const charSet = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(result)
}
