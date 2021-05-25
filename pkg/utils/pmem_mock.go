/*
Copyright 2021 The OpenYurt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/openyurtio/node-resource-manager/pkg/model"
)

// MockPmemer ..
type MockPmemer struct {
	ctrl     *gomock.Controller
	recorder *MockPmemerMockRecorder
}

// MockPmemerMockRecorder ...
type MockPmemerMockRecorder struct {
	mock *MockPmemer
}

// EXPECT ...
func (m *MockPmemer) EXPECT() *MockPmemerMockRecorder {
	return m.recorder
}

// NewMockPmemer ...
func NewMockPmemer(ctrl *gomock.Controller) *MockPmemer {
	mock := &MockPmemer{ctrl: ctrl}
	mock.recorder = &MockPmemerMockRecorder{mock}
	return mock
}

// GetRegions ...
func (m *MockPmemer) GetRegions() (*model.PmemRegions, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegions")
	ret0, _ := ret[0].(*model.PmemRegions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRegions ...
func (mr *MockPmemerMockRecorder) GetRegions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegions", reflect.TypeOf((*MockPmemer)(nil).GetRegions))
}

// CreateNamespace ...
func (m *MockPmemer) CreateNamespace(arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNamespace", arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateNamespace ...
func (mr *MockPmemerMockRecorder) CreateNamespace(arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNamespace", reflect.TypeOf((*MockPmemer)(nil).CreateNamespace), arg1, arg2)
}

// CheckNamespaceUsed ...
func (m *MockPmemer) CheckNamespaceUsed(arg1 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckNamespaceUsed", arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckNamespaceUsed ...
func (mr *MockPmemerMockRecorder) CheckNamespaceUsed(arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckNamespaceUsed", reflect.TypeOf((*MockPmemer)(nil).CheckNamespaceUsed), arg1)
}

// GetPmemNamespaceDeivcePath ...
func (m *MockPmemer) GetPmemNamespaceDeivcePath(arg1, arg2 string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPmemNamespaceDeivcePath", arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetPmemNamespaceDeivcePath ...
func (mr *MockPmemerMockRecorder) GetPmemNamespaceDeivcePath(arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPmemNamespaceDeivcePath", reflect.TypeOf((*MockPmemer)(nil).GetPmemNamespaceDeivcePath), arg1, arg2)
}

// CheckKMEMCreated ...
func (m *MockPmemer) CheckKMEMCreated(arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckKMEMCreated", arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckKMEMCreated ...
func (mr *MockPmemerMockRecorder) CheckKMEMCreated(arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckKMEMCreated", reflect.TypeOf((*MockPmemer)(nil).CheckKMEMCreated), arg1)
}

// MakeNamespaceMemory ...
func (m *MockPmemer) MakeNamespaceMemory(arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeNamespaceMemory", arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// MakeNamespaceMemory ...
func (mr *MockPmemerMockRecorder) MakeNamespaceMemory(arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeNamespaceMemory", reflect.TypeOf((*MockPmemer)(nil).MakeNamespaceMemory), arg1)
}
