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

package quotapath

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openyurtio/node-resource-manager/pkg/config"
	"github.com/openyurtio/node-resource-manager/pkg/model"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeValidResourceYaml() *model.ResourceYaml {
	return &model.ResourceYaml{
		Name: "/tmp/foo",
	}
}

func makeResourceYamlCustom(tweaks ...func(*model.ResourceYaml)) *model.ResourceYaml {
	resourceYaml := makeValidResourceYaml()
	for _, fn := range tweaks {
		fn(resourceYaml)
	}
	return resourceYaml
}

// EnsureVolumeGroupEnv ...
func EnsureVolumeGroupEnv() (string, error, *ResourceManager) {
	config.GlobalConfigVar.NodeInfo = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "foo",
			Labels: map[string]string{"bar": "foo"},
		},
	}
	configPath := "/tmp/quotapath"
	mdkirCmd := "mkdir"
	_, err := exec.LookPath(mdkirCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return "", fmt.Errorf("EnsureFolder:: %q executable not found in $PATH", mdkirCmd), nil
		}
		return "", err, nil
	}

	mkdirArgs := []string{"-p", filepath.Dir(configPath)}
	//log.Infof("mkdir for folder, the command is %s %v", mdkirCmd, mkdirArgs)
	_, err = exec.Command(mdkirCmd, mkdirArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("EnsureFolder:: mkdir for folder error: %v", err), nil
	}

	newMockVolumegroupResourceManager := func() *ResourceManager {
		return &ResourceManager{
			configPath: configPath,
		}
	}
	return configPath, nil, newMockVolumegroupResourceManager()
}

func TestAnalyseDiff(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	configPath, err, resourceManager := EnsureVolumeGroupEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configPath)
	mockPmemer := utils.NewMockPmemer(mockCtl)
	mockMounter := utils.NewMockMounter(mockCtl)
	resourceManager.mounter = mockMounter
	resourceManager.pmemer = mockPmemer
	setOpInOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpIn
		m.Value = "foo"
	}

	setPmemTopology := func(m *model.ResourceYaml) {
		m.Topology = model.Topology{
			Type:    "pmem",
			Options: "prjquota,shared",
			Fstype:  "ext4",
			Regions: []string{"region0"},
		}
	}
	setDeviceTopology := func(m *model.ResourceYaml) {
		m.Topology = model.Topology{
			Type:    "device",
			Options: "prjquota",
			Fstype:  "ext4",
			Devices: []string{"/dev/vdc", "/dev/vdd"},
		}
	}
	setDeviceNameElement := func(m *model.ResourceYaml) {
		m.Name = "/tmp/foo1"
	}
	testYamls := QPList{QuotaPaths: []model.ResourceYaml{
		*makeResourceYamlCustom(setOpInOperatorElement, setPmemTopology),
		*makeResourceYamlCustom(setDeviceNameElement, setOpInOperatorElement, setDeviceTopology),
	}}

	d, err := yaml.Marshal(&testYamls)
	if err != nil {
		t.Error()
	}
	err = ioutil.WriteFile(configPath, d, 0777)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, resourceManager.AnalyseConfigMap())
	assert.Equal(t, 1, len(resourceManager.RegionQuotaPath))
	assert.Equal(t, 1, len(resourceManager.DeviceQuotaPath))
	gomock.InOrder(
		mockMounter.EXPECT().EnsureFolder(
			gomock.Eq("/tmp/foo1")).Return(nil),
		mockMounter.EXPECT().FileExists(
			gomock.Eq("/dev/vdc")).Return(true),
		mockMounter.EXPECT().FormatAndMount(
			gomock.Eq("/dev/vdc"), gomock.Eq("/tmp/foo1"), gomock.Eq("ext4"), gomock.Eq([]string{"-O", "project,quota"}), gomock.Eq("prjquota")).Return(nil),
		mockPmemer.EXPECT().GetPmemNamespaceDeivcePath(
			gomock.Eq("region0"), gomock.Eq("fsdax")).Return("/dev/pmem0", "", nil),
		mockMounter.EXPECT().EnsureFolder(
			gomock.Eq("/tmp/foo")).Return(nil),
		mockMounter.EXPECT().FormatAndMount(
			gomock.Eq("/dev/pmem0"), gomock.Eq("/tmp/foo"), gomock.Eq("ext4"), gomock.Eq([]string{"-O", "project,quota"}), gomock.Eq("prjquota,shared")).Return(nil),
	)
	assert.Nil(t, resourceManager.ApplyResourceDiff())
}
