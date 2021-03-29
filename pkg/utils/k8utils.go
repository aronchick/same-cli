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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogetter "github.com/hashicorp/go-getter"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

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

func HasContext(cmd *cobra.Command) (currentContext string, err error) {
	// https://golang.org/pkg/os/exec/#example_Cmd_StdoutPipe
	scriptCmd := exec.Command("kubectl", "config", "current-context")
	log.Tracef("About to execute: %v", scriptCmd)
	scriptOutput, err := scriptCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to current context for Kubernetes. That's all we know: %v", err)
	}
	currentContextString := strings.TrimSpace(string(scriptOutput))
	if currentContextString != "" {
		return strings.TrimSpace(currentContextString), nil
	} else {
		return "", fmt.Errorf("kubectl config current-context is empty")
	}
}

func HasClusters(cmd *cobra.Command) (clusters []string, err error) {
	// strings.Split(result,`\n`)
	scriptCmd := exec.Command("kubectl", "config", "get-clusters")
	scriptOutput, err := scriptCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Failed to get current clusters. That's all we know: %v", err)
	}
	currentClusterString := string(scriptOutput)
	clusters = strings.Split(currentClusterString, "\n")
	if len(clusters) < 1 {
		return nil, fmt.Errorf("Error when getting clusters, but we don't know anything more about it.")
	} else if len(clusters) == 1 {
		return []string{}, fmt.Errorf("We were able to get clusters, but there were none in the kubeconfig.")
	}
	return clusters[1:], nil
}

func IsKFPReady(cmd *cobra.Command) (running bool, err error) {

	scriptCmd := exec.Command("/bin/bash", "-c", "kubectl get deployments --namespace=kubeflow -o json")
	scriptOutput, err := scriptCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("Failed to test if k3s is running. That's all we know: %v", err)
	}

	// Declared an empty interface
	//var result map[string]interface{}
	var result v1.DeploymentList

	//  waiting_pod_array=("k8s-app=kube-dns;kube-system"
	// "k8s-app=metrics-server;kube-system"

	// Unmarshal or Decode the JSON to the interface.
	//err = json.Unmarshal([]byte(scriptOutput), &result)
	err = json.Unmarshal(scriptOutput, &result)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshall result of kubeflow test: %v", err)
	}

	all_ready := true

	for _, deployment := range result.Items {
		all_ready = all_ready && (deployment.Status.ReadyReplicas > 0)
	}

	return all_ready, nil
}

func IsK3sHealthy(cmd *cobra.Command) (kubectlCommand string, err error) {
	_, err = GetUtils().DetectK3s()
	k3sRunning, k3sRunningErr := GetUtils().IsK3sRunning(cmd)
	if err != nil {
		if PrintError("k3s not installed/detected on path. Please run 'sudo same installK3s' to install: %v", err) {
			return "", err
		}
	} else if k3sRunningErr != nil {
		if PrintError("Error checking to see if k3s is running: %v", err) {
			return "", err
		}
	} else if !k3sRunning {
		if PrintError("Core k3s services aren't running, but the server looks correct. You may want to check back in a few minutes.", nil) {
			return "", err
		}
	}

	return "kubectl", nil
}

// NewKFPConfig : Create Kubernetes API config compatible with Pipelines from KubeConfig
func NewKFPConfig() *clientcmd.ClientConfig {
	// Load kubeconfig
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		panic("Could not find kube config!")
	}

	kubebytes, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		panic(err)
	}
	// uses kubeconfig current context
	config, err := clientcmd.NewClientConfigFromBytes(kubebytes)
	if err != nil {
		panic(err)
	}

	return &config
}

type k8sClient struct {
	clientset kubernetes.Interface
}

func GetKubernetesClient() (*k8sClient, error) {
	var err error
	client := k8sClient{}
	clientConfig := *NewKFPConfig()
	restConfig, _ := clientConfig.ClientConfig()
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
