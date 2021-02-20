/*
Copyright (c) 2016-2017 Bitnami
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
	"bytes"
	"fmt"
	"os"
	"strings"

	gogetter "github.com/hashicorp/go-getter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"io/ioutil"

	log "github.com/sirupsen/logrus"

	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"net/url"
	netUrl "net/url"
	"path"
)

const (
	YamlSeparator = "(?m)^---[ \t]*$"
	CertDir       = "/opt/ca"
	// controlPlaneLabel          = "control-plane"
	// katibMetricsCollectorLabel = "katib-metricscollector-injection"
	KfDefAnnotation   = "kfctl.kubeflow.io"
	ForceDelete       = "force-delete"
	SetAnnotation     = "set-kubeflow-annotation"
	KfDefInstance     = "kfdef-instance"
	InstallByOperator = "install-by-operator"
)

// Checks if the path configFile is remote (e.g. http://github...)
func IsRemoteFilePath(configFilePath string) (bool, error) {
	if configFilePath == "" {
		return false, fmt.Errorf("config file must be a URI or a path")
	}
	url, err := netUrl.Parse(configFilePath)
	if err != nil {
		return false, fmt.Errorf("unable to parse the configfile URL")
	}
	if url.Scheme == "" {
		message := fmt.Errorf("No scheme specified in the URL, so assuming local - '%v'", url)
		log.Infof(message.Error())
		return false, nil
	}

	return true, nil
}

func GetObjectKindFromUri(configFile string) (string, error) {
	isRemoteFile, err := IsRemoteFilePath(configFile)
	if err != nil {
		return "", err
	}

	// We will read from appFile.
	appFile := configFile
	if isRemoteFile {
		// Download it to a tmp file, and set appFile.
		appDir, err := ioutil.TempDir("", "")
		if err != nil {
			return "", fmt.Errorf("Create a temporary directory to copy the file to.")
		}
		// Open config file
		appFile = path.Join(appDir, "tmp.yaml")

		log.Infof("Downloading %v to %v", configFile, appFile)
		err = gogetter.GetFile(appFile, configFile)
		if err != nil {
			return "", fmt.Errorf("could not fetch specified config %s: %v", configFile, err)
		}
	}

	// Read contents
	configFileBytes, err := ioutil.ReadFile(appFile)
	if err != nil {
		return "", fmt.Errorf("could not read from config file %s: %v", configFile, err)
	}

	BUFSIZE := 1024
	buf := bytes.NewBufferString(string(configFileBytes))

	job := &unstructured.Unstructured{}
	err = k8syaml.NewYAMLOrJSONDecoder(buf, BUFSIZE).Decode(job)
	if err != nil {
		return "", fmt.Errorf("could not decode specified config %s: %v", configFile, err)
	}

	return job.GetKind(), nil
}

func JoinURL(basePath string, paths ...string) (*url.URL, error) {

	u, err := url.Parse(basePath)

	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}

	p2 := append([]string{u.Path}, paths...)

	result := path.Join(p2...)

	u.Path = result

	return u, nil
}

// FileToRetrive checks to see if the URL ends with same.yaml (or whatever is provided) and returns a well structured URL
func UrlToRetrive(url string, fileName string) (fullUrl url.URL, err error) {
	finalUrl, err := netUrl.Parse(url)
	if err != nil {
		message := fmt.Errorf("could not parse final url: %v", err)
		log.Error(message)
		return netUrl.URL{}, message
	}
	if strings.HasSuffix(finalUrl.String(), fileName) {
		return *finalUrl, nil
	}
	// https://raw.githubusercontent.com/SAME-Project/Sample-SAME-Data-Science/same.yaml
	// https://raw.githubusercontent.com/SAME-Project/Sample-SAME-Data-Science/same.yaml

	finalUrl, err = JoinURL(finalUrl.String(), fileName)
	if err != nil {
		message := fmt.Errorf("unable to join url (%v) and fileName (%v): %v", url, fileName, err)
		log.Error(message)
		return netUrl.URL{}, message
	}
	return *finalUrl, nil
}

// ResolveLocalFilePath takes local file path string, tests for its existence and resolves file:// to local path
func ResolveLocalFilePath(filePathToTest string) (filePath string, err error) {
	fileInfo, err := os.Stat(filePathToTest)
	if err != nil {
		log.Errorf("k8sutils.go: could not find file '%v': %v", filePathToTest, err)
		return "", err
	}
	u, err := url.ParseRequestURI(fileInfo.Name())
	if err != nil {
		log.Errorf("k8sutils.go: could not parse URL '%v': %v", u, err)
		return "", err
	}
	return u.String(), nil
}
