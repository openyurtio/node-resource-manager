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

package memory

import (
	"io/ioutil"
	"os"

	"github.com/openyurtio/node-resource-manager/pkg/config"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
)

// ResourceManager ...
type ResourceManager struct {
	Memory     []*MConfig
	pmem       utils.Pmemer
	configPath string
	recorder   record.EventRecorder
}

// NewResourceManager ...
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		Memory:     []*MConfig{},
		pmem:       utils.NewNodePmemer(),
		configPath: "/etc/unified-config/memory",
		recorder:   utils.NewEventRecorder(),
	}
}

// AnalyseConfigMap analyse memory resource config
func (mrm *ResourceManager) AnalyseConfigMap() error {
	memoryConfig := []*MConfig{}
	memoryList := &MList{}

	yamlFile, err := ioutil.ReadFile(mrm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Errorf("memory config file %s not exist", mrm.configPath)
			return nil
		}
		klog.Errorf("AnalyseConfigMap:: yamlFile.Get memory error %v", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, memoryList)
	if err != nil {
		klog.Errorf("AnalyseConfigMap:: parse yaml file error: %v", err)
		return err
	}

	nodeInfo := config.GlobalConfigVar.NodeInfo
	for _, memConfig := range memoryList.Memories {
		isMatched := utils.NodeFilter(memConfig.Operator, memConfig.Key, memConfig.Value, nodeInfo)
		if isMatched {
			conf := &MConfig{}
			if len(memConfig.Topology.Regions) != 1 {
				klog.Errorf("AnalyseConfigMap:: regions has multi config %s", memConfig.Topology.Regions)
				continue
			}
			conf.Region = memConfig.Topology.Regions[0]
			conf.Type = memConfig.Topology.Type
			memoryConfig = append(memoryConfig, conf)
		}
	}
	mrm.Memory = memoryConfig
	return nil
}

// ApplyResourceDiff apply memory resource to current node
func (mrm *ResourceManager) ApplyResourceDiff() error {
	klog.Infof("ApplyResourceDiff: matched node resources mrm.Memory: %v", mrm.Memory)
	for _, memConfig := range mrm.Memory {
		devicePath, _, err := mrm.pmem.GetPmemNamespaceDeivcePath(memConfig.Region, "devdax")
		if err != nil {
			err := mrm.pmem.CreateNamespace(memConfig.Region, "dax")
			if err != nil {
				klog.Errorf("applyResourceDiff:: create kmem namespace for region [%s], error: %v", memConfig.Region, err)
				continue
			}
			devicePath, _, err = mrm.pmem.GetPmemNamespaceDeivcePath(memConfig.Region, "devdax")
			if err != nil {
				klog.Errorf("applyResourceDiff:: list kmem namespace for region [%s], error: %v", memConfig.Region, err)
				continue
			}
		}
		isCreated, err := mrm.pmem.CheckKMEMCreated(devicePath[5:])
		if err != nil {
			klog.Errorf("applyResourceDiff:: check kmem create error: %v", err)
			continue
		}
		if !isCreated {
			err := mrm.pmem.MakeNamespaceMemory(devicePath[5:])
			if err != nil {
				klog.Errorf("applyRegionQuotaPath:: make kmem memory failed %v", err)
				continue
			}
		}
	}
	return nil
}
