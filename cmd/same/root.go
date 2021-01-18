package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

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

// Execute executes a specific version of the command
func Execute(version string) {
	ctx := context.Background()

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

	// Get the YAML from disk
	fileURI, _ := filepath.Abs("/home/daaronch/same-cli/same.yaml")
	sameConfig, err := ParseConfig(ctx, version, fileURI)

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
	nodepool, err := GetAgentPool(ctx, resourceGroupName, *aksCluster.Name, "np", *sameConfig)

	log.Debug(err)
	if err != nil {
		fmt.Printf("Error creating agent pool: %v", err.Error())
	}

	_ = nodepool

	// Create or mount a shared disk Azure Storage Gen2
	log.Debug(err)
	if CreateOrAttachDisks(ctx, resourceGroupName, aksCluster, *sameConfig) != nil {
		fmt.Printf("Error creating disks: %v", err.Error())
	}

	// Deploy Kubeflow to the Kubernetes (via Porter?)

	// Deploy a pipeline to the Kubeflow

	// Run against that specific workload

	// Change the parameters and re-run

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
