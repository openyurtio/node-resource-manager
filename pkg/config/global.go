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

package config

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
)

// GlobalConfig save global values for plugin
type GlobalConfig struct {
	//EcsClient  *ecs.Client
	//Region     string
	NodeInfo   *v1.Node
	KubeClient *kubernetes.Clientset
}

var (
	GlobalConfigVar GlobalConfig
)

const (
	// RegionIDTag is region id
	RegionIDTag = "region-id"
	// InstanceIDTag is instance id
	InstanceIDTag = "instance-id"
)

// GlobalConfigSet set Global Config
func GlobalConfigSet(nodeID, masterURL, kubeconfig string) {

	// Global Configs Set
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeID, metav1.GetOptions{})
	if err != nil {
		klog.Fatalf("Error get current node info: %s", err.Error())
	}

	// Global Config Set
	GlobalConfigVar = GlobalConfig{
		KubeClient: kubeClient,
		NodeInfo:   node,
	}
}
