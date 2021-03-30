package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func (dc *LiveDependencyCheckers) HasValidAzureToken(cmd *cobra.Command) (bool, error) {
	output, err := exec.Command("/bin/bash", "-c", "az aks list").Output()
	if (err != nil) || (strings.Contains(string(output), "refresh token has expired")) {
		cmd.Println("Azure authentication token invalid. Please execute 'az login' and run again..")
		return false, err
	}
	return true, nil
}

func (dc *LiveDependencyCheckers) IsClusterWithKubeflowCreated(cmd *cobra.Command) (bool, error) {
	output, err := exec.Command("/bin/bash", "-c", "kubectl get namespace kubeflow -o yaml").CombinedOutput()
	return strings.Contains(string(output), "name: kubeflow"), err
}

func (dc *LiveDependencyCheckers) IsStorageConfigured(cmd *cobra.Command) (bool, error) {
	output, err := exec.Command("/bin/bash", "-c", `[ "$(kubectl get sc blob -o=jsonpath='{.provisioner}')" == "blob.csi.azure.com" ]`).CombinedOutput()
	return (string(output) == ""), err
}

func (dc *LiveDependencyCheckers) CheckDependenciesInstalled(cmd *cobra.Command) error {
	kubectlPath, err := dc.IsKubectlOnPath(cmd)
	if err != nil || kubectlPath == "" {
		cmd.Printf("Could not find Kubectl on your path: %v", err.Error())
		return err
	}

	log.Tracef("Error")

	kubeconfigValue := os.Getenv("KUBECONFIG")
	if kubeconfigValue == "" {
		// From here: https://github.com/k3s-io/k3s/issues/3087
		message := INIT_ERROR_KUBECONFIG_UNSET_WARN
		cmd.Println(message)
	}

	if ok, err := dc.CanConnectToKubernetes(cmd); ok && err != nil {
		message := MISSING_KUBERNETES_ENDPOINT
		return fmt.Errorf(message)
	}

	if ok, err := dc.HasKubeflowNamespace(cmd); ok && err != nil {
		message := MISSING_KUBEFLOW_NAMESPACE
		cmd.Println(message)
	}

	return nil
}

func (dc *LiveDependencyCheckers) CreateAKSwithKubeflow(cmd *cobra.Command) error {

	_, err := exec.LookPath("az")
	if err != nil {

		cmd.Println("The Azure CLI is not installed.")
		cmd.Println("Install with https://aka.ms/getcli.")
		return err
	}

	_, err = exec.Command("/bin/bash", "-c", "az account list -otable").Output()
	if err != nil {

		cmd.Println("You are not logged in to Azure.")
		cmd.Println("Please run 'az login'")
		return err
	}

	testLogin := `
	#!/bin/bash
	set -e
	export CURRENT_LOGIN=` + "`" + `az account show -o json | jq '\''"\(.name) : \(.id)"'\''` + "`" + `
	echo "You are logged in with the following credentials: $CURRENT_LOGIN"
	echo "If this is not correct, please execute:"
	echo "az account list -o json | jq '\''.[] | \"\(.name) : \(.id)\"'\''"
	echo "az account set --subscription REPLACE_WITH_YOUR_SUBSCRIPTION_ID"
	`

	if err := utils.ExecuteInlineBashScript(cmd, testLogin, "Your account does not appear to be logged into Azure. Please execute `az login` to authorize this account."); err != nil {
		return err
	}

	log.Info("Tested login, working correctly.")

	// Instead of calling a bash script we will call the appropriate GO SDK functions or use Terraform
	theDEMOINSTALL := `
	#!/bin/bash
	set -e
	export SAME_RESOURCE_GROUP="SAME-GROUP-$RANDOM"
	export SAME_LOCATION="westus2"
	export SAME_CLUSTER_NAME="SAME-CLUSTER-$RANDOM"
	echo "export SAME_RESOURCE_GROUP=$SAME_RESOURCE_GROUP"
	echo "export SAME_LOCATION=$SAME_LOCATION"
	echo "export SAME_CLUSTER_NAME=$SAME_CLUSTER_NAME"
	echo "Creating Resource group $SAME_RESOURCE_GROUP in $SAME_LOCATION"
	az group create -n $SAME_RESOURCE_GROUP --location $SAME_LOCATION -onone
	echo "Creating AKS cluster $SAME_CLUSTER_NAME"
	az aks create --resource-group $SAME_RESOURCE_GROUP --name $SAME_CLUSTER_NAME --node-count 3 --generate-ssh-keys --node-vm-size Standard_D4s_v3 --location $SAME_LOCATION 1>/dev/null
	echo "Downloading AKS Kubeconfig credentials"
	az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP 1>/dev/null
	AKS_RESOURCE_ID=$(az aks show -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP --query id -otsv)
	echo "AKS is now installed, please execute 'same init' to install Kubeflow."
	`

	// TODO: Figure out how to check for quota violations. Example:
	// Operation failed with status: 'Bad Request'. Details: Provisioning of resource(s) for container service SAME-CLUSTER-23542 in resource group SAME-GROUP-10482 failed. Message: Operation could not be completed as it results in exceeding approved standardDSv2Family Cores quota. Additional details - Deployment Model: Resource Manager, Location: westus2, Current Limit: 200, Current Usage: 194, Additional Required: 24, (Minimum) New Limit Required: 218. Submit a request for Quota increase at https://aka.ms/ProdportalCRP/?#create/Microsoft.Support/Parameters/%7B%22subId%22:%222865c7d1-29fa-485a-8862-717377bdbf1b%22,%22pesId%22:%2206bfd9d3-516b-d5c6-5802-169c800dec89%22,%22supportTopicId%22:%22e12e3d1d-7fa0-af33-c6d0-3c50df9658a3%22%7D by specifying parameters listed in the ‘Details’ section for deployment to succeed. Please read more about quota limits at https://docs.microsoft.com/en-us/azure/azure-supportability/per-vm-quota-requests.. Details:
	cmd.Printf("About to execute: %v\n", theDEMOINSTALL)
	if err := utils.ExecuteInlineBashScript(cmd, theDEMOINSTALL, "Infrastructure set up failed. Please delete the SAME-GROUP resource group manually if it exsts."); err != nil {
		return err
	}
	return nil
}

func (dc *LiveDependencyCheckers) ConfigureStorage(cmd *cobra.Command) error {

	// Instead of calling a bash script we will call the appropriate GO SDK functions or use Terraform
	theDEMOINSTALL := `
	#!/bin/bash
	set -e
	echo "Installing Blob Storage Driver"
	curl -skSL https://raw.githubusercontent.com/kubernetes-sigs/blob-csi-driver/master/deploy/install-driver.sh | bash -s master -- 1>/dev/null
	echo "Enabling on demand storage provisioning."
	kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/blob-csi-driver/master/deploy/example/storageclass-blobfuse.yaml 1>/dev/null
	`

	// Note: To use the storage, create a PVC with spec.storageClassName: blob for dynamic provisioning

	if err := utils.ExecuteInlineBashScript(cmd, theDEMOINSTALL, "Configuring Storage failed."); err != nil {
		return err
	}
	return nil
}

func (dc *LiveDependencyCheckers) WriteCurrentContextToConfig() string {
	// TODO: This is the right way to do it, need to figure out why the struct didn't get the value set correctly
	//	currentContextScript := fmt.Sprintf("%v config current-context", dc.GetKubectlCmd())

	// HACK: Hard coded 'kubectl'
	currentContextScript := "kubectl config current-context"
	u, _ := user.Current()
	uidPlusGid := fmt.Sprintf("%v:%v", u.Username, u.Username)

	log.Tracef("About to set current context in config file: %v", currentContextScript)
	kubeConfigEnv := os.Getenv("KUBECONFIG")
	log.Tracef("KUBECONFIG value: %v", kubeConfigEnv)
	if kubeConfigEnv == "" {
		log.Info(fmt.Sprintf("KUBECONFIG is unset, setting it to %v/.kube/config.", u.HomeDir))
		err := os.Setenv("KUBECONFIG", fmt.Sprintf("%v/.kube/config", u.HomeDir))
		if err != nil {
			if utils.PrintError(fmt.Sprintf("Unable to set this user's ('%v') KUBECONFIG: ", u.Username)+"%v", err) {
				return ""
			}
		}
	}
	outputBytes, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("KUBECONFIG=%v ", kubeConfigEnv)+currentContextScript).CombinedOutput()
	if err != nil {
		outputString := string(outputBytes)
		log.Tracef("Output String: %v", outputString)
		if strings.Contains(outputString, "/etc/rancher") {
			if utils.PrintError(fmt.Sprintf(INIT_ERROR_KUBECONFIG_UNSET_FATAL, outputString)+"%v", err) {
				return ""
			}
		} else if strings.Contains(outputString, ".kube/config") || strings.Contains(outputString, "permission denied") {
			if utils.PrintError(fmt.Sprintf(INIT_ERROR_KUBECONFIG_PERMISSIONS, uidPlusGid, uidPlusGid)+"%v", err) {
				return ""
			}
		} else if strings.Contains(outputString, "current-context is not set") {
			if utils.PrintError(fmt.Sprintf(INIT_ERROR_CURRENT_CONTEXT_UNSET, currentContextScript, outputString)+": %v", err) {
				return ""
			}
		} else {
			if utils.PrintError(fmt.Sprintf(INIT_ERROR_CURRENT_CONTEXT_UNKNOWN_ERROR, currentContextScript, outputString)+": %v", err) {
				return ""
			}
		}
	}
	output := strings.TrimSpace(string(outputBytes))

	log.Tracef("Current config setting: %v\n", output)
	viper.Set("activecontext", output)
	err = viper.WriteConfig()
	if err != nil {
		if utils.PrintError(fmt.Sprintf("error writing activecontext ('%v') to config file: %v", output, viper.ConfigFileUsed()), err) {
			return ""
		}
	}

	log.Tracef("Wrote current context ('%v') as active context to file ('%v')\n", output, viper.ConfigFileUsed())
	return output

}
func (dc *LiveDependencyCheckers) CanConnectToKubernetes(cmd *cobra.Command) (bool, error) {

	k8sClient, err := utils.GetKubernetesClient()
	if err != nil {
		return false, fmt.Errorf("could not get a Kubernetes client. That's all we know: %v", err.Error())
	}
	_, err = k8sClient.GetVersion()
	if err != nil {
		return false, fmt.Errorf("could not connect to Kubernetes with the following message: %v", err.Error())
	}
	return true, nil
}

func (dc *LiveDependencyCheckers) HasKubeflowNamespace(cmd *cobra.Command) (bool, error) {
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

func (dc *LiveDependencyCheckers) IsK3sRunning(cmd *cobra.Command) (bool, error) {
	return utils.GetUtils().IsK3sRunning(cmd)
}

func (dc *LiveDependencyCheckers) IsKubectlOnPath(cmd *cobra.Command) (string, error) {
	kubectlPath, err := exec.LookPath("kubectl")
	if kubectlPath == "" || err != nil {
		if utils.PrintErrorAndReturnExit(cmd, "could not find kubectl on your path: %v %v", err) {
			return "", err
		}
	}
	return kubectlPath, nil
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
)
