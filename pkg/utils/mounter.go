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
	"io"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	utilexec "k8s.io/utils/exec"
	k8smount "k8s.io/utils/mount"

	CusErr "github.com/openyurtio/node-resource-manager/pkg/err"
)

const (
	// fsckErrorsCorrected tag
	fsckErrorsCorrected = 1
	// fsckErrorsUncorrected tag
	fsckErrorsUncorrected = 4
)

// Mounter is responsible for formatting and mounting volumes
type Mounter interface {
	k8smount.Interface
	utilexec.Interface
	// If the folder doesn't exist, it will call 'mkdir -p'
	EnsureFolder(string) error
	// FormatAndMount ...
	FormatAndMount(string, string, string, []string, string) error

	// IsMounted checks whether the target path is a correct mount (i.e:
	// propagated). It returns true if it's mounted. An error is returned in
	// case of system errors or if it's mounted incorrectly.
	IsMounted(target string) (bool, error)

	SafePathRemove(target string) error

	FileExists(file string) bool
}

// TODO(arslan): this is Linux only for now. Refactor this into a package with
// architecture specific code in the future, such as mounter_darwin.go,
// mounter_linux.go, etc..
type NodeMounter struct {
	k8smount.SafeFormatAndMount
	utilexec.Interface
}

// NewMounter returns a new mounter instance
func NewMounter() Mounter {
	return &NodeMounter{
		k8smount.SafeFormatAndMount{
			Interface: k8smount.New(""),
			Exec:      utilexec.New(),
		},
		utilexec.New(),
	}
}

// FileExists ...
func (m *NodeMounter) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// EnsureFolder ...
func (m *NodeMounter) EnsureFolder(target string) error {
	mkdirCmd := "mkdir"
	_, err := m.LookPath(mkdirCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return fmt.Errorf("%q executable not found in $PATH", mkdirCmd)
		}
		return err
	}

	mkdirCmd = NsenterCmd + mkdirCmd
	mkdirCmd += fmt.Sprintf(" -p %s", target)
	log.Infof("mkdir for folder, the command is %s", mkdirCmd)
	output, err := Run(mkdirCmd)
	if err != nil {
		return fmt.Errorf("EnsureFolder:: mkdir for folder output: %s error: %v", output, err)
	}
	return nil
}

// FormatAndMount ...
func (m *NodeMounter) FormatAndMount(source, target, fstype string, mkfsOptions []string, mountOptions string) error {
	// diskMounter.Interface = m.K8smounter
	readOnly := false

	if !readOnly {
		// Run fsck on the disk to fix repairable issues, only do this for volumes requested as rw.
		args := []string{"-a", source}

		out, err := m.Exec.Command("fsck", args...).CombinedOutput()
		if err != nil {
			ee, isExitError := err.(utilexec.ExitError)
			switch {
			case err == utilexec.ErrExecutableNotFound:
				log.Warningf("'fsck' not found on system; continuing mount without running 'fsck'.")
			case isExitError && ee.ExitStatus() == fsckErrorsCorrected:
				log.Infof("Device %s has errors which were corrected by fsck.", source)
			case isExitError && ee.ExitStatus() == fsckErrorsUncorrected:
				return fmt.Errorf("'fsck' found errors on device %s but could not correct them: %s", source, string(out))
			case isExitError && ee.ExitStatus() > fsckErrorsUncorrected:
			}
		}
	}

	// Try to mount the disk
	cmd := fmt.Sprintf("%smount -o %s %s %s", NsenterCmd, mountOptions, source, target)
	log.Infof("FormatAndMount:: cmd: %s", cmd)
	output, mountErr := Run(cmd)
	if mountErr != nil {
		// Mount failed. This indicates either that the disk is unformatted or
		// it contains an unexpected filesystem.
		existingFormat, err := m.GetDiskFormat(source)

		if err != nil {
			return err
		}
		if existingFormat == "" {
			if readOnly {
				// Don't attempt to format if mounting as readonly, return an error to reflect this.
				return errors.New("failed to mount unformatted volume as read only")
			}

			// Disk is unformatted so format it.
			args := []string{source}
			// Use 'ext4' as the default
			if len(fstype) == 0 {
				fstype = "ext4"
			}

			if fstype == "ext4" || fstype == "ext3" {
				args = []string{
					"-F",  // Force flag
					"-m0", // Zero blocks reserved for super-user
					source,
				}
				// add mkfs options
				if len(mkfsOptions) != 0 {
					args = []string{}
					for _, opts := range mkfsOptions {
						args = append(args, opts)
					}
					args = append(args, source)
				}
			}
			log.Infof("Disk %q appears to be unformatted, attempting to format as type: %q with options: %v", source, fstype, args)

			mkfsCmd := fmt.Sprintf("%s mkfs.%s %s", NsenterCmd, fstype, strings.Join(args, " "))
			log.Infof("FormatAndMount:: mkfscmd: %s", mkfsCmd)
			_, err = Run(mkfsCmd)
			if err == nil {
				// the disk has been formatted successfully try to mount it again.
				output, mountErr := Run(cmd)
				log.Infof("FormatAndMount:: cmd output %s", output)
				return mountErr
			}
			log.Errorf("format of disk %q failed: type:(%q) target:(%q) options:(%q) output: (%s) error:(%v)", source, fstype, target, mkfsOptions, output, err)
			return err
		}
		// Disk is already formatted and failed to mount
		if len(fstype) == 0 || fstype == existingFormat {
			// This is mount error
			return mountErr
		}
		// Block device is formatted with unexpected filesystem, let the user know
		return &CusErr.ExistsFormatErr{
			FsType:         fstype,
			ExistingFormat: existingFormat,
			MountErr:       mountErr,
		}
	}

	return mountErr
}

// IsMounted ...
func (m *NodeMounter) IsMounted(target string) (bool, error) {
	if target == "" {
		return false, errors.New("target is not specified for checking the mount")
	}
	findmntCmd := "grep"
	findmntArgs := []string{target, "/proc/mounts"}
	out, err := exec.Command(findmntCmd, findmntArgs...).CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		if outStr == "" {
			return false, nil
		}
		return false, fmt.Errorf("checking mounted failed: %v cmd: %q output: %q",
			err, findmntCmd, outStr)
	}
	if strings.Contains(outStr, target) {
		return true, nil
	}
	return false, nil
}

// SafePathRemove ...
func (m *NodeMounter) SafePathRemove(targetPath string) error {
	fo, err := os.Lstat(targetPath)
	if err != nil {
		return err
	}
	isMounted, err := m.IsMounted(targetPath)
	if err != nil {
		return err
	}
	if isMounted {
		return errors.New("Path is mounted, not remove: " + targetPath)
	}
	if fo.IsDir() {
		empty, err := IsDirEmpty(targetPath)
		if err != nil {
			return errors.New("Check path empty error: " + targetPath + err.Error())
		}
		if !empty {
			return errors.New("Cannot remove Path not empty: " + targetPath)
		}
	}
	err = os.Remove(targetPath)
	if err != nil {
		return err
	}
	return nil
}

// IsDirEmpty return status of dir empty or not
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)
	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
