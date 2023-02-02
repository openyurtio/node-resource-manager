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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGetMetaData(t *testing.T) {
	defer gock.Off()
	testExamples := []struct {
		resource       string
		httpReturnData string
		expectResult   string
	}{
		{
			resource:       "region-id",
			httpReturnData: "cn-hangzhou",
			expectResult:   "cn-hangzhou",
		},
	}

	for _, test := range testExamples {
		gock.New("http://100.100.100.200").
			Get("/latest/meta-data/" + test.resource).
			Reply(200).
			BodyString(test.httpReturnData)
		aResult, err := GetMetaData(test.resource)
		assert.Nil(t, err)
		assert.Equal(t, test.expectResult, aResult)
	}

}

func TestRun(t *testing.T) {
	testExamples := []struct {
		cmd       string
		expectOut string
	}{
		{
			cmd:       "ls | wc -l",
			expectOut: "8\n",
		},
		{
			cmd:       "xxx",
			expectOut: "",
		},
	}

	for _, test := range testExamples {
		out, _ := Run(test.cmd)
		assert.Equal(t, out, test.expectOut)
	}
}

func TestNodeFilter(t *testing.T) {

}

func TestConvertRegion2Namespace(t *testing.T) {

}

func TestConvertNamespace2LVMDevicePath(t *testing.T) {

}

func TestCheckFSType(t *testing.T) {

}

func TestIsPart(t *testing.T) {
	testExamples := []struct {
		largeList  []string
		smallList  []string
		expectFlag bool
	}{
		{
			largeList:  []string{"a", "b", "c"},
			smallList:  []string{"a"},
			expectFlag: true,
		},
		{
			largeList:  []string{"a", "b", "c"},
			smallList:  []string{"d"},
			expectFlag: false,
		},
	}
	for _, test := range testExamples {
		bo := IsPart(test.largeList, test.smallList)
		assert.Equal(t, test.expectFlag, bo)
	}

}
