package infra

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

func (dc *LiveDependencyCheckers) HasValidAzureToken(cmd *cobra.Command) error {
	output, err := exec.Command("/bin/bash", "-c", "az aks list").Output()
	if (err != nil) || (strings.Contains(string(output), "refresh token has expired")) {
		cmd.Println("Azure authentication token invalid. Please execute 'az login' and run again..")
		return err
	}
	return nil
}

func (dc *LiveDependencyCheckers) IsClusterWithKubeflowCreated(cmd *cobra.Command) error {
	return exec.Command("/bin/bash", "-c", "kubectl get namespace kubeflow").Run()
}

func (dc *LiveDependencyCheckers) IsStorageConfigured(cmd *cobra.Command) error {
	return exec.Command("/bin/bash", "-c", `[ "$(kubectl get sc blob -o=jsonpath='{.provisioner}')" == "blob.csi.azure.com" ]`).Run()
}

func (dc *LiveDependencyCheckers) CheckDependenciesInstalled(cmd *cobra.Command) error {
	_, err := exec.Command("/bin/bash", "-c", "az account list -otable").Output()
	if err != nil {

		cmd.Println("Azure CLI not installed on PATH or not logged in.")
		cmd.Println("Install with https://aka.ms/getcli and run 'az login'")
		return err
	}

	_, err = exec.Command("/bin/bash", "-c", "porter").Output()
	if err != nil {

		cmd.Println("Porter not installed or not on PATH")
		cmd.Println("Install porter at: https://porter.sh")
		return err
	}

	_, err = exec.Command("/bin/bash", "-c", "kubectl").Output()
	if err != nil {
		// No Kubectl, let's install
		cmd.Println("Running az aks install-cli to install kubectl.")
		_, err = exec.Command("/bin/bash", "-c", "az aks install-cli").Output()
		if err != nil {

			cmd.Println("Porter not installed or not on PATH")
			cmd.Println("Install porter at: https://porter.sh")
			return err
		}
	}
	return nil
}

func (dc *LiveDependencyCheckers) CreateAKSwithKubeflow(cmd *cobra.Command) error {
	credPORTER := `
	{
		"schemaVersion": "1.0.0-DRAFT+b6c701f",
		"name": "aks-kubeflow-msi",
		"created": "2021-01-28T00:15:33.5682494-08:00",
		"modified": "2021-01-28T00:15:33.5682494-08:00",
		"credentials": [
		  {
			"name": "kubeconfig",
			"source": {
			  "path": "$HOME/.kube/config"
			}
		  }
		]
	}
	`

	log.Info("Copying porter credential to local.")
	_, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' > ~/.porter/credentials/aks-kubeflow-msi.json", credPORTER)).Output()
	if err != nil {
		cmd.Println("Porter Setup: Could not create AKS credential mapping for Kubeflow Installer")
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
	echo "Installing Kubeflow into AKS Cluster via Porter"
	porter install -c aks-kubeflow-msi --reference ghcr.io/squillace/aks-kubeflow-msi:v0.1.7 1>/dev/null
	echo "Kubeflow installed."
	echo "TODO: Set up storage account."
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

	log.Tracef("About to set current context in config file: %v", currentContextScript)
	outputBytes, err := exec.Command("/bin/bash", "-c", currentContextScript).Output()
	if err != nil {
		if utils.PrintError("error getting current context", err) {
			return ""
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

type InitClusterMethods struct {
}
