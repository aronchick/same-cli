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
	"bufio"
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

type mockDependencyCheckers struct {
	mock.Mock
}

func (mock *mockDependencyCheckers) lookPath(s string) (string, error) {
	if utils.ContainsString(cmdArgs, "no-docker-path") {
		return "", fmt.Errorf("not find docker in your PATH")
	}

	return "VALID_PATH", nil
}

func (dc *mockDependencyCheckers) lookGroup(s string) (*user.Group, error) {
	if utils.ContainsString(cmdArgs, "not-in-docker-group") {
		return nil, fmt.Errorf("not part of the docker group")
	} else if utils.ContainsString(cmdArgs, "no-docker-group-on-system") {
		return nil, user.UnknownGroupError("NOT_FOUND")
	}

	return &user.Group{Gid: "1001", Name: "docker"}, nil
}

func (mock *mockDependencyCheckers) printError(cmd *cobra.Command, s string, err error) (exit bool) {
	message := fmt.Errorf(s, err)
	cmd.Printf(message.Error())
	log.Fatalf(message.Error())

	return true
}

type dependencyCheckers interface {
	lookPath(string) (string, error)
	lookGroup(string) (*user.Group, error)
	printError(*cobra.Command, string, error) bool
	checkDepenciesInstalled(*cobra.Command) error
}

type liveDependencyCheckers struct {
}

func (dc *liveDependencyCheckers) printError(cmd *cobra.Command, s string, err error) (exit bool) {
	message := fmt.Errorf(s, err)
	cmd.Printf(message.Error())
	log.Fatalf(message.Error())

	return false
}

func (dc *liveDependencyCheckers) lookPath(s string) (string, error) {
	return exec.LookPath(s)
}

func (dc *liveDependencyCheckers) lookGroup(s string) (*user.Group, error) {
	return user.LookupGroup("docker")
}

func (dc *mockDependencyCheckers) checkDepenciesInstalled(cmd *cobra.Command) error {
	return nil
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

		if utils.ContainsString(args, "--unittestmode") {
			i.dc = &mockDependencyCheckers{}

		}
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
			message := fmt.Errorf("Setup target '%v' not understood.", target)
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
	dockerPath, err := i.dc.lookPath("docker")
	if err != nil || dockerPath == "" {
		if i.dc.printError(cmd, "Could not find docker in your PATH: %v", err) {
			return nil
		}

	}

	dockerGroupId, err := i.dc.lookGroup("docker")

	if _, ok := err.(user.UnknownGroupError); ok {
		if i.dc.printError(cmd, "could not find the group 'docker' on your system. This is required to run.", err) {
			return nil
		}
	} else if err != nil {
		if i.dc.printError(cmd, "unknown error while trying to retrieve list of groups on your system. Sorry that's all we know: %v", err) {
			return nil
		}
	}

	currentUser, _ := user.Current()
	allGroups, err := currentUser.GroupIds()
	if err != nil {
		message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
		cmd.Printf(message.Error())
		log.Fatal(message.Error())
		if os.Getenv("TEST_PASS") == "1" {
			return nil
		}
	}

	if !utils.ContainsString(allGroups, dockerGroupId.Gid) {
		message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
		cmd.Printf(message.Error())
		log.Fatal(message)
		if os.Getenv("TEST_PASS") == "1" {
			return nil
		}
	}

	kindInstall := `
	#!/bin/bash
	set -e
	k3ai init
	`
	if err := executeInlineBashScript(cmd, kindInstall, "Kind installed."); err != nil {
		return err
	}

	return nil
}

func (i *initClusterMethods) setup_aks(cmd *cobra.Command) (err error) {
	hasProvisionedNewResources := false
	if !isClusterWithKubeflowCreated(cmd) {
		hasProvisionedNewResources = true
		if err := createAKSwithKubeflow(cmd); err != nil {
			return err
		}
	}

	if !isStorageConfigured(cmd) {
		hasProvisionedNewResources = true
		if err := configureStorage(cmd); err != nil {
			return err
		}
	}

	if hasProvisionedNewResources {
		cmd.Println("Infrastructure Setup Complete. Ready to create programs.")
	} else {
		programCmd.Println("Using existing infrastructure. Ready to create programs.")
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

	if err := executeInlineBashScript(cmd, testLogin, "Your account does not appear to be logged into Azure. Please execute `az login` to authorize this account."); err != nil {
		return err
	}

	// Instead of calling a bash script we will call the appropriate GO SDK functions or use Terraform
	theDEMOINSTALL := `
	#!/bin/bash
	set -e
	export SAME_RESOURCE_GROUP="SAME-GROUP-$RANDOM"
	export SAME_LOCATION="westus2"
	export SAME_CLUSTER_NAME="SAME-CLUSTER-$RANDOM"
	echo "Creating Resource group $SAME_RESOURCE_GROUP in $SAME_LOCATION"
	az group create -n $SAME_RESOURCE_GROUP --location $SAME_LOCATION -onone
	echo "Creating AKS cluster $SAME_CLUSTER_NAME"
	az aks create --resource-group $SAME_RESOURCE_GROUP --name $SAME_CLUSTER_NAME --node-count 3 --generate-ssh-keys --node-vm-size Standard_DS4_v2 --location $SAME_LOCATION 1>/dev/null
	echo "Downloading AKS Kubeconfig credentials"
	az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP 1>/dev/null
	AKS_RESOURCE_ID=$(az aks show -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP --query id -otsv)
	echo "Installing Kubeflow into AKS Cluster via Porter"
	porter install -c aks-kubeflow-msi --reference ghcr.io/squillace/aks-kubeflow-msi:v0.1.7 1>/dev/null
	echo "Kubeflow installed."
	echo "TODO: Set up storage account."
	`
	if err := executeInlineBashScript(cmd, theDEMOINSTALL, "Infrastructure set up failed. Please delete the SAME-GROUP resource group manually if it exsts."); err != nil {
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

	if err := executeInlineBashScript(cmd, theDEMOINSTALL, "Configuring Storage failed."); err != nil {
		return err
	}
	return nil
}

func executeInlineBashScript(cmd *cobra.Command, SCRIPT string, errorMessage string) error {
	scriptCMD := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' | bash -s --", SCRIPT))
	outPipe, err := scriptCMD.StdoutPipe()
	errPipe, _ := scriptCMD.StderrPipe()
	if err != nil {
		cmd.Println(errorMessage)
		return err
	}
	err = scriptCMD.Start()

	if err != nil {
		cmd.Println(errorMessage)
		return err
	}
	errScanner := bufio.NewScanner(errPipe)
	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		m := scanner.Text()
		cmd.Println(m)
	}
	err = scriptCMD.Wait()

	if err != nil {
		for errScanner.Scan() {
			m := errScanner.Text()
			cmd.Println(m)
		}
		cmd.Println(errorMessage)
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
