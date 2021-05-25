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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/openyurtio/node-resource-manager/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	// MetadataURL is metadata server url
	MetadataURL = "http://100.100.100.200/latest/meta-data/"
	// InstanceID is the instance id tag
	InstanceID = "instance-id"
	// RegionIDTag is the region id tag
	RegionIDTag = "region-id"
	// NsenterCmd is the nsenter command
	NsenterCmd = "/usr/bin/nsenter --mount=/proc/1/ns/mnt "
)

const (
	// ConfigPath the secret mount file
	ConfigPath = "/var/addon/token-config"
)

// AKInfo access key info
type AKInfo struct {
	// AccessKeyId access key id
	AccessKeyID string `json:"access.key.id"`
	// AccessKeySecret access key secret
	AccessKeySecret string `json:"access.key.secret"`
	// SecurityToken security token
	SecurityToken string `json:"security.token"`
	// Expiration expiration duration
	Expiration string `json:"expiration"`
	// Keyring key ring
	Keyring string `json:"keyring"`
}

// RoleAuth define STS Token Response
type RoleAuth struct {
	AccessKeyID     string
	AccessKeySecret string
	Expiration      time.Time
	SecurityToken   string
	LastUpdated     time.Time
	Code            string
}

func ListDevice(vgName string) []string {
	cmd := fmt.Sprintf("%s pvscan | grep \"VG %s\" | awk '{print $2}'", NsenterCmd, vgName)
	out, err := utils.Run(cmd)
	if err != nil {
		devs := make([]string, 0)
		return devs
	}

	outLines := strings.Split(out, "\n")
	deviceList := []string{}
	for _, line := range outLines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "WARNING") {
			deviceList = append(deviceList, line)
		}
	}
	return deviceList
}

func diffDevice(devList1, devList2 []string) bool {
	if len(devList1) != len(devList2) {
		return true
	}
	for _, dev1 := range devList1 {
		isSearched := false
		for _, dev2 := range devList2 {
			if dev1 == dev2 {
				isSearched = true
				break
			}
		}
		if !isSearched {
			return true
		}
	}

	return false
}

// NewEcsClient create a ecsClient object
func NewEcsClient(accessKeyID, accessKeySecret, accessToken string) (ecsClient *ecs.Client) {
	var err error
	if accessToken == "" {
		ecsClient, err = ecs.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessKeySecret)
		if err != nil {
			return nil
		}
	} else {
		ecsClient, err = ecs.NewClientWithStsToken("cn-hangzhou", accessKeyID, accessKeySecret, accessToken)
		if err != nil {
			return nil
		}
	}
	return
}

// GetMetaData get host regionid, zoneid
func GetMetaData(resource string) string {
	resp, err := http.Get(MetadataURL + resource)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// GetDefaultAK read default ak from local file or from STS
func GetDefaultAK() (string, string, string) {
	accessKeyID, accessSecret := GetLocalAK()

	accessToken := ""
	if accessKeyID == "" || accessSecret == "" {
		accessKeyID, accessSecret, accessToken = GetManagedToken()
		if accessKeyID != "" {
			log.Infof("Get AK: use Managed AK")
		}
	} else {
		log.Infof("Get AK: use Local AK")
	}

	if accessKeyID == "" || accessSecret == "" {
		accessKeyID, accessSecret, accessToken = GetSTSAK()
	}

	return accessKeyID, accessSecret, accessToken
}

// GetLocalAK read ossfs ak from local or from secret file
func GetLocalAK() (string, string) {
	accessKeyID, accessSecret := "", ""
	accessKeyID = os.Getenv("ACCESS_KEY_ID")
	accessSecret = os.Getenv("ACCESS_KEY_SECRET")

	return strings.TrimSpace(accessKeyID), strings.TrimSpace(accessSecret)
}

// GetSTSAK get STS AK and token from ecs meta server
func GetSTSAK() (string, string, string) {
	roleAuth := RoleAuth{}
	subpath := "ram/security-credentials/"
	roleName := GetMetaData(subpath)
	fullPath := filepath.Join(subpath, roleName)
	roleInfo := GetMetaData(fullPath)

	err := json.Unmarshal([]byte(roleInfo), &roleAuth)
	if err != nil {
		log.Errorf("GetSTSToken: unmarshal roleInfo: %s, with error: %s", roleInfo, err.Error())
		return "", "", ""
	}
	return roleAuth.AccessKeyID, roleAuth.AccessKeySecret, roleAuth.SecurityToken
}

// GetManagedToken get ak from csi secret
func GetManagedToken() (string, string, string) {
	var akInfo AKInfo
	AccessKeyID, AccessKeySecret, SecurityToken := "", "", ""
	if _, err := os.Stat(ConfigPath); err == nil {
		encodeTokenCfg, err := ioutil.ReadFile(ConfigPath)
		if err != nil {
			log.Errorf("failed to read token config, err: %v", err)
			return "", "", ""
		}
		err = json.Unmarshal(encodeTokenCfg, &akInfo)
		if err != nil {
			log.Errorf("error unmarshal token config: %v", err)
			return "", "", ""
		}
		keyring := akInfo.Keyring
		ak, err := Decrypt(akInfo.AccessKeyID, []byte(keyring))
		if err != nil {
			log.Errorf("failed to decode ak, err: %v", err)
			return "", "", ""
		}

		sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
		if err != nil {
			log.Errorf("failed to decode sk, err: %v", err)
			return "", "", ""
		}

		token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
		if err != nil {
			log.Errorf("failed to decode token, err: %v", err)
			return "", "", ""
		}
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, akInfo.Expiration)
		if err != nil {
			log.Errorf("Parse expiration error: %s", err.Error())
		}
		if t.Before(time.Now()) {
			log.Errorf("invalid token which is expired, expiration as: %s", akInfo.Expiration)
		}
		AccessKeyID = string(ak)
		AccessKeySecret = string(sk)
		SecurityToken = string(token)
	}
	return AccessKeyID, AccessKeySecret, SecurityToken
}

// PKCS5UnPadding get pkc
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// Decrypt secret Decrypt
func Decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Errorf("failed to decode base64 string, err: %v", err)
		return nil, err
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		log.Errorf("failed to new cipher, err: %v", err)
		return nil, err
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = PKCS5UnPadding(origData)
	return origData, nil
}
