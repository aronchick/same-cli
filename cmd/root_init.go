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

	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
)

var cmdArgs []string

type mockInstallers struct {
}

func (m *mockInstallers) InstallK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	return m.DetectK3s("k3s")
}

func (m *mockInstallers) StartK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	return m.DetectK3s("k3s")
}
func (m *mockInstallers) DetectK3s(s string) (string, error) {
	if utils.ContainsString(cmdArgs, "k3s-not-detected") {
		return "", fmt.Errorf("K3S NOT DETECTED")
	}

	return "VALID", nil
}

type mockDependencyCheckers struct {
	mock.Mock
	_cmd            *cobra.Command
	_kubectlCommand string
	_installers     utils.InstallerInterface
}

func (mockDC *mockDependencyCheckers) setCmd(cmd *cobra.Command) {
	mockDC._cmd = cmd
}

func (mockDC *mockDependencyCheckers) getCmd() *cobra.Command {
	return mockDC._cmd
}

func (mockDC *mockDependencyCheckers) setKubectlCmd(s string) {
	mockDC._kubectlCommand = s
}

func (mockDC *mockDependencyCheckers) getKubectlCmd() string {
	return mockDC._kubectlCommand
}

func (mockDC *mockDependencyCheckers) setInstallers(i utils.InstallerInterface) {
	mockDC._installers = i
}

func (mockDC *mockDependencyCheckers) getInstallers() utils.InstallerInterface {
	return mockDC._installers
}

func (mockDC *mockDependencyCheckers) detectDockerBin(s string) (string, error) {
	if utils.ContainsString(cmdArgs, "no-docker-path") {
		return "", fmt.Errorf("not find docker in your PATH")
	}

	return "VALID_PATH", nil
}

func (mockDC *mockDependencyCheckers) detectDockerGroup(s string) (*user.Group, error) {
	if utils.ContainsString(cmdArgs, "no-docker-group-on-system") {
		return nil, user.UnknownGroupError("NOT_FOUND")
	}

	return &user.Group{Gid: "1001", Name: "docker"}, nil
}

func (mockDC *mockDependencyCheckers) printError(s string, err error) (exit bool) {
	message := fmt.Errorf(s, err)
	mockDC.getCmd().Printf(message.Error())
	log.Fatalf(message.Error())

	return true
}

func (mockDC *mockDependencyCheckers) getUserGroups(u *user.User) (returnGroups []string, err error) {
	if utils.ContainsString(cmdArgs, "cannot-retrieve-groups") {
		return nil, fmt.Errorf("CANNOT RETRIEVE GROUPS")
	} else if utils.ContainsString(cmdArgs, "not-in-docker-group") {
		return []string{}, nil
	}

	return []string{"docker"}, nil
}

func (mockDC *mockDependencyCheckers) installKFP() (err error) {
	if utils.ContainsString(cmdArgs, "kfp-install-failed") {
		return fmt.Errorf("INSTALL KFP FAILED")
	}

	return nil
}

func (mockDC *mockDependencyCheckers) checkDepenciesInstalled(cmd *cobra.Command) error {
	return nil
}

type dependencyCheckers interface {
	detectDockerBin(string) (string, error)
	detectDockerGroup(string) (*user.Group, error)
	getUserGroups(*user.User) ([]string, error)
	printError(string, error) bool
	checkDepenciesInstalled(*cobra.Command) error
	installKFP() error
	getCmd() *cobra.Command
	setCmd(*cobra.Command)
	getKubectlCmd() string
	setKubectlCmd(string)
	getInstallers() utils.InstallerInterface
	setInstallers(utils.InstallerInterface)
}

type liveDependencyCheckers struct {
	_cmd            *cobra.Command
	_kubectlCommand string
	_installers     utils.InstallerInterface
}

func (dc *liveDependencyCheckers) setCmd(cmd *cobra.Command) {
	dc._cmd = cmd
}

func (dc *liveDependencyCheckers) getCmd() *cobra.Command {
	return dc._cmd
}

func (dc *liveDependencyCheckers) setKubectlCmd(kubectlCommand string) {
	dc._kubectlCommand = kubectlCommand
}

func (dc *liveDependencyCheckers) getKubectlCmd() string {
	return dc._kubectlCommand
}

func (dc *liveDependencyCheckers) setInstallers(i utils.InstallerInterface) {
	dc._installers = i
}

func (dc *liveDependencyCheckers) getInstallers() utils.InstallerInterface {
	return dc._installers
}

func (dc *liveDependencyCheckers) printError(s string, err error) (exit bool) {
	message := fmt.Errorf(s, err)
	dc.getCmd().Printf(message.Error())
	log.Fatalf(message.Error())

	return false
}

func (dc *liveDependencyCheckers) detectDockerBin(s string) (string, error) {
	return exec.LookPath(s)
}

func (dc *liveDependencyCheckers) detectDockerGroup(s string) (*user.Group, error) {
	return user.LookupGroup("docker")
}

func (dc *liveDependencyCheckers) getUserGroups(u *user.User) ([]string, error) {
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
		cmdArgs = args
		i.dc = &liveDependencyCheckers{}
		i.dc.setInstallers(&utils.Installers{})

		if utils.ContainsString(args, "--unittestmode") {
			i.dc = &mockDependencyCheckers{}
			i.dc.setInstallers(&mockInstallers{})
		}

		i.dc.setCmd(cmd)

		// len in go checks for both nil and 0
		if len(allSettings) == 0 {
			message := "Nil file or empty load config settings. Please run 'same config new' to initialize."
			cmd.PrintErr(message)
			log.Fatalf(message)
			return nil
		}

		if err := i.dc.checkDepenciesInstalled(cmd); err != nil {
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
			log.Fatalf("Error while setting up Kubernetes API: %v", err)
		}

		return nil

	},
}

func (i *initClusterMethods) setup_local(cmd *cobra.Command) (err error) {
	dockerPath, err := i.dc.detectDockerBin("docker")
	if err != nil || dockerPath == "" {
		if i.dc.printError("Could not find docker in your PATH: %v", err) {
			return nil
		}
	}

	dockerGroupId, err := i.dc.detectDockerGroup("docker")

	if _, ok := err.(user.UnknownGroupError); ok {
		if i.dc.printError("could not find the group 'docker' on your system. This is required to run.", err) {
			return nil
		}
	} else if err != nil {
		if i.dc.printError("unknown error while trying to retrieve list of groups on your system. Sorry that's all we know: %v", err) {
			return nil
		}
	}
	u, _ := user.Current()
	allGroups, err := i.dc.getUserGroups(u)
	if err != nil {
		if i.dc.printError("could not retrieve a list of groups for the current user: %v", err) {
			return nil
		}
	}

	if !utils.ContainsString(allGroups, dockerGroupId.Gid) && !utils.ContainsString(allGroups, dockerGroupId.Name) {
		if i.dc.printError("user not in the 'docker' group: %v", nil) {
			return nil
		}
	}

	k8sType := "k3s"

	switch k8sType {
	case "k3s":
		k3sCommand, err := i.dc.getInstallers().DetectK3s("k3s")
		if (err != nil) || (k3sCommand == "") {
			if i.dc.printError("k3s not installed/detected on path. Please run 'sudo same install_k3s' to install: %v", err) {
				return nil
			}
		}
		i.dc.setKubectlCmd("kubectl")
	default:
		if i.dc.printError("no local kubernetes type selected", nil) {
			return nil
		}
	}

	err = i.dc.installKFP()
	if err != nil {
		if i.dc.printError("kfp failed to install", err) {
			return nil
		}
	}

	return nil
}

func (i *initClusterMethods) setup_aks(cmd *cobra.Command) (err error) {
	log.Info("Testing AZ Token")
	err = hasValidAzureToken(cmd)
	if err != nil {
		return err
	}
	log.Info("Token passed, testing cluster exists.")
	hasProvisionedNewResources := false
	if !isClusterWithKubeflowCreated(cmd) {
		log.Info("Cluster does not exist, creating.")
		hasProvisionedNewResources = true
		if err := createAKSwithKubeflow(cmd); err != nil {
			return err
		}
		log.Info("Cluster created.")
	}

	log.Info("Cluster exists, testing to see if storage provisioned.")
	if !isStorageConfigured(cmd) {
		log.Info("Storage not provisioned, creating.")
		hasProvisionedNewResources = true
		if err := configureStorage(cmd); err != nil {
			return err
		}
		log.Info("Storage provisioned.")
	}

	if hasProvisionedNewResources {
		cmd.Println("Infrastructure Setup Complete. Ready to create programs.")
	} else {
		programCmd.Println("Using existing infrastructure. Ready to create programs.")
	}

	return nil
}

func hasValidAzureToken(cmd *cobra.Command) error {
	output, err := exec.Command("/bin/bash", "-c", "az aks list").Output()
	if (err != nil) || (strings.Contains(string(output), "refresh token has expired")) {
		cmd.Println("Azure authentication token invalid. Please execute 'az login' and run again..")
		return err
	}
	return nil
}

func isClusterWithKubeflowCreated(cmd *cobra.Command) bool {
	return exec.Command("/bin/bash", "-c", "kubectl get namespace kubeflow").Run() == nil
}

func isStorageConfigured(cmd *cobra.Command) bool {
	return exec.Command("/bin/bash", "-c", `[ "$(kubectl get sc blob -o=jsonpath='{.provisioner}')" == "blob.csi.azure.com" ]`).Run() == nil
}

func (dc *liveDependencyCheckers) checkDepenciesInstalled(cmd *cobra.Command) error {
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

func createAKSwithKubeflow(cmd *cobra.Command) error {
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

func configureStorage(cmd *cobra.Command) error {

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

func (dc *liveDependencyCheckers) installKFP() (err error) {

	kubectlCommand := dc.getKubectlCmd()
	cmd := dc.getCmd()
	kfpInstall := fmt.Sprintf(`
	#!/bin/bash
	set -e
	export PIPELINE_VERSION=1.4.1
	%v apply -k "github.com/kubeflow/pipelines/manifests/kustomize/cluster-scoped-resources?ref=$PIPELINE_VERSION"
	%v wait --for condition=established --timeout=60s crd/applications.app.k8s.io
	%v apply -k "github.com/kubeflow/pipelines/manifests/kustomize/env/platform-agnostic-pns?ref=$PIPELINE_VERSION"
	`, kubectlCommand, kubectlCommand, kubectlCommand)

	if err := utils.ExecuteInlineBashScript(cmd, kfpInstall, "KFP failed to install."); err != nil {
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
