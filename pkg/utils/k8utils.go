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
	"path/filepath"
	"strings"
	"time"

	gogetter "github.com/hashicorp/go-getter"
	"github.com/mitchellh/go-homedir"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

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
// NEED TO RATIONALIZE THIS WITH THE UTILS WE HAVE ELSEWHERE (probably by making this an interface as well)
func (u *UtilsLive) IsRemoteFilePath(configFilePath string) (bool, error) {
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

func (u *UtilsLive) GetObjectKindFromUri(configFile string) (string, error) {
	isRemoteFile, err := u.IsRemoteFilePath(configFile)
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
func (u *UtilsLive) UrlToRetrive(url string, fileName string) (fullUrl url.URL, err error) {
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
func ResolveLocalFilePath(filePathToTest string) (returnFilePath string, err error) {
	_, err = os.Stat(filePathToTest)
	if err != nil {
		log.Errorf("k8sutils.go: could not find file '%v': %v", filePathToTest, err)
		return "", err
	}
	u, err := netUrl.Parse(filePathToTest)
	if err != nil {
		log.Errorf("k8sutils.go: could not parse URL '%v': %v", u, err)
		return "", err
	}
	returnFilePath, _ = filepath.Abs(u.String())
	return returnFilePath, nil
}

// NewKFPConfig : Create Kubernetes API config compatible with Pipelines from KubeConfig
func NewKFPConfig() (clientcmd.ClientConfig, error) {
	// Load kubeconfig
	var kubeconfig string
	if os.Getenv("KUBECONFIG") == "" {
		if home, _ := homedir.Dir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			panic("Could not find kube config!")
		}
	} else {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	kubebytes, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	// uses kubeconfig current context
	configFromFile, err := clientcmd.NewClientConfigFromBytes(kubebytes)
	if err != nil {
		return nil, err
	}
	rawConfig, err := configFromFile.RawConfig()
	if err != nil {
		return nil, err
	}

	namespaceConfigOverride := clientcmd.ConfigOverrides{
		Context: api.Context{
			Namespace: "kubeflow",
		},
	}
	config := clientcmd.NewDefaultClientConfig(rawConfig, &namespaceConfigOverride)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type k8sClient struct {
	clientset kubernetes.Interface
}

func GetKubernetesClient(timeout time.Duration) (*k8sClient, error) {
	var err error
	client := k8sClient{}
	clientConfig, err := NewKFPConfig()
	if err != nil {
		return nil, err
	}
	restConfig, _ := clientConfig.ClientConfig()
	restConfig.Timeout = timeout
	client.clientset, err = kubernetes.NewForConfig(restConfig)

	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (k *k8sClient) GetVersion() (string, error) {
	version, err := k.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func GetKubeConfig() (string, error) {
	kubeConfigPath, err := findKubeConfig()
	if err != nil {
		return "", err
	}

	kubeConfig, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return "", err
	}

	kubeBytes, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		return "", err
	}

	return string(kubeBytes), nil
}

// findKubeConfig finds path from env:KUBECONFIG or ~/.kube/config
func findKubeConfig() (string, error) {
	env := os.Getenv("KUBECONFIG")
	if env != "" {
		return env, nil
	}
	path, err := homedir.Expand("~/.kube/config")
	if err != nil {
		return "", err
	}
	return path, nil
}
