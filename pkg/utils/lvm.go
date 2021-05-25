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
	"errors"
	"fmt"
	"strings"

	"github.com/openyurtio/node-resource-manager/pkg/model"
)

const (
	ProtectedTagName = "protected"
)

// LVM ...
type LVM interface {
	ListLV(listSpec string) ([]*model.LV, error)
	CreateLV(vg, name string, size uint64, mirrors uint32, tags []string) (string, error)
	RemoveLV(vg, name string) (string, error)
	CloneLV(src, dest string) (string, error)
	ListVG() ([]*model.VG, error)
	ListPhysicalVolume() ([]*model.PV, error)
	CreateVG(name, physicalVolume string, tags []string) (string, error)
	ExtendVG(name, physicalVolume string) (string, error)
	RemoveVG(name string) (string, error)
	AddTagLV(vg, name string, tags []string) (string, error)
	RemoveTagLV(vg, name string, tags []string) (string, error)
}

// NodeLVM ...
type NodeLVM struct {
}

// NewNodeLVM ...
func NewNodeLVM() *NodeLVM {
	return &NodeLVM{}
}

// ListLV ...
func (nl *NodeLVM) ListLV(listSpec string) ([]*model.LV, error) {
	cmdList := []string{NsenterCmd, "lvs", "--units=b", "--separator=\"<:SEP:>\"", "--nosuffix", "--noheadings",
		"-o", "lv_name,lv_size,lv_uuid,lv_attr,copy_percent,lv_kernel_major,lv_kernel_minor,lv_tags", "--nameprefixes", "-a", listSpec}
	cmd := strings.Join(cmdList, " ")
	out, err := Run(cmd)
	if err != nil {
		return nil, err
	}
	outStr := strings.TrimSpace(string(out))
	if outStr == "" {
		lvs := make([]*model.LV, 0)
		return lvs, nil
	}
	outLines := strings.Split(outStr, "\n")
	lvs := []*model.LV{}
	for _, line := range outLines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "LVM2_LV_NAME") {
			continue
		}
		lv, err := model.ParseLV(line)
		if err != nil {
			return nil, errors.New("Parse LVM: " + line + ", with error: " + err.Error())
		}
		lvs = append(lvs, lv)
	}
	return lvs, nil
}

// CreateLV ...
func (nl *NodeLVM) CreateLV(vg, name string, size uint64, mirrors uint32, tags []string) (string, error) {
	if size == 0 {
		return "", errors.New("size must be greater than 0")
	}

	args := []string{"lvcreate", "-v", "-n", name, "-L", fmt.Sprintf("%db", size)}
	if mirrors > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", mirrors), "--nosync")
	}
	for _, tag := range tags {
		args = append(args, "--add-tag", tag)
	}

	args = append(args, vg)
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)
	return string(out), err
}

// RemoveLV ...
func (nl *NodeLVM) RemoveLV(vg, name string) (string, error) {
	lvs, err := nl.ListLV(fmt.Sprintf("%s/%s", vg, name))
	if err != nil {
		return "", fmt.Errorf("failed to list LVs: %v", err)
	}
	if len(lvs) == 0 {
		return "lvm " + vg + "/" + name + " is not exist, skip remove", nil
	}
	if len(lvs) != 1 {
		return "", fmt.Errorf("expected 1 LV, got %d", len(lvs))
	}
	for _, tag := range lvs[0].Tags {
		if tag == ProtectedTagName {
			return "", errors.New("volume is protected")
		}
	}

	args := []string{NsenterCmd, "lvremove", "-v", "-f", fmt.Sprintf("%s/%s", vg, name)}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err

}

// CloneLV ...
func (nl *NodeLVM) CloneLV(src, dest string) (string, error) {
	args := []string{NsenterCmd, "dd", fmt.Sprintf("if=%s", src), fmt.Sprintf("of=%s", dest), "bs=4M"}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err
}

// ListVG ...
func (nl *NodeLVM) ListVG() ([]*model.VG, error) {
	args := []string{NsenterCmd, "vgs", "--units=b", "--separator=\"<:SEP:>\"", "--nosuffix", "--noheadings",
		"-o", "vg_name,vg_size,vg_free,vg_uuid,vg_tags", "--nameprefixes", "-a"}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)
	if err != nil {
		return nil, err
	}
	vgs := []*model.VG{}
	outStr := strings.TrimSpace(string(out))
	if outStr == "" {
		return vgs, nil
	}
	outLines := strings.Split(outStr, "\n")
	for _, line := range outLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "WARNING") {
			continue
		}
		vg, err := model.ParseVG(line)
		if err != nil {
			return nil, err
		}
		vgs = append(vgs, vg)
	}
	return vgs, nil
}

// ListPhysicalVolume ...
func (nl *NodeLVM) ListPhysicalVolume() ([]*model.PV, error) {
	args := []string{NsenterCmd, "pvs", "--units=b", "--separator=\"<:SEP:>\"", "--nosuffix", "--noheadings",
		"-o", "vg_name,pv_name,pv_size,pv_uuid", "--nameprefixes", "-a"}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)
	if err != nil {
		return nil, err
	}
	outStr := strings.TrimSpace(string(out))
	pvs := []*model.PV{}
	if outStr == "" {
		return pvs, nil
	}
	outLines := strings.Split(outStr, "\n")
	for _, line := range outLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "WARNING") {
			continue
		}
		pv, err := model.ParsePV(line)
		if err != nil {
			return nil, err
		}
		if pv.VgName != "" && pv.Name != "" {
			pvs = append(pvs, pv)
		}
	}
	return pvs, nil
}

// CreateVG ...
func (nl *NodeLVM) CreateVG(name, physicalVolume string, tags []string) (string, error) {
	args := []string{NsenterCmd, "vgcreate", name, physicalVolume, "-v"}
	for _, tag := range tags {
		args = append(args, "--add-tag", tag)
	}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err
}

// ExtendVG ...
func (nl *NodeLVM) ExtendVG(name, physicalVolume string) (string, error) {
	args := []string{NsenterCmd, "vgextend", name, physicalVolume, "-v"}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err
}

// RemoveVG ...
func (nl *NodeLVM) RemoveVG(name string) (string, error) {
	vgs, err := nl.ListVG()
	if err != nil {
		return "", fmt.Errorf("failed to list VGs: %v", err)
	}
	var vg *model.VG
	for _, v := range vgs {
		if v.Name == name {
			vg = v
			break
		}
	}
	if vg == nil {
		return "", fmt.Errorf("could not find vg to delete")
	}
	for _, tag := range vg.Tags {
		if tag == ProtectedTagName {
			return "", errors.New("volume is protected")
		}
	}

	args := []string{NsenterCmd, "vgremove", "-v", "-f", name}
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err

}

// AddTagLV ...
func (nl *NodeLVM) AddTagLV(vg, name string, tags []string) (string, error) {
	lvs, err := nl.ListLV(fmt.Sprintf("%s/%s", vg, name))
	if err != nil {
		return "", fmt.Errorf("failed to list LVs: %v", err)
	}
	if len(lvs) != 1 {
		return "", fmt.Errorf("expected 1 LV, got %d", len(lvs))
	}

	args := make([]string, 0)
	args = append(args, NsenterCmd)
	args = append(args, "lvchange")
	for _, tag := range tags {
		args = append(args, "--addtag", tag)
	}

	args = append(args, fmt.Sprintf("%s/%s", vg, name))
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)

	return string(out), err
}

// RemoveTagLV ....
func (nl *NodeLVM) RemoveTagLV(vg, name string, tags []string) (string, error) {

	lvs, err := nl.ListLV(fmt.Sprintf("%s/%s", vg, name))
	if err != nil {
		return "", fmt.Errorf("failed to list LVs: %v", err)
	}
	if len(lvs) != 1 {
		return "", fmt.Errorf("expected 1 LV, got %d", len(lvs))
	}

	args := make([]string, 0)
	args = append(args, NsenterCmd)
	args = append(args, "lvchange")
	for _, tag := range tags {
		args = append(args, "--deltag", tag)
	}

	args = append(args, fmt.Sprintf("%s/%s", vg, name))
	cmd := strings.Join(args, " ")
	out, err := Run(cmd)
	return string(out), err
}
