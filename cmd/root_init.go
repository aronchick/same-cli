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
	"strings"

	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createCmd represents the create command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes all base services for deploying a SAME (Kubernetes, Kubeflow, etc)",
	Long:  `Initializes all base services for deploying a SAME (Kubernetes, Kubeflow, etc). Longer Description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var dc = GetDependencyCheckers()
		dc.SetCmdArgs(args)
		dc.SetCmd(cmd)

		var i = GetClusterInstallMethods()
		i.SetCmdArgs(args)

		if err := dc.CheckDependenciesInstalled(cmd); err != nil {
			return err
		}

		target := strings.ToLower(viper.GetString("target"))
		if target == "" {
			message := "No 'target' set for deployment - using 'local' as a default. To change this, please execute 'same config set target=XXXX'\n"
			target = "local"
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
			err = SetupLocal(cmd, dc, i)
		case "aks":
			message := "Executing AKS setup."
			log.Trace(message)
			cmd.Println(message)
			err = SetupAKS(cmd, dc, i)
		default:
			message := fmt.Errorf("Setup target '%v' not understood.\n", target)
			cmd.Printf(message.Error())
			log.Fatalf(message.Error())
			if os.Getenv("TEST_PASS") == "1" {
				return nil
			}
		}

		if err != nil {
			if utils.PrintError("Error while setting up Kubernetes API: %v", err) {
				return err
			}
		}

		return nil

	},
}

func GetDependencyCheckers() infra.DependencyCheckers {
	if os.Getenv("TEST_PASS") == "" {
		return &infra.LiveDependencyCheckers{}
	} else {
		return &infra.MockDependencyCheckers{}
	}
}

func GetClusterInstallMethods() infra.InstallerInterface {
	if os.Getenv("TEST_PASS") == "" {
		return &infra.LiveInstallers{}
	} else {
		return &infra.MockInstallers{}
	}
}

func SetupLocal(cmd *cobra.Command, dc infra.DependencyCheckers, i infra.InstallerInterface) (err error) {

	k8sType := "k3s"

	switch k8sType {
	case "k3s":
		_, err := i.DetectK3s("k3s")
		k3sRunning, k3sRunningErr := utils.K3sRunning(cmd)
		if err != nil {
			if utils.PrintError("k3s not installed/detected on path. Please run 'sudo same installK3s' to install: %v", err) {
				return err
			}
		} else if k3sRunningErr != nil {
			if utils.PrintError("Error checking to see if k3s is running: %v", err) {
				return err
			}
		} else if !k3sRunning {
			if utils.PrintError("Core k3s services aren't running, but the server looks correct. You may want to check back in a few minutes.", nil) {
				return err
			}
		}
		i.SetKubectlCmd("kubectl")
	default:
		if utils.PrintError("no local kubernetes type selected", nil) {
			return err
		}
	}
	log.Traceln("k3s detected, proceeding to install KFP.")
	log.Tracef("k3s path: %v", i.GetKubectlCmd())

	currentContext := dc.WriteCurrentContextToConfig()
	log.Infof("Wrote kubectl current context as: %v", currentContext)

	err = i.InstallKFP(cmd)
	if err != nil {
		if utils.PrintError("kfp failed to install: %v", err) {
			return err
		}
	}

	return nil
}

func SetupAKS(cmd *cobra.Command, dc infra.DependencyCheckers, i infra.InstallerInterface) (err error) {
	log.Trace("Testing AZ Token")
	hasToken, err := dc.HasValidAzureToken(cmd)
	if !hasToken || err != nil {
		return err
	}
	log.Trace("Token passed, testing cluster exists.")
	hasProvisionedNewResources := false
	if clusterCreated, err := dc.IsClusterWithKubeflowCreated(cmd); !clusterCreated || err != nil {
		log.Trace("Cluster does not exist, creating.")
		hasProvisionedNewResources = true
		if err := dc.CreateAKSwithKubeflow(cmd); err != nil {
			return err
		}
		log.Info("Cluster created.")
	}

	log.Trace("Cluster exists, testing to see if storage provisioned.")
	if storageConfigured, err := dc.IsStorageConfigured(cmd); !storageConfigured || err != nil {
		log.Trace("Storage not provisioned, creating.")
		hasProvisionedNewResources = true
		if err := dc.ConfigureStorage(cmd); err != nil {
			return err
		}
		log.Trace("Storage provisioned.")
	}

	if hasProvisionedNewResources {
		cmd.Println("Infrastructure Setup Complete. Ready to create programs.")
	} else {
		cmd.Println("Using existing infrastructure. Ready to create programs.")
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
