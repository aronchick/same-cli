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
	"time"

	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/pkg/errors"

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
		var dc = infra.GetDependencyCheckers(cmd, args)

		var i = GetClusterInstallMethods()
		i.SetCmd(cmd)
		i.SetCmdArgs(args)

		target := strings.ToLower(viper.GetString("target"))
		if target == "" {
			message := "No target set - using current-context from kubeconfig.\n"
			target = "local"
			cmd.Print(message)
			if os.Getenv("TEST_PASS") == "1" {
				return nil
			}
		}

		switch target {
		case "local":
			message := "Executing local setup."
			cmd.Println(message)
			err = SetupLocal(cmd, dc, i)
		case "aks":
			message := "Executing AKS setup."
			log.Trace(message)
			cmd.Println(message)
			err = SetupAKS(cmd, dc, i)
		default:
			message := fmt.Errorf("setup target '%v' not understood", target)
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

		cmd.Println("Your installation is complete and running!")
		shouldWait, _ := cmd.PersistentFlags().GetBool("wait")
		if shouldWait {
			cmd.Println("Waiting for SAME to become ready... often takes > 5 minutes. \n You can cancel at any time and it will not affect setup. Use 'same init --ready' to check manually.")
			elapsedTime := 0
			for {
				cmd.Printf("%v...", elapsedTime)
				if isReady, _ := utils.IsKFPReady(cmd); isReady {
					cmd.Println("SAME pipeline is ready.")
					return nil
				}
				time.Sleep(5 * time.Second)

				// Printed after sleep is over
				elapsedTime += 5

			}
		} else {
			cmd.Println("SAME deployed but we did not check to see if everything is running. Please do so using 'kubectl get deployments namespace=kubeflow'.")
		}

		return nil

	},
}

func GetClusterInstallMethods() infra.InstallerInterface {
	if os.Getenv("TEST_PASS") == "" {
		return &infra.LiveInstallers{}
	} else {
		return &infra.MockInstallers{}
	}
}

func SetupLocal(cmd *cobra.Command, dc infra.DependencyCheckers, i infra.InstallerInterface) (err error) {
	clusters, err := utils.HasClusters(cmd)
	if err != nil {
		if utils.PrintErrorAndReturnExit(cmd, "error while checking for active clusters using the following command 'kubectl config get-clusters': %v", err) {
			return nil
		}
	} else if len(clusters) < 1 {
		if utils.PrintErrorAndReturnExit(cmd, "We were able to check for current clusters, but you don't have any. Please create a k8s cluster. %v", fmt.Errorf("")) {
			return nil
		}
	}
	context, err := utils.HasContext(cmd)
	if err != nil {
		if utils.PrintErrorAndReturnExit(cmd, "error while checking for current context - using the following command 'kubectl config current-context': %v", err) {
			return nil
		}
	} else if context == "" {
		if utils.PrintErrorAndReturnExit(cmd, "We were able to check for current context, but you don't have any. %v", fmt.Errorf("")) {
			return nil
		}
	}
	log.Traceln("K8s cluster and context detected, proceeding to install KFP.")
	log.Tracef("kubectl path: %v", i.GetKubectlCmd())

	currentContext := dc.WriteCurrentContextToConfig()
	log.Infof("Wrote kubectl current context as: %v", currentContext)

	log.Infof("Cmd: %v", i.GetCmd())

	err = i.InstallKFP()
	if err != nil {
		if utils.PrintError("kfp failed to install: %v", err) {
			return err
		}
	}

	return nil
}

func SetupAKS(cmd *cobra.Command, dc infra.DependencyCheckers, i infra.InstallerInterface) (err error) {
	kubeconfig, err := utils.NewKFPConfig()
	if err != nil {
		return err
	}

	forceCreate := viper.GetBool("force-create")
	if !forceCreate && kubeconfig != nil {
		currentConfig, _ := kubeconfig.ClientConfig()
		if strings.Contains(currentConfig.Host, "azmk8s.io") {
			cmd.Printf("Reusing your existing AKS cluster (%v). Ready to install KFP.", currentConfig.String())
			return nil
		} else {
			if utils.PrintErrorAndReturnExit(cmd, fmt.Sprintf("Your current Kubernetes context (%v) is pointing at a cluster which does not appear to be hosted on AKS. Please update your current context with kubectl config set-context to your AKS cluster, or use --force-create with the same command", currentConfig.ServerName), fmt.Errorf("")) {
				return errors.Errorf("current Kubectl context pointing at non-AKS cluster")
			}
		}
	} else {
		log.Trace("Testing AZ Token")
		hasToken, err := dc.HasValidAzureToken()
		if !hasToken || err != nil {
			return err
		}
		log.Trace("Token passed, testing cluster exists.")
		if clusterCreated, err := dc.IsClusterWithKubeflowCreated(); !clusterCreated || err != nil {
			log.Trace("Cluster does not exist, creating.")
			if err := dc.CreateAKSwithKubeflow(); err != nil {
				return err
			}
			log.Info("Cluster created.")
		}

		log.Trace("Cluster exists, testing to see if storage provisioned.")
		if storageConfigured, err := dc.IsStorageConfigured(); !storageConfigured || err != nil {
			log.Trace("Storage not provisioned, creating.")
			if err := dc.ConfigureStorage(); err != nil {
				return err
			}
			log.Trace("Storage provisioned.")
		}
	}
	cmd.Println("Infrastructure Setup Complete.")
	return nil
}

func init() {
	initCmd.Flags().BoolP("wait", "w", true, "Wait for SAME to be ready before exiting. Can be run or quit with no impact.")
	initCmd.Flags().BoolP("ready", "r", false, "Run a check to see if all SAME components are ready in the cluster.")
	initCmd.Flags().BoolP("force-create", "", false, "Force creation of a new cluster, even if one already exists.")
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
