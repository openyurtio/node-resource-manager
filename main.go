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

package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"time"

	"github.com/openyurtio/node-resource-manager/pkg/manager"
	"github.com/openyurtio/node-resource-manager/pkg/signals"
	klog "k8s.io/klog/v2"
)

const (
	// LogfilePrefix prefix of log file
	LogfilePrefix = "/var/log/openyurt/"
	// MBSIZE MB size
	MBSIZE = 1024 * 1024
)

// BRANCH is NRM Branch
var _BRANCH_ = ""

// VERSION is NRM Version
var _VERSION_ = ""

// BUILDTIME is NRM Buildtime
var _BUILDTIME_ = ""

var (
	nodeID         = flag.String("nodeid", "", "node id")
	cmName         = flag.String("cm-name", "node-resource-manager", "used configmap name")
	cmNameSpace    = flag.String("cm-namespace", "kube-system", "used configmap namespace")
	updateInterval = flag.Int("update-interval", 30, "Node Storage update internal time(s)")
	masterURL      = flag.String("master", "", "The address of the Kubernetes API server (https://hostname:port, overrides any value in kubeconfig)")
	kubeconfig     = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information")
)

// main
func main() {
	flag.Parse()

	// set log config
	setLogAttribute("unified-resource-manager")
	klog.Infof("Unified Resource Manager, BuildBranch: %s Version: %s, Build Time: %s", _BRANCH_, _VERSION_, _BUILDTIME_)

	// new signal handler
	stopCh := signals.SetupSignalHandler()

	// New Controller Manager
	manager := manager.NewManager(*nodeID, *cmName, *cmNameSpace, *updateInterval, *masterURL, *kubeconfig)
	manager.Run(stopCh)

	os.Exit(0)
}

func init() {
	flag.Set("logtostderr", "true")
}

// rotate log file by 2M bytes
// default print log to stdout and file both.
func setLogAttribute(driver string) {
	logType := os.Getenv("LOG_TYPE")
	logType = strings.ToLower(logType)
	if logType != "stdout" && logType != "host" {
		logType = "both"
	}
	if logType == "stdout" {
		return
	}

	os.MkdirAll(LogfilePrefix, os.FileMode(0755))
	logFile := LogfilePrefix + driver + ".log"
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}

	// rotate the log file if too large
	if fi, err := f.Stat(); err == nil && fi.Size() > 2*MBSIZE {
		f.Close()
		timeStr := time.Now().Format("-2006-01-02-15:04:05")
		timedLogfile := LogfilePrefix + driver + timeStr + ".log"
		os.Rename(logFile, timedLogfile)
		f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}
	}
	if logType == "both" {
		mw := io.MultiWriter(os.Stdout, f)
		klog.SetOutput(mw)
	} else {
		klog.SetOutput(f)
	}
}
