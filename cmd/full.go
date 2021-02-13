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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
)

// fullCmd represents the full command
var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("full called")
		ExecuteFull()

	},
}

func init() {
	RootCmd.AddCommand(fullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Version
var (
	Version string = "0.0.1"
)

func aksNamingString(sourceString string) (returnString string) {
	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToLower(reg.ReplaceAllString(sourceString, ""))
}

func printVersion() {
	log.Infof("Go Version: %s", runtime.Version())
	log.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Infof("same version: %v", Version)
}

// ExecuteFull executes a full provisioning - it will likely be slow
func ExecuteFull() {
	ctx := context.Background()

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

	// Get the YAML from disk
	fileURI, _ := filepath.Abs("/home/daaronch/same-cli/same.yaml")
	sameConfig, err := ParseConfig(ctx, fileURI)

	if err != nil {
		fmt.Printf("failed to load config: %v", err.Error())
	}

	// Connect to AKS
	resourceGroupName := os.Getenv("SAME_CLUSTER_RG")
	if len(resourceGroupName) == 0 {
		fmt.Printf("expected to have an environment variable named: SAME_CLUSTER_RG")
		os.Exit(1)
	}

	clusterName := os.Getenv("SAME_CLUSTER_NAME")
	if len(resourceGroupName) == 0 {
		fmt.Printf("expected to have an environment variable named: SAME_CLUSTER_NAME")
		os.Exit(1)
	}

	aksCluster, err := GetAKS(ctx, resourceGroupName, clusterName)

	if err != nil {
		fmt.Print(err.Error())
	}

	// Create or get Node Pool
	// Marking false while debugging
	nodePool := containerservice.AgentPool{}
	if true {
		nodePool, err = GetAgentPool(ctx, resourceGroupName, *aksCluster.Name, "np", *sameConfig)
	}

	_ = nodePool

	log.Debug(err)
	if err != nil {
		fmt.Printf("Error creating agent pool: %v\n", err.Error())
	}

	// Create or mount a shared disk Azure Storage Gen2
	err = CreateOrAttachDisks(ctx, resourceGroupName, aksCluster, *sameConfig)
	if err != nil {
		log.Debug(err)
		fmt.Printf("Error creating disks: %v\n", err.Error())
	}

	// Deploy Kubeflow to the Kubernetes (via Porter?)
	err = DeployorUpdateKubeflow(ctx, resourceGroupName, aksCluster, *sameConfig)
	if err != nil {
		log.Debug(err)
		fmt.Printf("Error deploying Kubeflow: %v\n", err.Error())
	}

	// Run against that specific workload
	err = RunAnExperiment(ctx, resourceGroupName, aksCluster, *sameConfig)
	if err != nil {
		log.Debug(err)
		fmt.Printf("Error running against a specific workload: %v\n", err.Error())
	}

	// Change the parameters and re-run
	err = RunAnExperiment(ctx, resourceGroupName, aksCluster, *sameConfig)
	if err != nil {
		log.Debug(err)
		fmt.Printf("Error running against a specific workload with new params: %v\n", err.Error())
	}

	// See what happens when you do that all with systems already in place (e.g. can we check to see if something is already installed)

	// Just making it absolute for now - obviously needs changing for anyone else's machine

	_ = sameConfig
}

// If 'create':
// 		If a file is provided for a SAME - process it
// 		-- will start with just a yaml that describes a bunch of sub steps

// 		Connect to Kubernetes
// 		-- elegant fail if not

// 		Check the status of the cluster - are there enough machines of the type we need?

// 				-- 	Provision machines in a new node pool if not

// 		Deploy via Porter any step named "Kubeflow"

// 				-- Only support specific versions for now
// 				-- What kind of support can we have for specific services (e.g. the TF CRD? what version is it? )

// 		Get the credentials for KF for the SDK

// 		See if the pipeline named in the SAME config.yaml is deployed (with the same version)
// 				-- If not, deploy it
// 				-- Should be available locally in a .tgz to start?

// If 'run':

// 		See if the pipeline named is deployed (with the same version)
//		If not, fail elegantly

//		If deployed, run with the correct parameters.
//		-- Check to see if resources have gone away?
//		Say it's deployed -- give a URL to the dashboard

//		Do something when it's done? probably not.

// If 'export':

//		EASIER: Grab all yaml/etc necessary to create the pipeline? Argo?
//		HARD: Grab KF configuration necessary to run?
//		HARDER: Grab HW requirements
//		HARDEST: Grab metadata information

// 		Create a Yaml/package out of it all.
