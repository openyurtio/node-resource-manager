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

// MockLVM ...
type MockLVM struct {
	ctrl     *gomock.Controller
	recorder *MockLVMMockRecorder
}

// MockLVMMockRecorder ...
type MockLVMMockRecorder struct {
	mock *MockLVM
}

// EXPECT ...
func (m *MockLVM) EXPECT() *MockLVMMockRecorder {
	return m.recorder
}

// NewMockLVM ...
func NewMockLVM(ctrl *gomock.Controller) *MockLVM {
	mock := &MockLVM{ctrl: ctrl}
	mock.recorder = &MockLVMMockRecorder{mock}
	return mock
}

// ListLV ...
func (m *MockLVM) ListLV(listSpec string) ([]*model.LV, error) {

	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLV", listSpec)
	ret0, _ := ret[0].([]*model.LV)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLV ...
func (mr *MockLVMMockRecorder) ListLV(arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLV", reflect.TypeOf((*MockLVM)(nil).ListLV), arg1)
}

// CreateLV ...
func (m *MockLVM) CreateLV(vg, name string, size uint64, mirrors uint32, tags []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLV", vg, name, size, mirrors, tags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateLV ...
func (mr *MockLVMMockRecorder) CreateLV(arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLV", reflect.TypeOf((*MockLVM)(nil).CreateLV), arg1, arg2, arg3, arg4, arg5)
}

// RemoveLV ...
func (m *MockLVM) RemoveLV(vg, name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveLV", vg, name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveLV ...
func (mr *MockLVMMockRecorder) RemoveLV(arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveLV", reflect.TypeOf((*MockLVM)(nil).RemoveLV), arg1, arg2)
}

// CloneLV ...
func (m *MockLVM) CloneLV(src, dest string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloneLV", src, dest)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloneLV ...
func (mr *MockLVMMockRecorder) CloneLV(arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloneLV", reflect.TypeOf((*MockLVM)(nil).CloneLV), arg1, arg2)
}

// ListVG ...
func (m *MockLVM) ListVG() ([]*model.VG, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListVG")
	ret0, _ := ret[0].([]*model.VG)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListVG ...
func (mr *MockLVMMockRecorder) ListVG() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListVG", reflect.TypeOf((*MockLVM)(nil).ListVG))
}

// ListPhysicalVolume ...
func (m *MockLVM) ListPhysicalVolume() ([]*model.PV, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPhysicalVolume")
	ret0, _ := ret[0].([]*model.PV)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPhysicalVolume ...
func (mr *MockLVMMockRecorder) ListPhysicalVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPhysicalVolume", reflect.TypeOf((*MockLVM)(nil).ListPhysicalVolume))
}

// CreateVG ...
func (m *MockLVM) CreateVG(name, physicalVolume string, tags []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVG", name, physicalVolume, tags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1

}

// CreateVG ...
func (mr *MockLVMMockRecorder) CreateVG(arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVG", reflect.TypeOf((*MockLVM)(nil).CreateVG), arg1, arg2, arg3)
}

// ExtendVG ...
func (m *MockLVM) ExtendVG(name, physicalVolume string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtendVG", name, physicalVolume)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExtendVG ...
func (mr *MockLVMMockRecorder) ExtendVG(arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExtendVG", reflect.TypeOf((*MockLVM)(nil).ExtendVG), arg1, arg2)
}

// RemoveVG ...
func (m *MockLVM) RemoveVG(name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtendVG", name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveVG ...
func (mr *MockLVMMockRecorder) RemoveVG(arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveVG", reflect.TypeOf((*MockLVM)(nil).RemoveVG), arg1)
}

// AddTagLV ...
func (m *MockLVM) AddTagLV(vg, name string, tags []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTagLV", vg, name, tags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddTagLV ...
func (mr *MockLVMMockRecorder) AddTagLV(arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTagLV", reflect.TypeOf((*MockLVM)(nil).AddTagLV), arg1, arg2, arg3)
}

// RemoveTagLV ...
func (m *MockLVM) RemoveTagLV(vg, name string, tags []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveTagLV", vg, name, tags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveTagLV ...
func (mr *MockLVMMockRecorder) RemoveTagLV(arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveTagLV", reflect.TypeOf((*MockLVM)(nil).RemoveTagLV), arg1, arg2, arg3)
}
