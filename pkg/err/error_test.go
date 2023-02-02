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

package err

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistsFormatErr(t *testing.T) {
	test := struct {
		fsType         string
		existingFormat string
		mountErr       error
		expectString   string
	}{
		fsType:         "ext4",
		existingFormat: "xfs",
		mountErr:       fmt.Errorf("error mount"),
		expectString:   "Failed to mount the volume as :ext4, volume has already contains xfs, volume Mount error: error mount",
	}

	exErr := ExistsFormatErr{
		FsType:         test.fsType,
		ExistingFormat: test.existingFormat,
		MountErr:       test.mountErr,
	}
	assert.Equal(t, test.expectString, exErr.Error())
}

func TestDeviceNotExistsErr(t *testing.T) {
	test := struct {
		device       string
		expectString string
	}{
		device:       "/dev/vdc",
		expectString: "Device [/dev/vdc] not exists in current node",
	}
	dr := DeviceNotExistsErr{
		Device: test.device,
	}
	assert.Equal(t, test.expectString, dr.Error())
}
