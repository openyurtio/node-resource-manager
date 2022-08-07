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
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openyurtio/node-resource-manager/pkg/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
)

const (
	// MetadataURL is metadata url
	MetadataURL = "http://100.100.100.200/latest/meta-data/"

	// NsenterCmd use to init resource
	NsenterCmd = "/usr/bin/nsenter --mount=/proc/1/ns/mnt --ipc=/proc/1/ns/ipc --net=/proc/1/ns/net --uts=/proc/1/ns/uts "

	// NodeResourceManager is resource manager
	NodeResourceManager = "node-resource-manager"
)

// ErrParse ...
var ErrParse = errors.New("Cannot parse output of blkid")

// GetMetaData get metadata from ecs meta-server
func GetMetaData(resource string) (string, error) {
	resp, err := http.Get(MetadataURL + resource)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Run run shell command
func Run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run cmd: " + cmd + ", with out: " + string(out) + ", with error: " + err.Error())
	}
	return string(out), nil
}

// NodeFilter go through all configmap to find current node config
func NodeFilter(configOperator metav1.LabelSelectorOperator, configKey, configValue string, nodeInfo *v1.Node) bool {

	isMatched := false
	switch configOperator {
	case metav1.LabelSelectorOpIn:
		for key, value := range nodeInfo.Labels {
			if key == configKey && value == configValue {
				isMatched = true
			}
		}
	case metav1.LabelSelectorOpNotIn:
		flag := false
		for key, value := range nodeInfo.Labels {
			if key == configKey && value == configValue {
				flag = true
			}
		}
		if flag == false {
			isMatched = true
		}
	case metav1.LabelSelectorOpExists:
		for key := range nodeInfo.Labels {
			if key == configKey {
				isMatched = true
			}
		}
	case metav1.LabelSelectorOpDoesNotExist:
		flag := false
		for key := range nodeInfo.Labels {
			if key == configKey {
				flag = true
			}
		}
		if flag == false {
			isMatched = true
		}
	default:
		klog.Errorf("Get unsupported operator: %s", configOperator)
	}
	return isMatched
}

// ConvertRegion2Namespace ...
func ConvertRegion2Namespace(region string) string {
	regionIndex := region[6:]
	return fmt.Sprintf("namespace%s.0", regionIndex)
}

// ConvertNamespace2LVMDevicePath ...
func ConvertNamespace2LVMDevicePath(namespace string, regions *model.PmemRegions) string {
	for _, region := range regions.Regions {
		for _, actualNamespace := range region.Namespaces {
			if actualNamespace.Dev == namespace {
				return filepath.Join("/dev", actualNamespace.BlockDev)
			}
		}
	}
	return ""
}

func checkFSType(devicePath string) (string, error) {
	// We use `file -bsL` to determine whether any filesystem type is detected.
	// If a filesystem is detected (ie., the output is not "data", we use
	// `blkid` to determine what the filesystem is. We use `blkid` as `file`
	// has inconvenient output.
	// We do *not* use `lsblk` as that requires udev to be up-to-date which
	// is often not the case when a device is erased using `dd`.
	output, err := exec.Command("file", "-bsL", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "data" {
		return "", nil
	}
	output, err = exec.Command("blkid", "-c", "/dev/null", "-o", "export", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "=")
		if len(fields) != 2 {
			return "", ErrParse
		}
		if fields[0] == "TYPE" {
			return fields[1], nil
		}
	}
	return "", ErrParse
}

// IsPart if smallList is part of or equal to largeList, return true;
func IsPart(largeList, smallList []string) bool {
	isPartFlag := true
	for _, smalltmp := range smallList {
		flag := false
		for _, largetmp := range largeList {
			if smalltmp == largetmp {
				flag = true
			}
		}
		if flag == false {
			isPartFlag = false
		}
	}
	return isPartFlag
}

func NewEventRecorder() record.EventRecorder {
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("NewEventRecorder:: Failed to create cluster config: %v", err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	source := v1.EventSource{Component: NodeResourceManager}
	if broadcaster != nil {
		sink := &v1core.EventSinkImpl{
			Interface: v1core.New(clientset.CoreV1().RESTClient()).Events(""),
		}
		broadcaster.StartRecordingToSink(sink)
	}
	return broadcaster.NewRecorder(scheme.Scheme, source)
}
