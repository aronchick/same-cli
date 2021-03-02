package cmd

/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type dependencyCheckers interface {
	PrintError(string, error) bool
	CheckDependenciesInstalled(*cobra.Command) error
	HasValidAzureToken(*cobra.Command) error
	IsClusterWithKubeflowCreated(*cobra.Command) error
	CreateAKSwithKubeflow(*cobra.Command) error
	IsStorageConfigured(*cobra.Command) error
	ConfigureStorage(*cobra.Command) error
	InstallKFP() error
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetKubectlCmd() string
	SetKubectlCmd(string)
	GetInstallers() utils.InstallerInterface
	SetInstallers(utils.InstallerInterface)
	GetCmdArgs() []string
	SetCmdArgs([]string)
}

type liveDependencyCheckers struct {
	_cmd            *cobra.Command
	_kubectlCommand string
	_installers     utils.InstallerInterface
	_cmdArgs        []string
}

func (dc *liveDependencyCheckers) SetCmd(cmd *cobra.Command) {
	dc._cmd = cmd
}

func (dc *liveDependencyCheckers) GetCmd() *cobra.Command {
	return dc._cmd
}

func (dc *liveDependencyCheckers) SetCmdArgs(args []string) {
	dc._cmdArgs = args
}

func (dc *liveDependencyCheckers) GetCmdArgs() []string {
	return dc._cmdArgs
}

func (dc *liveDependencyCheckers) SetKubectlCmd(kubectlCommand string) {
	dc._kubectlCommand = kubectlCommand
}

func (dc *liveDependencyCheckers) GetKubectlCmd() string {
	return dc._kubectlCommand
}

func (dc *liveDependencyCheckers) SetInstallers(i utils.InstallerInterface) {
	dc._installers = i
}

func (dc *liveDependencyCheckers) GetInstallers() utils.InstallerInterface {
	return dc._installers
}

func (dc *liveDependencyCheckers) PrintError(s string, err error) (exit bool) {
	return utils.PrintError(s, err)
}

func (dc *liveDependencyCheckers) DetectDockerBin(s string) (string, error) {
	return exec.LookPath(s)
}

func (dc *liveDependencyCheckers) DetectDockerGroup(s string) (*user.Group, error) {
	return user.LookupGroup("docker")
}

func (dc *liveDependencyCheckers) GetUserGroups(u *user.User) ([]string, error) {
	return u.GroupIds()
}

type initClusterMethods struct {
	dc dependencyCheckers
}

// createCmd represents the create command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes all base services for deploying a SAME (Kubernetes, Kubeflow, etc)",
	Long:  `Initializes all base services for deploying a SAME (Kubernetes, Kubeflow, etc). Longer Description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// for simplicity we currently rely on Porter, Azure CLI and Kubectl
		allSettings := viper.AllSettings()

		var i = &initClusterMethods{}
		i.dc = &liveDependencyCheckers{}
		i.dc.SetInstallers(&utils.Installers{})

		if os.Getenv("TEST_PASS") != "" {
			i.dc = &mocks.MockDependencyCheckers{}
			i.dc.SetInstallers(&mocks.MockInstallers{})
		}

		i.dc.SetCmdArgs(args)

		i.dc.SetCmd(cmd)

		// len in go checks for both nil and 0
		if len(allSettings) == 0 {
			message := "Nil file or empty load config settings. Please run 'same config new' to initialize."
			cmd.PrintErr(message)
			log.Fatalf(message)
			return nil
		}

		if err := i.dc.CheckDependenciesInstalled(cmd); err != nil {
			return err
		}

		target := strings.ToLower(viper.GetString("target"))
		if target == "" {
			message := "No 'target' set for deployment - using 'local' as a default. To change this, please execute 'same config set target=XXXX'"
			cmd.Print(message)
			if os.Getenv("TEST_PASS") == "1" {
				return nil
			}
		}

		switch target {
		case "local":
			message := "Executing local setup."
			log.Trace(message)
			cmd.Println(message)
			err = i.setup_local(cmd)
		case "aks":
			message := "Executing AKS setup."
			log.Trace(message)
			cmd.Println(message)
			err = i.setup_aks(cmd)
		default:
			message := fmt.Errorf("Setup target '%v' not understood.\n", target)
			cmd.Printf(message.Error())
			log.Fatalf(message.Error())
			if os.Getenv("TEST_PASS") == "1" {
				return nil
			}
		}

		if err != nil {
			if i.dc.PrintError("Error while setting up Kubernetes API: %v", err) {
				return err
			}
		}

		return nil

	},
}

func (i *initClusterMethods) setup_local(cmd *cobra.Command) (err error) {

	k8sType := "k3s"

	switch k8sType {
	case "k3s":
		k3sCommand, err := i.dc.GetInstallers().DetectK3s("k3s")
		if (err != nil) || (k3sCommand == "") {
			if i.dc.PrintError("k3s not installed/detected on path. Please run 'sudo same install_k3s' to install: %v", err) {
				return err
			}
		}
		i.dc.SetKubectlCmd("kubectl")
	default:
		if i.dc.PrintError("no local kubernetes type selected", nil) {
			return err
		}
	}
	log.Traceln("k3s detected, proceeding to install KFP.")
	log.Tracef("k3s path: %v", i.dc.GetKubectlCmd())

	err = i.dc.InstallKFP()
	if err != nil {
		if i.dc.PrintError("kfp failed to install: ", err) {
			return err
		}
	}

	return nil
}

func (i *initClusterMethods) setup_aks(cmd *cobra.Command) (err error) {
	log.Trace("Testing AZ Token")
	err = i.dc.HasValidAzureToken(cmd)
	if err != nil {
		return err
	}
	log.Trace("Token passed, testing cluster exists.")
	hasProvisionedNewResources := false
	if i.dc.IsClusterWithKubeflowCreated(cmd) != nil {
		log.Trace("Cluster does not exist, creating.")
		hasProvisionedNewResources = true
		if err := i.dc.CreateAKSwithKubeflow(cmd); err != nil {
			return err
		}
		log.Info("Cluster created.")
	}

	log.Trace("Cluster exists, testing to see if storage provisioned.")
	if i.dc.IsStorageConfigured(cmd) != nil {
		log.Trace("Storage not provisioned, creating.")
		hasProvisionedNewResources = true
		if err := i.dc.ConfigureStorage(cmd); err != nil {
			return err
		}
		log.Trace("Storage provisioned.")
	}

	if hasProvisionedNewResources {
		cmd.Println("Infrastructure Setup Complete. Ready to create programs.")
	} else {
		programCmd.Println("Using existing infrastructure. Ready to create programs.")
	}

	return nil
}

func (dc *liveDependencyCheckers) HasValidAzureToken(cmd *cobra.Command) error {
	output, err := exec.Command("/bin/bash", "-c", "az aks list").Output()
	if (err != nil) || (strings.Contains(string(output), "refresh token has expired")) {
		cmd.Println("Azure authentication token invalid. Please execute 'az login' and run again..")
		return err
	}
	return nil
}

func (dc *liveDependencyCheckers) IsClusterWithKubeflowCreated(cmd *cobra.Command) error {
	return exec.Command("/bin/bash", "-c", "kubectl get namespace kubeflow").Run()
}

func (dc *liveDependencyCheckers) IsStorageConfigured(cmd *cobra.Command) error {
	return exec.Command("/bin/bash", "-c", `[ "$(kubectl get sc blob -o=jsonpath='{.provisioner}')" == "blob.csi.azure.com" ]`).Run()
}

func (dc *liveDependencyCheckers) CheckDependenciesInstalled(cmd *cobra.Command) error {
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

func (dc *liveDependencyCheckers) CreateAKSwithKubeflow(cmd *cobra.Command) error {
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

func (dc *liveDependencyCheckers) ConfigureStorage(cmd *cobra.Command) error {

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

func (dc *liveDependencyCheckers) InstallKFP() (err error) {

	log.Tracef("Inside InstallKFP()")
	kubectlCommand := dc.GetKubectlCmd()
	log.Tracef("kubectlCommand: %v\n", kubectlCommand)
	cmd := dc.GetCmd()
	kfpInstall := fmt.Sprintf(`
	#!/bin/bash
	set -e
	export PIPELINE_VERSION=1.4.1
	export KUBECTL_COMMAND=%v
	$KUBECTL_COMMAND create namespace kubeflow || true
	$KUBECTL_COMMAND config set-context --current --namespace=kubeflow
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/cluster-scoped-resources?ref=$PIPELINE_VERSION"
	$KUBECTL_COMMAND wait --for condition=established --timeout=60s crd/applications.app.k8s.io
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/env/platform-agnostic-pns?ref=$PIPELINE_VERSION"
	`, kubectlCommand)

	log.Tracef("About to execute: %v\n", kfpInstall)
	if err := utils.ExecuteInlineBashScript(cmd, kfpInstall, "KFP failed to install."); err != nil {
		log.Tracef("Error executing: %v\n", err.Error())
		return err
	}

	return nil
}

func init() {
	RootCmd.AddCommand(initCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
