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
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/openyurtio/node-resource-manager/pkg/config"
	CusErr "github.com/openyurtio/node-resource-manager/pkg/err"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
)

// ResourceManager ...
type ResourceManager struct {
	DeviceQuotaPath map[string]*QpConfig
	RegionQuotaPath map[string]*QpConfig
	mounter         utils.Mounter
	mkfsOption      []string
	pmemer          utils.Pmemer
	configPath      string
	recorder        record.EventRecorder
}

// NewResourceManager ...
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		DeviceQuotaPath: make(map[string]*QpConfig),
		RegionQuotaPath: make(map[string]*QpConfig),
		mounter:         utils.NewMounter(),
		pmemer:          utils.NewNodePmemer(),
		configPath:      "/etc/unified-config/quotapath",
		recorder:        utils.NewEventRecorder(),
	}
}

// AnalyseConfigMap analyse quotapath resource config
func (qrm *ResourceManager) AnalyseConfigMap() error {
	deviceQuotaConfig := map[string]*QpConfig{}
	regionQuotaConfig := map[string]*QpConfig{}
	quotaPathList := &QPList{}
	yamlFile, err := ioutil.ReadFile(qrm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Errorf("quota config file %s not exist", qrm.configPath)
			return nil
		}
		klog.Errorf("AnalyseConfigMap:: yamlFile.Get error %v", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, quotaPathList)
	if err != nil {
		klog.Errorf("AnalyseConfigMap:: parse yaml file error: %v", err)
		return err
	}
	mountPathMap := map[string]string{}
	nodeInfo := config.GlobalConfigVar.NodeInfo
	for _, quotaConfig := range quotaPathList.QuotaPaths {
		isMatched := utils.NodeFilter(quotaConfig.Operator, quotaConfig.Key, quotaConfig.Value, nodeInfo)
		if isMatched {
			if _, ok := mountPathMap[quotaConfig.Name]; ok {
				klog.Errorf("AnalyseConfigMap:: quotapath config has multi mount path config [%s] on same node", quotaConfig.Name)
				continue
			}
			switch quotaConfig.Topology.Type {
			case "device":
				conf := &QpConfig{}
				conf.Devices = quotaConfig.Topology.Devices
				conf.Fstype = quotaConfig.Topology.Fstype
				conf.Options = quotaConfig.Topology.Options
				conf.Type = quotaConfig.Topology.Type
				deviceQuotaConfig[quotaConfig.Name] = conf
			case "pmem":
				conf := &QpConfig{}
				if len(quotaConfig.Topology.Regions) != 1 {
					klog.Errorf("AnalyseConfigMap:: quotapath regions [%s] config only support one device", quotaConfig.Topology.Regions)
					continue
				}
				conf.Region = quotaConfig.Topology.Regions[0]
				conf.Fstype = quotaConfig.Topology.Fstype
				conf.Options = quotaConfig.Topology.Options
				conf.Type = quotaConfig.Topology.Type
				regionQuotaConfig[quotaConfig.Name] = conf
			default:
				klog.Errorf("AnalyseConfigMap:: not support quotapath config type: [%v]", quotaConfig.Topology.Type)
				continue
			}
			mountPathMap[quotaConfig.Name] = ""
		}
	}

	qrm.DeviceQuotaPath = deviceQuotaConfig
	qrm.RegionQuotaPath = regionQuotaConfig
	return nil
}

// ApplyResourceDiff apply quotapath resource to current node
func (qrm *ResourceManager) ApplyResourceDiff() error {
	klog.Infof("ApplyResourceDiff: matched node resources qrm.DeviceQuotaPath: %v, qrm.RegionQuotaPath: %v", qrm.DeviceQuotaPath, qrm.RegionQuotaPath)
	qrm.mkfsOption = strings.Split("-O project,quota", " ")
	err := qrm.applyDeivceQuotaPath()
	if err != nil {
		klog.Errorf("ApplyResourceDiff:: apply deivce quotapath error: %v", err)
	}
	err = qrm.applyRegionQuotaPath()
	if err != nil {
		klog.Errorf("ApplyResourceDiff:: apply region quotapath error: %v", err)
	}
	return err
}

func (qrm *ResourceManager) applyDeivceQuotaPath() error {

	ref := &v1.ObjectReference{
		Kind:      "pods",
		Name:      os.Getenv("POD_NAME"),
		Namespace: "kube-system",
	}
	for mountPath, deivceQuotaPathConfig := range qrm.DeviceQuotaPath {
		err := qrm.mounter.EnsureFolder(mountPath)
		if err != nil {
			klog.Errorf("applyDeivceQuotaPath:: ensure quotapath error: %v", err)
			continue
		}
		klog.Infof("applyDeivceQuotaPath:: device quotapath config devices: %v", deivceQuotaPathConfig.Devices)
		for _, device := range deivceQuotaPathConfig.Devices {
			if !qrm.mounter.FileExists(device) {
				klog.Errorf("applyDeivceQuotaPath:: device %v not exists", device)
				continue
			}
			err = qrm.mounter.FormatAndMount(device, mountPath, deivceQuotaPathConfig.Fstype, qrm.mkfsOption, deivceQuotaPathConfig.Options)
			if err != nil {
				if errors.Is(err, &CusErr.ExistsFormatErr{}) {
					qrm.recorder.Event(ref, v1.EventTypeWarning, "ExistsFormatErr", err.Error())
				}
				klog.Errorf("applyDeivceQuotaPath:: device: %v, mounter FormatAndMount error: %v", device, err)
				continue
			}
			break
		}
	}
	return nil
}

func (qrm *ResourceManager) applyRegionQuotaPath() error {
	for mountPath, regionQuotaPathConfig := range qrm.RegionQuotaPath {
		devicePath, _, err := qrm.pmemer.GetPmemNamespaceDeivcePath(regionQuotaPathConfig.Region, "fsdax")
		if err != nil {
			if strings.Contains(err.Error(), "list Namespace for region get 0 or multi namespaces") {
				err := qrm.pmemer.CreateNamespace(regionQuotaPathConfig.Region, "lvm")
				if err != nil {
					klog.Errorf("applyRegionQuotaPath:: create namespace for region [%s], error: %v", regionQuotaPathConfig.Region, err)
					continue
				}
				devicePath, _, err = qrm.pmemer.GetPmemNamespaceDeivcePath(regionQuotaPathConfig.Region, "fsdax")
				if err != nil {
					klog.Errorf("applyRegionQuotaPath:: get namespace device path for region [%s], error: %v", regionQuotaPathConfig.Region, err)
					continue
				}
			} else {
				klog.Errorf("applyRegionQuotaPath:: get region [%s] namespace device path error: %v", regionQuotaPathConfig.Region, err)
				continue
			}
		}
		err = qrm.mounter.EnsureFolder(mountPath)
		if err != nil {
			klog.Errorf("applyRegionQuotaPath:: ensure quotapath error: %v", err)
			continue
		}
		err = qrm.mounter.FormatAndMount(devicePath, mountPath, regionQuotaPathConfig.Fstype, qrm.mkfsOption, regionQuotaPathConfig.Options)
		if err != nil {
			klog.Errorf("applyRegionQuotaPath:: mounter FormatAndMount error: %v", err)
			continue
		}
	}
	return nil
}
