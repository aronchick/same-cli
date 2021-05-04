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
	"time"

	"github.com/azure-octo/same-cli/pkg/infra"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes all base services for deploying a SAME (Kubeflow, etc)",
	Long:  `Initializes all base services for deploying a SAME (Kubeflow, etc). Longer Description.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var i = infra.GetInstallers(cmd, args)
		var dc = infra.GetDependencyCheckers(cmd, args)

		fmt.Printf("args: %v\n", args)

		canConnect, err := dc.CanConnectToKubernetes()
		if err != nil || !canConnect {
			return fmt.Errorf("Could not connect to Kubernetes. Check your settings with 'kubectl get version': %v", err)
		}

		clusters, err := dc.HasClusters()
		if err != nil {
			return fmt.Errorf("error while checking for active clusters using the following command 'kubectl config get-clusters': %v", err)
		}

		if len(clusters) < 1 {
			return fmt.Errorf("We were able to check for current clusters, but you don't have any. Please create a k8s cluster.")
		}

		context, err := dc.HasContext()
		if err != nil {
			return fmt.Errorf("error while checking for current context - using the following command 'kubectl config current-context': %v", err)
		}

		if context == "" {
			return fmt.Errorf("We were able to check for current context, but you don't have any. %v", fmt.Errorf(""))
		}
		log.Traceln("K8s cluster and context detected, proceeding to install KFP.")

		log.Infof("Cmd: %v", cmd)

		err = i.InstallKFP()
		if err != nil {
			return fmt.Errorf("kfp failed to install: %v", err)
		}

		cmd.Println("Your installation is complete and running!")
		shouldWait, _ := cmd.PersistentFlags().GetBool("wait")
		if shouldWait {
			cmd.Println("Waiting for SAME to become ready... often takes > 5 minutes. \n You can cancel at any time and it will not affect setup. Use 'same init --ready' to check manually.")
			elapsedTime := 0
			for {
				cmd.Printf("%v...", elapsedTime)
				if isReady, _ := dc.IsKFPReady(); isReady {
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
