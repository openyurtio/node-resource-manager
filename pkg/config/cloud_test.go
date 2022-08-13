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

package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGetDefaultAK(t *testing.T) {
	defer gock.Off()
	testExamples := []struct {
		key            string
		value          string
		stsToken       string
		roleName       string
		httpReturnData string

		expectKey      string
		expectValue    string
		expectSTSToken string
	}{
		{
			key:            "xxx1",
			value:          "vvv1",
			expectKey:      "xxx1",
			expectValue:    "vvv1",
			expectSTSToken: "",
		},
		{
			key:            "",
			value:          "",
			stsToken:       "",
			roleName:       "test1",
			httpReturnData: "{\"AccessKeyId\": \"test1\", \"AccessKeySecret\": \"keySecret1\", \"Expiration\": \"2022-08-22T14:03:26Z\", \"SecurityToken\": \"token\",\"LastUpdated\": \"2022-08-22T08:03:26Z\", \"Code\": \"Success\"}",
			expectKey:      "test1",
		},
	}

	for _, test := range testExamples {
		if test.roleName == "" {
			os.Setenv("ACCESS_KEY_ID", test.key)
			os.Setenv("ACCESS_KEY_SECRET", test.value)
			actualKey, actualValue, actualSTSToken := GetDefaultAK()
			assert.Equal(t, test.expectKey, actualKey)
			assert.Equal(t, test.expectValue, actualValue)
			assert.Equal(t, test.expectSTSToken, actualSTSToken)
		} else {
			os.Setenv("ACCESS_KEY_ID", "")
			os.Setenv("ACCESS_KEY_SECRET", "")
			gock.New("http://100.100.100.200").
				Get("/latest/meta-data/ram/security-credentials/").
				Reply(200).
				BodyString(test.roleName)
			gock.New("http://100.100.100.200").
				Get("/latest/meta-data/ram/security-credentials/" + test.roleName).
				Reply(200).
				BodyString(test.httpReturnData)
			actualKey, _, _ := GetDefaultAK()
			assert.Equal(t, test.expectKey, actualKey)
		}
	}

}

func TestGetLocalAK(t *testing.T) {
	testExamples := []struct {
		key   string
		value string

		expectKey   string
		expectValue string
	}{
		{
			key:         "xxx1",
			value:       "vvv1",
			expectKey:   "xxx1",
			expectValue: "vvv1",
		},
	}
	for _, test := range testExamples {
		os.Setenv("ACCESS_KEY_ID", test.key)
		os.Setenv("ACCESS_KEY_SECRET", test.value)
		actualKey, actualValue := GetLocalAK()
		assert.Equal(t, test.expectKey, actualKey)
		assert.Equal(t, test.expectValue, actualValue)

	}
}

func TestGetSTSAK(t *testing.T) {
	defer gock.Off()

	testExamples := []struct {
		roleName       string
		roleError      error
		credReturnData string
		credError      error
		expectKeyID    string
	}{
		{
			roleName:    "test1",
			roleError:   errors.New("test"),
			expectKeyID: "",
		},
		{
			roleName:       "test2",
			credReturnData: "{\"AccessKeyId\": \"test2\", \"AccessKeySecret\": \"keySecret2\", \"Expiration\": \"2022-08-22T14:03:26Z\", \"SecurityToken\": \"token\",\"LastUpdated\": \"2022-08-22T08:03:26Z\", \"Code\": \"Success\"}",
			credError:      errors.New("test"),
			expectKeyID:    "",
		},
		{
			roleName:       "test3",
			credReturnData: "{\"AccessKeyId\": \"test2\", \"AccessKeySecret\": \"keySecret2\", \"Expiration\": \"2022-08-22T14:03:26Z\", \"SecurityToken\": \"token\",\"LastUpdated\": \"2022-08-22T08:03:26Z\", \"Code\": 'Success\"}",
			expectKeyID:    "",
		},
		{
			roleName:       "test4",
			credReturnData: "{\"AccessKeyId\": \"test4\", \"AccessKeySecret\": \"keySecret2\", \"Expiration\": \"2022-08-22T14:03:26Z\", \"SecurityToken\": \"token\",\"LastUpdated\": \"2022-08-22T08:03:26Z\", \"Code\": \"Success\"}",
			expectKeyID:    "test4",
		},
	}

	for _, test := range testExamples {
		if test.roleError != nil {
			gock.New("http://100.100.100.200").
				Get("/latest/meta-data/ram/security-credentials/").
				ReplyError(test.roleError)
			actualKey, _, _ := GetSTSAK()
			assert.Equal(t, test.expectKeyID, actualKey)
			continue
		}
		if test.credError != nil {
			gock.New("http://100.100.100.200").
				Get("/latest/meta-data/ram/security-credentials/").
				Reply(200).
				BodyString(test.roleName)
			gock.New("http://100.100.100.200").
				Get("/latest/meta-data/ram/security-credentials/" + test.roleName).
				ReplyError(test.credError)
			actualKey, _, _ := GetSTSAK()
			assert.Equal(t, test.expectKeyID, actualKey)
			continue
		}

		gock.New("http://100.100.100.200").
			Get("/latest/meta-data/ram/security-credentials/").
			Reply(200).
			BodyString(test.roleName)
		gock.New("http://100.100.100.200").
			Get("/latest/meta-data/ram/security-credentials/" + test.roleName).
			Reply(200).
			BodyString(test.credReturnData)
		actualKey, _, _ := GetSTSAK()
		assert.Equal(t, test.expectKeyID, actualKey)
	}
}
