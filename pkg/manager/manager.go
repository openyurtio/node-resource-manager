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

package manager

import (
	"context"
	"time"

	"github.com/openyurtio/node-resource-manager/pkg/config"
	"github.com/openyurtio/node-resource-manager/pkg/manager/memory"
	"github.com/openyurtio/node-resource-manager/pkg/manager/quotapath"
	"github.com/openyurtio/node-resource-manager/pkg/manager/volumegroup"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
)

// Manager interface define local resource manager's action
type Manager interface {
	AnalyseConfigMap() error
	ApplyResourceDiff() error
}

// UnifiedResourceManager is global resource manager struct
type UnifiedResourceManager struct {
	KubeClientSet  *kubernetes.Clientset
	UpdateInterval int
	NodeID         string
}

// NewDriver create a cpfs driver object
func NewManager(nodeID, cmName, cmNameSpace string, updateInterval int, masterURL, kubeconfig string) *UnifiedResourceManager {
	manager := &UnifiedResourceManager{
		NodeID:         nodeID,
		UpdateInterval: updateInterval,
	}
	// Config GlobalVar
	config.GlobalConfigSet(nodeID, masterURL, kubeconfig)

	return manager
}

// Run Start the UnifiedResourceManager
// First will check UnifiedResource exist or not;
// Update UnifiedResource every internal seconds, the total/free capacity, storage status;
// Maintain Unified Resource, Like: local disk in alibaba cloud cluster.
func (urm *UnifiedResourceManager) Run(stopCh <-chan struct{}) {
	ctx := context.Background()

	// Create UnifiedResource CR if not exist
	urm.CreateUnifiedResourceCRD(ctx)

	// Maintain VolumeGroup if set in configMap
	go urm.BuildUnifiedResource()

	// UpdateUnifiedStorage
	// go wait.Until(urm.RecordUnifiedResources, time.Duration(urm.UpdateInterval)*time.Second, stopCh)

	klog.V(3).Infof("Starting to update node storage on %s...", urm.NodeID)
	<-stopCh
	klog.V(3).Infof("Stop to update node storage...")
}

// CreateUnifiedResourceCRD ...
func (urm *UnifiedResourceManager) CreateUnifiedResourceCRD(ctx context.Context) error {
	return nil
}

// BuildUnifiedResource ...
func (urm *UnifiedResourceManager) BuildUnifiedResource() {
	klog.Infof("BuildUnifiedResource:: Starting to maintain unified storage...")
	rms := []Manager{volumegroup.NewResourceManager(), quotapath.NewResourceManager(), memory.NewResourceManager()}

	for {
		for _, rm := range rms {
			err := BuildResource(rm)
			if err != nil {
				continue
			}
		}
		time.Sleep(time.Duration(20) * time.Second)
	}
}

// BuildResource ...
func BuildResource(m Manager) error {

	// Get Desired VolumeGroup from ConfigMap
	err := m.AnalyseConfigMap()
	if err != nil {
		return err
	}
	return m.ApplyResourceDiff()

}

// Update Unified Storage CRD every internal seconds
func (urm *UnifiedResourceManager) RecordUnifiedResources() {
	// Get Unified Storage Object
}
