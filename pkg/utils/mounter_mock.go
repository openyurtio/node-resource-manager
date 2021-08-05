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
	utilexec "k8s.io/utils/exec"
	k8smount "k8s.io/utils/mount"
)

// MockMounter ...
type MockMounter struct {
	k8smount.SafeFormatAndMount
	utilexec.Interface
	ctrl     *gomock.Controller
	recorder *MockMounterMockRecorder
}

// MockMounterMockRecorder ...
type MockMounterMockRecorder struct {
	mock *MockMounter
}

// NewMockMounter ...
func NewMockMounter(ctrl *gomock.Controller) *MockMounter {
	mock := &MockMounter{ctrl: ctrl}
	mock.recorder = &MockMounterMockRecorder{mock}
	return mock
}

// EXPECT ...
func (m *MockMounter) EXPECT() *MockMounterMockRecorder {
	return m.recorder
}

// FileExists ...
func (m *MockMounter) FileExists(filename string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FileExists", filename)
	ret0, _ := ret[0].(bool)
	return ret0
}

// EnsureFolder ...
func (mr MockMounterMockRecorder) FileExists(filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FileExists", reflect.TypeOf((*MockMounter)(nil).FileExists), filename)
}

// EnsureFolder ...
func (m *MockMounter) EnsureFolder(target string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureFolder", target)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnsureFolder ...
func (mr MockMounterMockRecorder) EnsureFolder(target interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureFolder", reflect.TypeOf((*MockMounter)(nil).EnsureFolder), target)
}

// FormatAndMount ...
func (m *MockMounter) FormatAndMount(source, target, fstype string, mkfsOptions []string, mountOptions string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatAndMount", source, target, fstype, mkfsOptions, mountOptions)
	ret0, _ := ret[0].(error)
	return ret0
}

// FormatAndMount ...
func (mr MockMounterMockRecorder) FormatAndMount(arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatAndMount", reflect.TypeOf((*MockMounter)(nil).FormatAndMount), arg1, arg2, arg3, arg4, arg5)
}

// IsMounted ...
func (m *MockMounter) IsMounted(target string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsMounted", target)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsMounted ...
func (mr MockMounterMockRecorder) IsMounted(target interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsMounted", reflect.TypeOf((*MockMounter)(nil).IsMounted), target)
}

// SafePathRemove ...
func (m MockMounter) SafePathRemove(targetPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SafePathRemove", targetPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// SafePathRemove ...
func (mr MockMounterMockRecorder) SafePathRemove(target interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SafePathRemove", reflect.TypeOf((*MockMounter)(nil).SafePathRemove), target)
}
