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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/openyurtio/node-resource-manager/pkg/model"
	log "github.com/sirupsen/logrus"
)

// Pmemer ...
type Pmemer interface {
	GetRegions() (*model.PmemRegions, error)
	CreateNamespace(string, string) error
	CheckNamespaceUsed(string) bool
	GetPmemNamespaceDeivcePath(string, string) (string, string, error)
	MakeNamespaceMemory(chardev string) error
	CheckKMEMCreated(chardev string) (bool, error)
}

// NodePmemer ...
type NodePmemer struct {
}

// NewNodePmemer create new NodePmemer struct
func NewNodePmemer() *NodePmemer {
	return &NodePmemer{}
}

// GetRegions ...
func (np *NodePmemer) GetRegions() (*model.PmemRegions, error) {
	regions := &model.PmemRegions{}
	getRegionCmd := fmt.Sprintf("%s ndctl list -RN", NsenterCmd)
	regionOut, err := Run(getRegionCmd)
	if err != nil {
		return regions, err
	}
	err = json.Unmarshal(([]byte)(regionOut), regions)
	if err != nil {
		if strings.HasPrefix(regionOut, "[") {
			regionList := []model.PmemRegion{}
			err = json.Unmarshal(([]byte)(regionOut), &regionList)
			if err != nil {
				return regions, err
			}
			regions.Regions = regionList
		} else {
			return regions, err
		}
	}
	return regions, nil
}

// CreateNamespace ...
func (np *NodePmemer) CreateNamespace(region, pmemType string) error {
	var createCmd string
	if pmemType == "lvm" {
		createCmd = fmt.Sprintf("%s ndctl create-namespace -r %s", NsenterCmd, region)
	} else {
		createCmd = fmt.Sprintf("%s ndctl create-namespace -r %s --mode=devdax", NsenterCmd, region)
	}
	_, err := Run(createCmd)
	if err != nil {
		log.Errorf("Create NameSpace for region %s error: %v", region, err)
		return err
	}
	log.Infof("Create NameSpace for region %s successful", region)
	return nil
}

// CheckNamespaceUsed device used in block
func (np *NodePmemer) CheckNamespaceUsed(devicePath string) bool {
	pvCheckCmd := fmt.Sprintf("%s pvs %s 2>&1 | grep -v \"Failed to \" | grep /dev | awk '{print $2}' | wc -l", NsenterCmd, devicePath)
	out, err := Run(pvCheckCmd)
	if err == nil && strings.TrimSpace(out) != "0" {
		log.Infof("CheckNamespaceUsed: NameSpace %s used for pv", devicePath)
		return true
	}

	out, err = checkFSType(devicePath)
	if err == nil && strings.TrimSpace(out) != "" {
		log.Infof("CheckNamespaceUsed: NameSpace %s format as %s", devicePath, out)
		return true
	}
	return false
}

// GetPmemNamespaceDeivcePath ...
func (np *NodePmemer) GetPmemNamespaceDeivcePath(region, mode string) (devicePath string, namespaceName string, err error) {
	regions, err := np.getRegionNamespaceInfo(region)
	if err != nil {
		return "", "", err
	}
	namespace := regions.Regions[0].Namespaces[0]
	if namespace.Mode != mode {
		log.Errorf("GetPmemNamespaceDeivcePath namespace mode %s wrong with: %s", namespace.Mode, mode)
		return "", "", errors.New("GetPmemNamespaceDeivcePath pmem namespace wrong mode" + namespace.Mode)
	}
	if mode == "fsdax" {
		devicePath = "/dev/" + namespace.BlockDev
	} else {
		devicePath = "/dev/" + namespace.CharDev
	}
	return devicePath, namespace.Dev, nil
}

func (np *NodePmemer) getRegionNamespaceInfo(region string) (*model.PmemRegions, error) {
	listCmd := fmt.Sprintf("%s ndctl list -RN -r %s", NsenterCmd, region)

	out, err := Run(listCmd)
	if err != nil {
		log.Errorf("List NameSpace for region %s error: %v", region, err)
		return nil, err
	}
	regions := &model.PmemRegions{}
	err = json.Unmarshal(([]byte)(out), regions)
	if err != nil {
		log.Errorf("getRegionNamespaceInfo:: unmarshal regions err: %v", err)
		return nil, err
	}
	if len(regions.Regions) == 0 {
		log.Errorf("list Namespace for region %s get 0 region, out: %s", region, out)
		return nil, errors.New("list Namespace get 0 region by " + region)
	}

	if len(regions.Regions[0].Namespaces) != 1 {
		log.Errorf("list Namespace for region %s get 0 or multi namespaces", region)
		return nil, errors.New("list Namespace for region get 0 or multi namespaces" + region)
	}
	return regions, nil
}

// MakeNamespaceMemory ...
func (np *NodePmemer) MakeNamespaceMemory(chardev string) error {
	makeCmd := fmt.Sprintf("%s daxctl reconfigure-device -m system-ram %s", NsenterCmd, chardev)
	_, err := Run(makeCmd)
	return err
}

// CheckKMEMCreated ...
func (np *NodePmemer) CheckKMEMCreated(chardev string) (bool, error) {
	listCmd := fmt.Sprintf("%s daxctl list", NsenterCmd)
	out, err := Run(listCmd)
	if err != nil {
		log.Errorf("CheckKMEMCreated:: List daxctl error: %v", err)
		return false, err
	}
	memList := []*model.DaxctrlMem{}
	err = json.Unmarshal(([]byte)(out), &memList)
	if err != nil {
		return false, err
	}
	for _, mem := range memList {
		if mem.Chardev == chardev && mem.Mode == "system-ram" {
			return true, nil
		}
	}
	return false, nil
}
