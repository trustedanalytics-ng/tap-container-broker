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

import (
	"encoding/json"
	"errors"
	"reflect"
)

var Registry = map[string]reflect.Type{}

func RegisterType(name string, t reflect.Type) {
	Registry[name] = t
}

type PatchOperation string

const (
	OperationAdd    PatchOperation = "Add"
	OperationUpdate PatchOperation = "Update"
	OperationDelete PatchOperation = "Delete"
)

type Patch struct {
	Operation PatchOperation   `json:"op"`
	Field     *string          `json:"field"`
	Value     *json.RawMessage `json:"value"`
	PrevValue json.RawMessage  `json:"prevValue,omitempty"`
	Username  string           `json:"username"`
}

func ValidatePatchStructure(patch Patch) error {
	if patch.Value == nil {
		return errors.New("field value is empty!")
	} else if patch.Field == nil || *patch.Field == "" {
		return errors.New("field field is empty!")
	}
	return nil
}

const (
	WatchFromNow uint64 = 0
)

type StateChange struct {
	Id    string
	State string
	Index uint64
}
