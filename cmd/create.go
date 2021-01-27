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
	"os/exec"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// for simplicity we currently rely on Porter, Azure CLI and Kubectl

		if err := checkDepenciesInstalled(); err != nil {
			return err
		}

		hasProvisionedNewResources := false
		if !isClusterWithKubeflowCreated() {
			hasProvisionedNewResources = true
			if err := createAKSwithKubeflow(); err != nil {
				return err
			}
		}

		if !isStorageConfigured() {
			hasProvisionedNewResources = true
			if err := configureStorage(); err != nil {
				return err
			}
		}

		if hasProvisionedNewResources {
			fmt.Println("Infrastructure Setup Complete. Ready to create programs.")
		} else {
			fmt.Println("Using existing infrastructure. Ready to create programs.")
		}

		return nil

	},
}

func isClusterWithKubeflowCreated() bool {
	return exec.Command("/bin/bash", "-c", "kubectl get namespace kubeflow").Run() == nil
}

func isStorageConfigured() bool {
	return exec.Command("/bin/bash", "-c", `[ "$(kubectl get sc blob -o=jsonpath='{.provisioner}')" == "blob.csi.azure.com" ]`).Run() == nil
}

func checkDepenciesInstalled() error {
	_, err := exec.Command("/bin/bash", "-c", "az account list -otable").Output()
	if err != nil {

		println("Azure CLI not installed on PATH or not logged in.")
		println("Install with https://aka.ms/getcli and run 'az login'")
		return err
	}

	_, err = exec.Command("/bin/bash", "-c", "porter").Output()
	if err != nil {

		println("Porter not installed or not on PATH")
		println("Install porter at: https://porter.sh")
		return err
	}

	_, err = exec.Command("/bin/bash", "-c", "kubectl").Output()
	if err != nil {
		// No Kubectl, let's install
		println("Running az aks install-cli to install kubectl.")
		_, err = exec.Command("/bin/bash", "-c", "az aks install-cli").Output()
		if err != nil {

			println("Porter not installed or not on PATH")
			println("Install porter at: https://porter.sh")
			return err
		}
	}
	return nil
}

func createAKSwithKubeflow() error {
	credPORTER := `
	{
		"schemaVersion": "1.0.0-DRAFT+b6c701f",
		"name": "aks-kubeflow",
		"created": "2021-01-25T22:04:51.022055-08:00",
		"modified": "2021-01-25T22:04:51.022055-08:00",
		"credentials": [
		{
			"name": "AZURE_CLIENT_ID",
			"source": {
			"env": "SP_CLIENT_ID"
			}
		},
		{
			"name": "AZURE_SP_PASSWORD",
			"source": {
			"env": "SP_PASSWORD"
			}
		},
		{
			"name": "AZURE_SUBSCRIPTION_ID",
			"source": {
			"env": "SUBSCRIPTION_ID"
			}
		},
		{
			"name": "TENANT_ID_OR_DNS",
			"source": {
			"env": "SP_TENANT"
			}
		}
		]
	}
	`

	_, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' > ~/.porter/credentials/aks-kubeflow.json", credPORTER)).Output()
	if err != nil {
		fmt.Println("Porter Setup: Could not create AKS credential mapping for Kubeflow Installer")
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
	az aks create --resource-group $SAME_RESOURCE_GROUP --name $SAME_CLUSTER_NAME --node-count 3 --generate-ssh-keys --node-vm-size Standard_DS4_v2 --location $SAME_LOCATION --only-show-errors
	echo "Downloading AKS Kubeconfig credentials"
	az aks get-credentials -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP --only-show-errors
	AKS_RESOURCE_ID=$(az aks show -n $SAME_CLUSTER_NAME -g $SAME_RESOURCE_GROUP --query id -otsv)
	SP_NAME="SAME-AKS-read-$RANDOM"
	echo "Creating Service Principal for Kubeflow deploy with ID: http://$SP_NAME"
	export SP_PASSWORD=$(az ad sp create-for-rbac -n $SP_NAME --role "Azure Kubernetes Service Cluster User Role" --scopes $AKS_RESOURCE_ID --query password -otsv)
	export SP_TENANT=$(az ad sp show --id http://$SP_NAME --query appOwnerTenantId -otsv)
	export SP_CLIENT_ID=$(az ad sp show --id http://$SP_NAME --query appId -otsv)
	export SUBSCRIPTION_ID=$(az account show --query id -otsv)
	echo "Created Servive Principal http://$SP_NAME with App ID $SP_CLIENT_ID for Subscription $SUBSCRIPTION_ID belonging to Tenant $SP_TENANT"
	echo "Installing Kubeflow into AKS Cluster via Porter"
	porter install -c aks-kubeflow --tag ghcr.io/squillace/aks-kubeflow:v0.3.1 --param AZURE_RESOURCE_GROUP=$SAME_RESOURCE_GROUP --param CLUSTER_NAME=$SAME_CLUSTER_NAME 1>/dev/null
	echo "Kubeflow installed."
	echo "TODO: Set up storage account."
	`
	if err := executeInlineBashScript(theDEMOINSTALL, "Infrastructure set up failed. Please delete the SAME-GROUP resource group manually if it exsts."); err != nil {
		return err
	}
	return nil
}

func configureStorage() error {

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

	if err := executeInlineBashScript(theDEMOINSTALL, "Configuring Storage failed."); err != nil {
		return err
	}
	return nil
}

func executeInlineBashScript(SCRIPT string, errorMessage string) error {
	scriptCMD := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' | bash -s --", SCRIPT))
	outPipe, err := scriptCMD.StdoutPipe()
	errPipe, _ := scriptCMD.StderrPipe()
	if err != nil {
		fmt.Println(errorMessage)
		return err
	}
	err = scriptCMD.Start()
	if err != nil {
		fmt.Println(errorMessage)
		return err
	}
	errScanner := bufio.NewScanner(errPipe)
	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	err = scriptCMD.Wait()

	if err != nil {
		for errScanner.Scan() {
			m := errScanner.Text()
			println(m)
		}
		fmt.Println(errorMessage)
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
