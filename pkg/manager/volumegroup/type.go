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

package volumegroup

import (
	"github.com/openyurtio/node-resource-manager/pkg/model"
)

// VgDeviceConfig ...
type VgDeviceConfig struct {
	Name            string   `json:"name,omitempty"`
	PhysicalVolumes []string `json:"physicalVolumes,omitempty"`
}

// VgList ...
type VgList struct {
	VolumeGroups []model.ResourceYaml `yaml:"volumegroup,omitempty"`
}
