package infra

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"

	log "github.com/sirupsen/logrus"
)

type LiveDependencyCheckers struct {
	_cmd            *cobra.Command
	_kubectlCommand string
	_cmdArgs        []string
}

func (dc *LiveDependencyCheckers) SetCmd(cmd *cobra.Command) {
	dc._cmd = cmd
}

func (dc *LiveDependencyCheckers) GetCmd() *cobra.Command {
	return dc._cmd
}

func (dc *LiveDependencyCheckers) SetCmdArgs(args []string) {
	dc._cmdArgs = args
}

func (dc *LiveDependencyCheckers) GetCmdArgs() []string {
	return dc._cmdArgs
}

func (dc *LiveDependencyCheckers) SetKubectlCmd(kubectlCommand string) {
	dc._kubectlCommand = kubectlCommand
}

func (dc *LiveDependencyCheckers) GetKubectlCmd() string {
	return dc._kubectlCommand
}

func (dc *LiveDependencyCheckers) IsKubectlOnPath() (string, error) {
	kubectlPath, err := exec.LookPath("kubectl")
	if kubectlPath == "" || err != nil {
		return "", fmt.Errorf("could not find kubectl on your path: %v", err)
	}
	return kubectlPath, nil
}

func (dc *LiveDependencyCheckers) CheckDependenciesInstalled() error {
	cmd := dc.GetCmd()
	kubectlPath, err := dc.IsKubectlOnPath()
	if err != nil || kubectlPath == "" {
		cmd.Printf("Could not find Kubectl on your path: %v", err.Error())
		return err
	}

	log.Tracef("Error")

	// kubeconfigValue := os.Getenv("KUBECONFIG")
	// if kubeconfigValue == "" {
	// 	// From here: https://github.com/k3s-io/k3s/issues/3087
	// 	message := INIT_ERROR_KUBECONFIG_UNSET_WARN
	// 	cmd.Println(message)
	// }

	if clusters, err := dc.HasClusters(); len(clusters) > 0 && err != nil {
		message := MISSING_CLUSTERS
		return fmt.Errorf(message)
	}

	if current_context, err := dc.HasContext(); current_context != "" && err != nil {
		message := MISSING_CONTEXT
		return fmt.Errorf(message)
	}

	if ok, err := dc.CanConnectToKubernetes(); ok && err != nil {
		message := MISSING_KUBERNETES_ENDPOINT
		return fmt.Errorf(message)
	}

	if ok, err := dc.HasKubeflowNamespace(); ok && err != nil {
		message := MISSING_KUBEFLOW_NAMESPACE
		cmd.Println(message)
	}

	return nil
}

func (dc *LiveDependencyCheckers) CanConnectToKubernetes() (bool, error) {
	kfpConfig, err := utils.NewKFPConfig()
	if err != nil || kfpConfig == nil {
		return false, fmt.Errorf("could not retrieve KFP Config: %v", err.Error())
	}
	restConfig, _ := kfpConfig.ClientConfig()

	if ok, err := utils.GetUtils(dc.GetCmd(), dc.GetCmdArgs()).IsEndpointReachable(restConfig.Host); !ok || err != nil {
		return false, fmt.Errorf("could not reach Kubernetes endpoint (%v): %v", restConfig.Host, err.Error())
	}

	k8sClient, err := utils.GetKubernetesClient(20 * time.Second)
	if err != nil {
		return false, fmt.Errorf("could not get a Kubernetes client. That's all we know: %v", err.Error())
	}
	_, err = k8sClient.GetVersion()
	if err != nil {
		return false, fmt.Errorf("could not connect to Kubernetes with the following message: %v", err.Error())
	}
	return true, nil
}

func (dc *LiveDependencyCheckers) HasKubeflowNamespace() (bool, error) {
	kubectlCommand := dc.GetKubectlCmd()
	scriptCmd := exec.Command(kubectlCommand, "get deployments -o json")
	scriptOutput, err := scriptCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("Failed to see if there's a kubeflow namespace. That's all we know: %v", err)
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
		return false, fmt.Errorf("Failed to unmarshall result of kubeflow namespace test: %v", err)
	}

	if len(result.Items) < 1 {
		return false, fmt.Errorf(MISSING_KUBEFLOW_NAMESPACE)
	}

	return true, nil
}

func (dc *LiveDependencyCheckers) HasContext() (currentContext string, err error) {
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

func (dc *LiveDependencyCheckers) HasClusters() (clusters []string, err error) {
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

func (dc *LiveDependencyCheckers) IsKFPReady() (running bool, err error) {

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

type InitClusterMethods struct {
}

var (
	INIT_ERROR_KUBECONFIG_UNSET_WARN string = `
Your KUBECONFIG variable is not explicitly set. This may cause issues when you run locally, such as 'error: open /etc/rancher/k3s/k3s.yaml.lock: permission denied'. While not critical, you can ensure the proper functioning of SAME by executing the following two commands.

echo "export KUBECONFIG=$HOME/.kube/config" >> $HOME/.bashrc
export KUBECONFIG=$HOME/.kube/config 
`
	INIT_ERROR_KUBECONFIG_UNSET_FATAL string = `
Unable to set your current context because your KUBECONFIG is either unset, or pointing at '/etc/rancher/k3s/k3s.yaml' (to which you don't have permissions).
Please set it (to make it easy, you can use the following command).

export KUBECONFIG=$HOME/.kube/config

Raw error: %v

Cmd error: `
	INIT_ERROR_KUBECONFIG_PERMISSIONS string = `
It appears either your $HOME/.kube, $HOME/.kube/config don't exist, it is empty or you don't have permissions to write to it. Please execute the following commands:
sudo chown %v $HOME/.kube
sudo chown %v $HOME/.kube/config

# If using k3s - 
sudo KUBECONFIG=/etc/rancher/k3s/k3s.yaml:$HOME/.kube/config kubectl config view --flatten > $HOME/.kube/config

Raw error: `
	INIT_ERROR_CURRENT_CONTEXT_UNSET string = `
Your current context is not set. This is often because it is empty. To set it, using your local k3s file, execute the following.
KUBECONFIG=/etc/rancher/k3s/k3s.yaml:$HOME/.kube/config sudo kubectl config view --flatten > $HOME/.kube/config

command: %v
output: %v

Raw error: `
	INIT_ERROR_CURRENT_CONTEXT_UNKNOWN_ERROR string = `
error getting current context - 
command: %v
output: %v

Raw error:`

	MISSING_KUBERNETES_ENDPOINT string = `
We could not connect to a Kubernetes endpoint via your kubeconfig. Please check that your Kubernetes is up and running and you can connect with the following command:

kubectl version

Which should have both a "Client Version" and a "Server Version" section.`

	MISSING_KUBEFLOW_NAMESPACE string = `
We could not find a kubeflow namespace. Unfortunately, we require that currently (and it must be named 'kubeflow').

Please re-run: 
same init

To have same automatically create one for you and install Kubeflow Pipelines.`

	MISSING_CLUSTERS string = `
We could not find any clusters in your current KUBECONFIG. Please check with:

kubectl config get-clusters`

	MISSING_CONTEXT string = `
We could not find any context in your current KUBECONFIG. Please check with:

kubectl config get-contexts
kubectl config current-context`
)
