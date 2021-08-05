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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/openyurtio/node-resource-manager/pkg/config"
	CusErr "github.com/openyurtio/node-resource-manager/pkg/err"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const (
	StorageCmNameSpace = "kube-system"
	StorageCmName      = "cm-node-resource"
	VGConfigKey        = "volumegroup.json"
	AliyunLocalDisk    = "aliyun-local-disk"

	VgConfigFile = "/etc/unified-config/volumegroup"
	VgTypeDevice = "device"
	VgTypePvc    = "pvc"
	VgTypeLocal  = "alibabacloud-local-disk"
	VgTypePmem   = "pmem"
)

// ResourceManager ...
type ResourceManager struct {
	volumeGroupDeviceMap map[string]*VgDeviceConfig
	volumeGroupRegionMap map[string][]string
	mounter              utils.Mounter
	pmemer               utils.Pmemer
	lvmer                utils.LVM
	configPath           string
	recorder             record.EventRecorder
}

// NewResourceManager ...
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		volumeGroupDeviceMap: make(map[string]*VgDeviceConfig),
		volumeGroupRegionMap: make(map[string][]string),
		pmemer:               utils.NewNodePmemer(),
		mounter:              utils.NewMounter(),
		lvmer:                utils.NewNodeLVM(),
		configPath:           "/etc/unified-config/volumegroup",
		recorder:             utils.NewEventRecorder(),
	}
}

// DeviceChars ...
var DeviceChars = []string{"b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

// Create VolumeGroup
func (vrm *ResourceManager) createVg(vgName string, desirePvList []string) error {
	pvListStr := strings.Join(desirePvList, " ")
	tagList := []string{}
	out, err := vrm.lvmer.CreateVG(vgName, pvListStr, tagList)
	if err != nil {
		log.Errorf("createVg:: Create Vg(%s) error with: %s", vgName, err.Error())
		return err
	}
	log.Infof("createVg:: Successful Create Vg(%s) with out: %s", vgName, out)
	return nil
}

// Upgrade VolumeGroup, only support extend pv;
func (vrm *ResourceManager) updateVg(vgName string, expectPvList, realPvList []string) error {
	if len(expectPvList) < len(realPvList) {
		msg := fmt.Sprintf("updateVg:: VolumeGroup: %s, expected pv list should be more than current pv list when update volume group: %v, %v", vgName, expectPvList, realPvList)
		log.Errorf(msg)
		return errors.New(msg)
	}

	// removed pv: pv exist in current node, but not in expect
	removePv := difference(realPvList, expectPvList)
	if len(removePv) > 0 {
		msg := fmt.Sprintf("updateVg:: VolumeGroup: %s, expected pv list should be more than current pv list when update volume group: expect %v, current %v, and not support remove pv now", vgName, expectPvList, realPvList)
		log.Errorf(msg)
		return errors.New(msg)
	}

	// added pv: pv exist in expect, but not in current node.
	addedPv := difference(expectPvList, realPvList)
	if len(addedPv) == 0 {
		msg := fmt.Sprintf("updateVg:: VolumeGroup: %s, expected pv list same with current pv list, expect %v, current %v", vgName, expectPvList, realPvList)
		log.Errorf(msg)
		return errors.New(msg)
	}

	pvListStr := strings.Join(addedPv, " ")
	_, err := vrm.lvmer.ExtendVG(vgName, pvListStr)
	if err != nil {
		msg := fmt.Sprintf("updateVg:: Extend vg(%s) error: %v", vgName, err)
		log.Errorf(msg)
		return errors.New(msg)
	}
	log.Infof("updateVg:: Successful Add pvs %s to VolumeGroup %s", addedPv, vgName)
	return nil
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func getPvListForLocalDisk(mounter utils.Mounter) []string {
	// Step 2: Get LocalDisk Number
	localDeviceList := []string{}
	localDeviceNum, err := getLocalDeviceNum()
	if err != nil {
		log.Errorf("getPvListForLocalDisk:: LocalDiskMount:: Get Local Disk Number Error, Error: %s", err.Error())
		return localDeviceList
	}
	if localDeviceNum < 1 {
		log.Errorf("getPvListForLocalDisk:: VG not exist and also not local disk exist, localDeivceNum: %v", localDeviceNum)
		return localDeviceList
	}

	// Step 3: Get LocalDisk device
	deviceStartWith := "vdb"
	deviceNamePrefix := "vd"
	deviceStartIndex := 0
	deviceNameLen := len(deviceStartWith)
	if deviceNameLen > 1 {
		deviceStartChar := deviceStartWith[deviceNameLen-1 : deviceNameLen]
		for index := 0; index < 15; index++ {
			if deviceStartChar == DeviceChars[index] {
				deviceStartIndex = index
			}
		}
		deviceNamePrefix = deviceStartWith[0 : deviceNameLen-1]
	}
	if mounter.FileExists(deviceNamePrefix + DeviceChars[deviceStartIndex]) {
		for i := deviceStartIndex; i < localDeviceNum; i++ {
			deviceName := deviceNamePrefix + DeviceChars[i]
			devicePath := filepath.Join("/dev", deviceName)
			localDeviceList = append(localDeviceList, devicePath)
		}
	}
	//	nvmeDevicePattern := "nvme%vn1"
	//	nvmeStartIndex := 0
	//	if mounter.FileExists(fmt.Sprintf(nvmeDevicePattern, nvmeStartIndex)) {
	//		for i :=nvmeStartIndex; i < localDeviceNum; i++ {
	//			devicePath := filepath.Join("/dev", fmt.Sprintf(nvmeDevicePattern, i))
	//			localDeviceList = append(localDeviceList, devicePath)
	//		}
	//	}
	log.Infof("getPvListForLocalDisk:: Starting LocalDisk Mount: LocalDisk Number: %d, LocalDisk: %v", localDeviceNum, localDeviceList)
	return localDeviceList
}

// Get Local Disk Number from ecs API
// Requirements: The instance must have role which contains ecs::DescribeInstances, ecs::DescribeInstancesType.
func getLocalDeviceNum() (int, error) {
	instanceID := GetMetaData(InstanceID)
	regionID := GetMetaData(RegionIDTag)
	localDeviceNum := 0
	akID, akSecret, token := GetDefaultAK()
	client := NewEcsClient(regionID, akID, akSecret, token)

	// Get Instance Type
	request := ecs.CreateDescribeInstancesRequest()
	request.RegionId = regionID
	request.InstanceIds = "[\"" + instanceID + "\"]"
	instanceResponse, err := client.DescribeInstances(request)
	if err != nil {
		log.Errorf("getLocalDeviceNum:: Describe Instance: %s Error: %s", instanceID, err.Error())
		return -1, err
	}
	if instanceResponse == nil || len(instanceResponse.Instances.Instance) == 0 {
		log.Infof("getLocalDeviceNum:: Describe Instance Error, with empty response: %s", instanceID)
		return -1, err
	}

	// Get Instance LocalDisk Number
	instanceTypeID := instanceResponse.Instances.Instance[0].InstanceType
	instanceTypeFamily := instanceResponse.Instances.Instance[0].InstanceTypeFamily
	instanceTypeRequest := ecs.CreateDescribeInstanceTypesRequest()
	instanceTypeRequest.InstanceTypeFamily = instanceTypeFamily
	response, err := client.DescribeInstanceTypes(instanceTypeRequest)
	if err != nil {
		log.Errorf("getLocalDeviceNum:: Describe Instance: %s, Type: %s, Family: %s Error: %s", instanceID, instanceTypeID, instanceTypeFamily, err.Error())
		return -1, err
	}
	for _, instance := range response.InstanceTypes.InstanceType {
		if instance.InstanceTypeId == instanceTypeID {
			localDeviceNum = instance.LocalStorageAmount
			log.Infof("getLocalDeviceNum:: Instance: %s, InstanceType: %s, InstanceLocalDiskNum: %d", instanceID, instanceTypeID, localDeviceNum)
			break
		}
	}
	return localDeviceNum, nil
}

// Get current VolumeGroup in node
// echo vgOjbect contains: vgName, pvList;
func (vrm *ResourceManager) getRealVgList() ([]*VgDeviceConfig, error) {
	physicalVolumeList, err := vrm.lvmer.ListPhysicalVolume()
	if err != nil {
		log.Errorf("List PhysicalVolume get error %v", err)
		return nil, err
	}
	log.Debugf("Real VolumeGroup List: %+v", physicalVolumeList)
	nodeVgList := []*VgDeviceConfig{}
	for _, physicalVolume := range physicalVolumeList {
		isAlreadyAdded := false
		for _, nodeVg := range nodeVgList {
			if nodeVg.Name == physicalVolume.VgName {
				nodeVg.PhysicalVolumes = append(nodeVg.PhysicalVolumes, physicalVolume.Name)
				isAlreadyAdded = true
				log.Debugf("getRealVgList:: physicalVolumes: %v, physicalVolumeName: %s, volumeGroupName: %v", nodeVg.PhysicalVolumes, physicalVolume.Name, physicalVolume.VgName)
			}
		}
		if isAlreadyAdded == false {
			vgPvConfig := &VgDeviceConfig{}
			vgPvConfig.Name = physicalVolume.VgName
			vgPvConfig.PhysicalVolumes = append(vgPvConfig.PhysicalVolumes, physicalVolume.Name)
			nodeVgList = append(nodeVgList, vgPvConfig)
			log.Debugf("Add New VolumeGroupPvConfig: %s, %s, %v", physicalVolume.Name, physicalVolume.VgName, vgPvConfig)
		}
	}
	return nodeVgList, nil
}

// AnalyseConfigMap analyse pmem resource config
func (vrm *ResourceManager) AnalyseConfigMap() error {
	ref := &v1.ObjectReference{
		Kind:      "pods",
		Name:      os.Getenv("POD_NAME"),
		Namespace: "kube-system",
	}
	getExistDevices := func(storages []string) (exists []string) {
		for _, device := range storages {
			cerr := &CusErr.DeviceNotExistsErr{Device: device}
			if !vrm.mounter.FileExists(device) {
				vrm.recorder.Event(ref, v1.EventTypeNormal, "DeviceNotExists", cerr.Error())
			} else {
				exists = append(exists, device)
			}
		}
		return
	}

	vgDeviceMap := map[string]*VgDeviceConfig{}
	vgRegionMap := map[string][]string{}

	volumeGroupList := &VgList{}
	yamlFile, err := ioutil.ReadFile(vrm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("volume config file %s not exist", vrm.configPath)
			return nil
		}
		log.Errorf("AnalyseConfigMap:: ReadFile: yamlFile.Get error #%v ", err)
		return err
	}

	err = yaml.Unmarshal(yamlFile, volumeGroupList)
	if err != nil {
		log.Errorf("AnalyseConfigMap:: Unmarshal: parse yaml file error: %v", err)
		return err
	}

	nodeInfo := config.GlobalConfigVar.NodeInfo
	for _, devConfig := range volumeGroupList.VolumeGroups {
		vgDeviceConfig := &VgDeviceConfig{}

		isMatched := utils.NodeFilter(devConfig.Operator, devConfig.Key, devConfig.Value, nodeInfo)
		log.Infof("AnalyseConfigMap:: isMatched: %v, devConfig: %+v", isMatched, devConfig)

		if isMatched {
			switch devConfig.Topology.Type {
			case VgTypeDevice:
				vgDeviceConfig.PhysicalVolumes = getExistDevices(devConfig.Topology.Devices)
				vgDeviceMap[devConfig.Name] = vgDeviceConfig
			case VgTypeLocal:
				tmpConfig := &VgDeviceConfig{}
				tmpConfig.PhysicalVolumes = getPvListForLocalDisk(vrm.mounter)
				vgDeviceMap[devConfig.Name] = tmpConfig
			case VgTypePvc:
				// not support yet
				continue
			case VgTypePmem:
				vgRegionMap[devConfig.Name] = getExistDevices(devConfig.Topology.Regions)
			default:
				log.Errorf("AnalyseConfigMap:: Get unsupported volumegroup type: %s", devConfig.Topology.Type)
				continue
			}
		}
	}
	vrm.volumeGroupDeviceMap = vgDeviceMap
	vrm.volumeGroupRegionMap = vgRegionMap
	return nil
}

// ApplyResourceDiff apply volume group resource to current node
func (vrm *ResourceManager) ApplyResourceDiff() error {

	// Get Actual VolumeGroup on node.
	actualVgConfig, err := vrm.getRealVgList()
	if err != nil {
		log.Errorf("ApplyResourceDiff:: Get Node Actual VolumeGroup Error: %s", err.Error())
		return err
	}
	if len(vrm.volumeGroupDeviceMap) > 0 {
		vrm.applyDeivce(actualVgConfig)
	}
	if len(vrm.volumeGroupRegionMap) > 0 {
		vrm.applyRegion(actualVgConfig)
	}

	log.Infof("ApplyResourceDiff:: Finish volumegroup loop...")
	return nil
}

func (vrm *ResourceManager) applyDeivce(actualVgConfig []*VgDeviceConfig) error {
	// process each expect volume group
	for expectVgName, expectVg := range vrm.volumeGroupDeviceMap {
		log.Infof("applyDevice:: expectName: %s, expectVgDevices: %v", expectVgName, expectVg.PhysicalVolumes)
		isVgExist := false
		isVgNeedUpdate := false
		realPhysicalVolumeList := []string{}

		for _, realVg := range actualVgConfig {
			if expectVgName == realVg.Name {
				isVgExist = true
				diffs := difference(expectVg.PhysicalVolumes, realVg.PhysicalVolumes)
				if len(diffs) != 0 {
					realPhysicalVolumeList = realVg.PhysicalVolumes
					isVgNeedUpdate = true
				}
				break
			}
		}
		if !isVgExist {
			log.Infof("Create VolumeGroup:: %+v, %+v", expectVgName, expectVg.PhysicalVolumes)
			return vrm.createVg(expectVgName, expectVg.PhysicalVolumes)
		} else if isVgNeedUpdate {
			log.Infof("Update VolumeGroup:: %+v, %+v", expectVgName, expectVg.PhysicalVolumes)
			return vrm.updateVg(expectVgName, expectVg.PhysicalVolumes, realPhysicalVolumeList)
		}
	}
	return nil
}

func (vrm *ResourceManager) applyRegion(actualVgConfig []*VgDeviceConfig) error {
	regions, err := vrm.pmemer.GetRegions()
	if err != nil {
		log.Errorf("applyRegion: get pmem regions error: %v", err)
		return err
	}

	for _, expectRegions := range vrm.volumeGroupRegionMap {
		expectNamespaces := []string{}
		for _, expectRegion := range expectRegions {
			expectRegionExists := false
			for _, region := range regions.Regions {
				if expectRegion == region.Dev {
					expectRegionExists = true
					if len(region.Namespaces) == 0 {
						vrm.pmemer.CreateNamespace(region.Dev, "lvm")
					}
				}
			}
			if !expectRegionExists {
				err := fmt.Errorf("applyRegion:: expect region %s not exists", expectRegion)
				return err
			}
			expectNamespaces = append(expectNamespaces, utils.ConvertRegion2Namespace(expectRegion))
		}
	}
	updatedRegions, err := vrm.pmemer.GetRegions()
	if err != nil {
		log.Errorf("applyRegion: get pmem regions error: %v", err)
		return err
	}
	for expectVgName, expectRegions := range vrm.volumeGroupRegionMap {
		log.Infof("applyDevice:: expectVgName: %v, expectRegions: %v", expectVgName, expectRegions)
		expectLvmInUseDevices := []string{}
		expectLvmNotInUseDevices := []string{}
		for _, expectRegion := range expectRegions {
			devicePath := utils.ConvertNamespace2LVMDevicePath(utils.ConvertRegion2Namespace(expectRegion), updatedRegions)
			if devicePath == "" {
				log.Errorf("applyRegion:: did not get namespace.Blockdev from expectRegion: %s, regions: %v", expectRegion, updatedRegions)
			}
			if vrm.pmemer.CheckNamespaceUsed(devicePath) {
				log.Warnf("NameSpace heen used region: %v, devicePath: %s", expectRegion, devicePath)
				expectLvmInUseDevices = append(expectLvmInUseDevices, devicePath)
			}
			expectLvmNotInUseDevices = append(expectLvmNotInUseDevices, devicePath)
		}

		isVgNeedCreate := true
		for _, actualVg := range actualVgConfig {
			if expectVgName == actualVg.Name {
				isVgNeedCreate = false
				if otherUsage := difference(expectLvmInUseDevices, actualVg.PhysicalVolumes); len(otherUsage) != 0 {
					log.Errorf("applyRegion:: device [%s] is used in other usage", otherUsage)
					break
				}
				updatePvs := difference(expectLvmNotInUseDevices, actualVg.PhysicalVolumes)
				if len(updatePvs) == 0 {
					break
				}
				vrm.updatePmemVg(expectVgName, expectLvmNotInUseDevices)
			}
		}
		if isVgNeedCreate {
			if len(expectLvmInUseDevices) != 0 {
				log.Errorf("applyRegion:: attempt to use inused devices [%s] to create volumegroup", expectLvmInUseDevices)
				continue
			}
			vrm.createVg(expectVgName, expectLvmNotInUseDevices)
		}
	}

	return nil
}

func (vrm *ResourceManager) updatePmemVg(vgName string, addedPv []string) error {

	pvListStr := strings.Join(addedPv, " ")
	_, err := vrm.lvmer.ExtendVG(vgName, pvListStr)
	if err != nil {
		msg := fmt.Sprintf("updatePmemVg:: Extend vg(%s) error: %v", vgName, err)
		log.Errorf(msg)
		return errors.New(msg)
	}
	log.Infof("updatePmemVg:: Successful Add pvs %s to VolumeGroup %s", addedPv, vgName)
	return nil
}
