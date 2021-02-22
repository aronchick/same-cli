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
	"context"
	"fmt"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()

		fmt.Println("test called")

		ctx := context.Background()
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(path) // for example /home/user

		fileURI, _ := filepath.Abs("same.yaml")
		sameConfig, err := ParseSAME(ctx, fileURI)

		if err != nil {
			fmt.Printf("failed to load config: %v", err.Error())
		}

		// Connect to AKS
		resourceGroupName := os.Getenv("SAME_CLUSTER_RG")
		if len(resourceGroupName) == 0 {
			fmt.Printf("expected to have an environment variable named: SAME_CLUSTER_RG")
			return
		}

		clusterName := os.Getenv("SAME_CLUSTER_NAME")
		if len(resourceGroupName) == 0 {
			fmt.Printf("expected to have an environment variable named: SAME_CLUSTER_NAME")
			return
		}

		aksCluster, err := GetAKS(ctx, resourceGroupName, clusterName)

		if err != nil {
			fmt.Print(err.Error())
		}

		if len(args) < 1 {
			fmt.Printf("Please name a method to test. Code must already be included in test.go to handle it.")
			return
		}

		switch args[0] {
		case "DeployorUpdateKubeflow":
			err = DeployorUpdateKubeflow(ctx, resourceGroupName, aksCluster, *sameConfig)
			if err != nil {
				log.Debug(err)
				fmt.Printf("Error deploying Kubeflow: %v\n", err.Error())
			}
		}

		_ = sameConfig
	},
}

func init() {
	RootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
