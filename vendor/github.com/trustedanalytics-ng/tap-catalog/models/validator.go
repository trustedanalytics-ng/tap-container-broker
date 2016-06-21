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
package models

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	RegexpProperSystemEnvName = "^[A-Za-z_][A-Za-z0-9_]*$"
	RegexpDnsLabelLowercase   = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	IdFieldHasToBeEmptyMsg    = "Id field has to be empty!"
)

func CheckIfMatchingRegexp(content, regexpRule string) error {
	if ok, _ := regexp.MatchString(regexpRule, content); !ok {
		return fmt.Errorf("content: %s doesn't match regexp: %s !", content, regexpRule)
	}
	return nil
}

func GetIdFieldHasToBeEmptyError() error {
	return errors.New(IdFieldHasToBeEmptyMsg)
}

func GetInvalidValueError(field, value string, err error) error {
	return fmt.Errorf("field: %s has incorrect value: %s; %v", field, value, err)
}
