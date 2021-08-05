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

import "fmt"

type ExistsFormatErr struct {
	FsType         string
	ExistingFormat string
	MountErr       error
}

func (e *ExistsFormatErr) Error() string {
	return fmt.Sprintf("Failed to mount the volume as :%v, volume has already contains %v, volume Mount error: %v", e.FsType, e.ExistingFormat, e.MountErr)

}

type DeviceNotExistsErr struct {
	Device string
}

func (e *DeviceNotExistsErr) Error() string {
	return fmt.Sprintf("Device [%s] not exists in current node", e.Device)
}
