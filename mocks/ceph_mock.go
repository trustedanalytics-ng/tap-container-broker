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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: vendor/github.com/trustedanalytics-ng/tap-ceph-broker/client/client.go

package mocks

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/trustedanalytics-ng/tap-ceph-broker/model"
)

// Mock of CephBroker interface
type MockCephBroker struct {
	ctrl     *gomock.Controller
	recorder *_MockCephBrokerRecorder
}

// Recorder for MockCephBroker (not exported)
type _MockCephBrokerRecorder struct {
	mock *MockCephBroker
}

func NewMockCephBroker(ctrl *gomock.Controller) *MockCephBroker {
	mock := &MockCephBroker{ctrl: ctrl}
	mock.recorder = &_MockCephBrokerRecorder{mock}
	return mock
}

func (_m *MockCephBroker) EXPECT() *_MockCephBrokerRecorder {
	return _m.recorder
}

func (_m *MockCephBroker) CreateRBD(device model.RBD) (int, error) {
	ret := _m.ctrl.Call(_m, "CreateRBD", device)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCephBrokerRecorder) CreateRBD(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateRBD", arg0)
}

func (_m *MockCephBroker) DeleteRBD(name string) (int, error) {
	ret := _m.ctrl.Call(_m, "DeleteRBD", name)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCephBrokerRecorder) DeleteRBD(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteRBD", arg0)
}

func (_m *MockCephBroker) ListLocks() ([]model.Lock, int, error) {
	ret := _m.ctrl.Call(_m, "ListLocks")
	ret0, _ := ret[0].([]model.Lock)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockCephBrokerRecorder) ListLocks() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListLocks")
}

func (_m *MockCephBroker) DeleteLock(lock model.Lock) (int, error) {
	ret := _m.ctrl.Call(_m, "DeleteLock", lock)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCephBrokerRecorder) DeleteLock(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteLock", arg0)
}

func (_m *MockCephBroker) GetCephBrokerHealth() (int, error) {
	ret := _m.ctrl.Call(_m, "GetCephBrokerHealth")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCephBrokerRecorder) GetCephBrokerHealth() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetCephBrokerHealth")
}
