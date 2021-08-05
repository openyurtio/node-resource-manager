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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openyurtio/node-resource-manager/pkg/config"
	"github.com/openyurtio/node-resource-manager/pkg/model"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

func makeValidResourceYaml() *model.ResourceYaml {
	return &model.ResourceYaml{
		Name: "foo",
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
	configPath := "/tmp/volumegroup"
	err := EnsureFolder(filepath.Dir(configPath))
	if err != nil {
		return "", err, nil
	}

	newMockVolumegroupResourceManager := func() *ResourceManager {
		return &ResourceManager{
			configPath: configPath,
		}
	}
	return configPath, nil, newMockVolumegroupResourceManager()
}

func TestAnalyseConfigMap(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	// set test node info
	configPath, err, resourceManager := EnsureVolumeGroupEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configPath)
	mockMounter := utils.NewMockMounter(mockCtl)
	resourceManager.mounter = mockMounter
	resourceManager.recorder = record.NewFakeRecorder(10)
	setOpInOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpIn
		m.Value = "foo"
	}
	setOpNotInOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpNotIn
		m.Value = "foo"
	}
	setOpExistsOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpExists
		m.Value = "foo"
	}
	setOpNotExistsOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpExists
		m.Value = "foo"
	}
	setRyNameBar1Element := func(m *model.ResourceYaml) {
		m.Name = "bar1"
	}
	setDeviceTopology := func(m *model.ResourceYaml) {
		m.Topology = model.Topology{
			Type:    "device",
			Devices: []string{"/dev/vdb", "/dev/vdc"},
		}
	}
	setPmemTopology := func(m *model.ResourceYaml) {
		m.Topology = model.Topology{
			Type:    "pmem",
			Regions: []string{"/dev/vdb", "/dev/vdc"},
		}
	}

	testYamls := VgList{VolumeGroups: []model.ResourceYaml{
		*makeResourceYamlCustom(setOpInOperatorElement, setDeviceTopology),
		*makeResourceYamlCustom(setOpNotInOperatorElement),
		*makeResourceYamlCustom(setOpNotExistsOperatorElement),
		*makeResourceYamlCustom(setOpExistsOperatorElement, setRyNameBar1Element, setPmemTopology),
	}}
	d, err := yaml.Marshal(&testYamls)
	defer os.Remove(configPath)
	if err != nil {
		t.Error()
	}
	ioutil.WriteFile(configPath, d, 0777)

	gomock.InOrder(
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdb")).Return(true),
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdc")).Return(false),
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdb")).Return(false),
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdc")).Return(true),
	)

	assert.Nil(t, resourceManager.AnalyseConfigMap())
	assert.Equal(t, 1, len(resourceManager.volumeGroupDeviceMap))
	assert.Equal(t, 1, len(resourceManager.volumeGroupRegionMap))
	assert.Equal(t, 1, len(resourceManager.volumeGroupDeviceMap["foo"].PhysicalVolumes))
	assert.Equal(t, "/dev/vdc", resourceManager.volumeGroupRegionMap["bar1"][0])
}

// EnsureFolder ...
func EnsureFolder(target string) error {
	mdkirCmd := "mkdir"
	_, err := exec.LookPath(mdkirCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return fmt.Errorf("EnsureFolder:: %q executable not found in $PATH", mdkirCmd)
		}
		return err
	}

	mkdirArgs := []string{"-p", target}
	//log.Infof("mkdir for folder, the command is %s %v", mdkirCmd, mkdirArgs)
	_, err = exec.Command(mdkirCmd, mkdirArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("EnsureFolder:: mkdir for folder error: %v", err)
	}
	return nil
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
	resourceManager.pmemer = mockPmemer
	mockLVM := utils.NewMockLVM(mockCtl)
	resourceManager.lvmer = mockLVM
	mockMounter := utils.NewMockMounter(mockCtl)
	resourceManager.mounter = mockMounter
	resourceManager.recorder = record.NewFakeRecorder(10)
	setOpInOperatorElement := func(m *model.ResourceYaml) {
		m.Key = "bar"
		m.Operator = metav1.LabelSelectorOpIn
		m.Value = "foo"
	}
	setVgDeviceTopology := func(m *model.ResourceYaml) {
		m.Topology = model.Topology{
			Type:    "device",
			Devices: []string{"/dev/vdb", "/dev/vdc"},
		}
	}
	setVgDeviceName := func(m *model.ResourceYaml) {
		m.Name = "volumegroup1"
	}

	testYamls := VgList{VolumeGroups: []model.ResourceYaml{
		*makeResourceYamlCustom(setOpInOperatorElement, setVgDeviceTopology, setVgDeviceName),
	}}
	d, err := yaml.Marshal(&testYamls)
	if err != nil {
		t.Error()
	}
	err = ioutil.WriteFile(configPath, d, 0777)
	if err != nil {
		t.Fatal(err)
	}
	prListStr := strings.Join([]string{"/dev/vdb", "/dev/vdc"}, " ")
	gomock.InOrder(
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdb")).Return(true),
		mockMounter.EXPECT().FileExists(gomock.Eq("/dev/vdc")).Return(true),
	)
	assert.Nil(t, resourceManager.AnalyseConfigMap())
	gomock.InOrder(
		mockLVM.EXPECT().ListPhysicalVolume().Return([]*model.PV{}, nil),
		mockLVM.EXPECT().CreateVG(gomock.Eq("volumegroup1"), gomock.Eq(prListStr), gomock.Eq([]string{})).Return("", nil),
	)
	assert.Nil(t, resourceManager.ApplyResourceDiff())
}
