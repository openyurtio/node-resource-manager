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

package model

import (
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeType is volume type
type VolumeType byte

const (
	// separator for vg info
	separator = "<:SEP:>"
)

// types
const (
	VolumeTypeMirrored                  VolumeType = 'm'
	VolumeTypeMirroredWithoutSync       VolumeType = 'M'
	VolumeTypeOrigin                    VolumeType = 'o'
	VolumeTypeOriginWithMergingSnapshot VolumeType = 'O'
	VolumeTypeRAID                      VolumeType = 'r'
	VolumeTypeRAIDWithoutSync           VolumeType = 'R'
	VolumeTypeSnapshot                  VolumeType = 's'
	VolumeTypeMergingSnapshot           VolumeType = 'S'
	VolumeTypePVMove                    VolumeType = 'p'
	VolumeTypeVirtualMirror             VolumeType = 'v'
	VolumeTypeVirtualRaidImage          VolumeType = 'i'
	VolumeTypeRaidImageOutOfSync        VolumeType = 'I'
	VolumeTypeMirrorLog                 VolumeType = 'l'
	VolumeTypeUnderConversion           VolumeType = 'c'
	VolumeTypeThin                      VolumeType = 'V'
	VolumeTypeThinPool                  VolumeType = 't'
	VolumeTypeThinPoolData              VolumeType = 'T'
	VolumeTypeRaidOrThinPoolMetadata    VolumeType = 'e'
)

// VolumePermissions is volume permissions
type VolumePermissions rune

// permissions
const (
	VolumePermissionsWriteable          VolumePermissions = 'w'
	VolumePermissionsReadOnly           VolumePermissions = 'r'
	VolumePermissionsReadOnlyActivation VolumePermissions = 'R'
)

// VolumeAllocation is volume allocation policy
type VolumeAllocation rune

// allocations
const (
	VolumeAllocationAnywhere         VolumeAllocation = 'a'
	VolumeAllocationContiguous       VolumeAllocation = 'c'
	VolumeAllocationInherited        VolumeAllocation = 'i'
	VolumeAllocationCling            VolumeAllocation = 'l'
	VolumeAllocationNormal           VolumeAllocation = 'n'
	VolumeAllocationAnywhereLocked   VolumeAllocation = 'A'
	VolumeAllocationContiguousLocked VolumeAllocation = 'C'
	VolumeAllocationInheritedLocked  VolumeAllocation = 'I'
	VolumeAllocationClingLocked      VolumeAllocation = 'L'
	VolumeAllocationNormalLocked     VolumeAllocation = 'N'
)

// VolumeFixedMinor is volume fixed minor
type VolumeFixedMinor rune

// fixed minor
const (
	VolumeFixedMinorEnabled  VolumeFixedMinor = 'm'
	VolumeFixedMinorDisabled VolumeFixedMinor = '-'
)

func (t VolumeFixedMinor) toProto() bool {
	return t == VolumeFixedMinorEnabled
}

// VolumeState is volume state
type VolumeState rune

// states
const (
	VolumeStateActive                               VolumeState = 'a'
	VolumeStateSuspended                            VolumeState = 's'
	VolumeStateInvalidSnapshot                      VolumeState = 'I'
	VolumeStateInvalidSuspendedSnapshot             VolumeState = 'S'
	VolumeStateSnapshotMergeFailed                  VolumeState = 'm'
	VolumeStateSuspendedSnapshotMergeFailed         VolumeState = 'M'
	VolumeStateMappedDevicePresentWithoutTables     VolumeState = 'd'
	VolumeStateMappedDevicePresentWithInactiveTable VolumeState = 'i'
)

// VolumeOpen is volume open
type VolumeOpen rune

// open
const (
	VolumeOpenIsOpen    VolumeOpen = 'o'
	VolumeOpenIsNotOpen VolumeOpen = '-'
)

// VolumeTargetType is volume taget type
type VolumeTargetType rune

// target type
const (
	VolumeTargetTypeMirror   VolumeTargetType = 'm'
	VolumeTargetTypeRAID     VolumeTargetType = 'r'
	VolumeTargetTypeSnapshot VolumeTargetType = 's'
	VolumeTargetTypeThin     VolumeTargetType = 't'
	VolumeTargetTypeUnknown  VolumeTargetType = 'u'
	VolumeTargetTypeVirtual  VolumeTargetType = 'v'
)

// VolumeZeroing is volume zeroing
type VolumeZeroing rune

// zeroing
const (
	VolumeZeroingIsZeroing    VolumeZeroing = 'z'
	VolumeZeroingIsNonZeroing VolumeZeroing = '-'
)

// VolumeHealth is volume health
type VolumeHealth rune

// health
const (
	VolumeHealthOK              VolumeHealth = '-'
	VolumeHealthPartial         VolumeHealth = 'p'
	VolumeHealthRefreshNeeded   VolumeHealth = 'r'
	VolumeHealthMismatchesExist VolumeHealth = 'm'
	VolumeHealthWritemostly     VolumeHealth = 'w'
)

// VolumeActivationSkipped is volume activation
type VolumeActivationSkipped rune

// activation
const (
	VolumeActivationSkippedIsSkipped    VolumeActivationSkipped = 's'
	VolumeActivationSkippedIsNotSkipped VolumeActivationSkipped = '-'
)

func (t VolumeActivationSkipped) toProto() bool {
	return t == VolumeActivationSkippedIsSkipped
}

// ResourceYaml ...
type ResourceYaml struct {
	Name     string                       `yaml:"name,omitempty"`
	Key      string                       `yaml:"key,omitempty"`
	Operator metav1.LabelSelectorOperator `yaml:"operator,omitempty"`
	Value    string                       `yaml:"value,omitempty"`
	Topology Topology                     `yaml:"topology,omitempty"`
}

// Topology ...
type Topology struct {
	Type    string `yaml:"type,omitempty"`
	Options string `yaml:"options,omitempty"`
	Fstype  string `yaml:"fstype,omitempty"`

	Devices []string            `yaml:"devices,omitempty"`
	Volumes []map[string]string `yaml:"volumes,omitempty"`
	Regions []string            `yaml:"regions,omitempty"`
}

// PmemRegions list all regions
type PmemRegions struct {
	Regions []PmemRegion `json:"regions"`
}

// PmemRegion define on pmem region
type PmemRegion struct {
	Dev               string          `json:"dev"`
	Size              int64           `json:"size,omitempty"`
	AvailableSize     int64           `json:"available_size,omitempty"`
	MaxAvailableExent int64           `json:"max_available_extent,omitempty"`
	RegionType        string          `json:"type,omitempty"`
	IsetID            int64           `json:"iset_id,omitempty"`
	PersistenceDomain string          `json:"persistence_domain,omitempty"`
	Namespaces        []PmemNameSpace `json:"namespaces,omitempty"`
}

// PmemNameSpace define one pmem namespaces
type PmemNameSpace struct {
	Dev        string `json:"dev,omitempty"`
	Mode       string `json:"mode,omitempty"`
	MapType    string `json:"map,omitempty"`
	Size       int64  `json:"size,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	SectorSize int64  `json:"sectorsize,omitempty"`
	Align      int64  `json:"align,omitempty"`
	BlockDev   string `json:"blockdev,omitempty"`
	CharDev    string `json:"chardev,omitempty"`
	Name       string `json:"name,omitempty"`
}

// DaxctrlMem list all mems
type DaxctrlMem struct {
	Chardev    string `json:"chardev"`
	Size       int64  `json:"size"`
	TargetNode int    `json:"target_node"`
	Mode       string `json:"mode"`
	Movable    bool   `json:"movable"`
}

// LV is a logical volume
type LV struct {
	Name               string
	Size               uint64
	UUID               string
	Attributes         LVAttributes
	CopyPercent        string
	ActualDevMajNumber uint32
	ActualDevMinNumber uint32
	Tags               []string
}

// VG is volume group
type VG struct {
	Name     string
	Size     uint64
	FreeSize uint64
	UUID     string
	Tags     []string
}

// PV is Physica lVolume
type PV struct {
	Name   string
	VgName string
	Size   uint64
	UUID   string
}

// LVAttributes is attributes
type LVAttributes struct {
	Type              VolumeType
	Permissions       VolumePermissions
	Allocation        VolumeAllocation
	FixedMinor        VolumeFixedMinor
	State             VolumeState
	Open              VolumeOpen
	TargetType        VolumeTargetType
	Zeroing           VolumeZeroing
	Health            VolumeHealth
	ActivationSkipped VolumeActivationSkipped
}

// ParseLV ...
func ParseLV(line string) (*LV, error) {
	// lvs --units=b --separator="<:SEP:>" --nosuffix --noheadings -o lv_name,lv_size,lv_uuid,lv_attr,copy_percent,lv_kernel_major,lv_kernel_minor,lv_tags --nameprefixes -a
	// todo: devices, lv_ancestors, lv_descendants, lv_major, lv_minor, mirror_log, modules, move_pv, origin, region_size
	//       seg_count, seg_size, seg_start, seg_tags, segtype, snap_percent, stripes, stripe_size
	fields, err := parse(line, 8)
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseUint(fields["LVM2_LV_SIZE"], 10, 64)
	if err != nil {
		return nil, err
	}

	kernelMajNumber, err := strconv.ParseUint(fields["LVM2_LV_KERNEL_MAJOR"], 10, 32)
	if err != nil {
		return nil, err
	}

	kernelMinNumber, err := strconv.ParseUint(fields["LVM2_LV_KERNEL_MINOR"], 10, 32)
	if err != nil {
		return nil, err
	}

	attrs, err := parseAttrs(fields["LVM2_LV_ATTR"])
	if err != nil {
		return nil, err
	}

	return &LV{
		Name:               fields["LVM2_LV_NAME"],
		Size:               size,
		UUID:               fields["LVM2_LV_UUID"],
		Attributes:         *attrs,
		CopyPercent:        fields["LVM2_COPY_PERCENT"],
		ActualDevMajNumber: uint32(kernelMajNumber),
		ActualDevMinNumber: uint32(kernelMinNumber),
		Tags:               strings.Split(fields["LVM2_LV_TAGS"], ","),
	}, nil

}

// ParseVG parse volume group
func ParseVG(line string) (*VG, error) {
	// vgs --units=b --separator="<:SEP:>" --nosuffix --noheadings -o vg_name,vg_size,vg_free,vg_uuid,vg_tags --nameprefixes -a
	fields, err := parse(line, 5)
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseUint(fields["LVM2_VG_SIZE"], 10, 64)
	if err != nil {
		return nil, err
	}

	freeSize, err := strconv.ParseUint(fields["LVM2_VG_FREE"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &VG{
		Name:     fields["LVM2_VG_NAME"],
		Size:     size,
		FreeSize: freeSize,
		UUID:     fields["LVM2_VG_UUID"],
		Tags:     strings.Split(fields["LVM2_VG_TAGS"], ","),
	}, nil
}

// ParsePV parse volume group
func ParsePV(line string) (*PV, error) {
	// vgs --units=b --separator="<:SEP:>" --nosuffix --noheadings -o vg_name,vg_size,vg_free,vg_uuid,vg_tags --nameprefixes -a
	fields, err := parse(line, 4)
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseUint(fields["LVM2_PV_SIZE"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &PV{
		Name:   fields["LVM2_PV_NAME"],
		VgName: fields["LVM2_VG_NAME"],
		Size:   size,
		UUID:   fields["LVM2_PV_UUID"],
	}, nil
}

func parse(line string, numComponents int) (map[string]string, error) {
	components := strings.Split(line, separator)
	if len(components) != numComponents {
		return nil, fmt.Errorf("expected %d components, got %d", numComponents, len(components))
	}

	fields := map[string]string{}
	for _, c := range components {
		idx := strings.Index(c, "=")
		if idx == -1 {
			return nil, fmt.Errorf("failed to parse component '%s'", c)
		}
		key := c[0:idx]
		value := c[idx+1:]
		if len(value) < 2 {
			return nil, fmt.Errorf("failed to parse component '%s'", c)
		}
		if value[0] != '\'' || value[len(value)-1] != '\'' {
			return nil, fmt.Errorf("failed to parse component '%s'", c)
		}
		value = value[1 : len(value)-1]
		fields[key] = value
	}

	return fields, nil
}

func parseAttrs(attrs string) (*LVAttributes, error) {
	if len(attrs) != 10 {
		return nil, fmt.Errorf("incorrect attrs block size, expected 10, got %d in %s", len(attrs), attrs)
	}

	ret := &LVAttributes{}
	ret.Type = VolumeType(attrs[0])
	ret.Permissions = VolumePermissions(attrs[1])
	ret.Allocation = VolumeAllocation(attrs[2])
	ret.FixedMinor = VolumeFixedMinor(attrs[3])
	ret.State = VolumeState(attrs[4])
	ret.Open = VolumeOpen(attrs[5])
	ret.TargetType = VolumeTargetType(attrs[6])
	ret.Zeroing = VolumeZeroing(attrs[7])
	ret.Health = VolumeHealth(attrs[8])
	ret.ActivationSkipped = VolumeActivationSkipped(attrs[9])

	return ret, nil
}
