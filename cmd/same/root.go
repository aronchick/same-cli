package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// Version
var (
	Version string = "0.0.1"
)

func cleanString(sourceString string) (returnString string) {
	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(sourceString, "")
}

func printVersion() {
	log.Infof("Go Version: %s", runtime.Version())
	log.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Infof("same version: %v", Version)
}

func getAKSClient(subscriptionID string) (aksClient containerservice.ManagedClustersClient, err error) {
	aksClient = containerservice.NewManagedClustersClient(subscriptionID)

	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		fmt.Println(err)
		return aksClient, fmt.Errorf("authorizer is nil for an unknown reason")

	}

	fmt.Println("Auth: Successful")
	aksClient.Authorizer = authorizer

	aksClient.PollingDuration = time.Hour * 1
	return aksClient, nil
}

func getAgentPoolClient(subscriptionID string) (agentPoolClient containerservice.AgentPoolsClient, err error) {
	agentPoolClient = containerservice.NewAgentPoolsClient(subscriptionID)

	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		fmt.Println(err)
		return agentPoolClient, fmt.Errorf("authorizer is nil for an unknown reason")

	}

	fmt.Println("Agent Pool Client Authorizer Assigned: Successful")
	agentPoolClient.Authorizer = authorizer

	agentPoolClient.PollingDuration = time.Hour * 1
	return agentPoolClient, nil

}

// GetAKS returns an existing AKS cluster given a resource group name and resource name
func GetAKS(ctx context.Context, resourceGroupName, resourceName string) (c containerservice.ManagedCluster, err error) {
	aksClient, err := getAKSClient(getSubscriptionID())
	if err != nil {
		return c, fmt.Errorf("cannot get AKS client: %v", err)
	}

	c, err = aksClient.Get(ctx, resourceGroupName, resourceName)
	if err != nil {
		return c, fmt.Errorf("cannot get AKS managed cluster %v from resource group %v: %v", resourceName, resourceGroupName, err)
	}

	return c, nil
}

// GetAgentPool creates or gets and returns a client for a specific agent pool
func GetAgentPool(ctx context.Context, resourceGroupName, resourceName string, agentPoolNamePrefix string) (agentPool containerservice.AgentPool, err error) {
	experimentSHA := "a9eou0aue"
	agentPoolProperties := containerservice.ManagedClusterAgentPoolProfileProperties{}
	agentPoolProperties.Count = to.Int32Ptr(5)
	agentPoolProperties.VMSize = containerservice.StandardDS3V2
	agentPoolProperties.OsDiskSizeGB = to.Int32Ptr(30)
	agentPoolProperties.OsDiskType = containerservice.OSDiskType("Ephemeral")

	tags := make(map[string]*string)

	// Just creating the below tag for future use. Will also be useful for bulk deleting if things get stuck around.
	tags["same_created_agent_pool_tag"] = to.StringPtr(fmt.Sprintf("%v", experimentSHA))
	agentPoolProperties.Tags = tags

	// Same with labels
	labels := make(map[string]*string)
	labels["same_created_agent_pool_label"] = to.StringPtr(fmt.Sprintf("%v", experimentSHA))
	agentPoolProperties.NodeLabels = labels

	// // When we are ready to enable scaling - TODO: Just starting with 5 for now
	// agentPoolProfile.MaxCount = Int32(1)
	// agentPoolProfile.MinCount = Int32(1)
	// agentPoolProfile.EnableAutoScaling = Bool(true)

	// Could also enable scaling into spot
	// 	// SpotMaxPrice - SpotMaxPrice to be used to specify the maximum price you are willing to pay in US Dollars. Possible values are any decimal value greater than zero or -1 which indicates default price to be up-to on-demand.
	// 	SpotMaxPrice *float64 `json:"spotMaxPrice,omitempty"`

	agentPoolClient, err := getAgentPoolClient(getSubscriptionID())
	agentPoolName := to.StringPtr(cleanString(fmt.Sprintf("%v_%v", agentPoolNamePrefix, experimentSHA)[:12]))

	if err != nil {
		return agentPool, fmt.Errorf("cannot provision agentPool named '%v' on cluster '%v': %v", agentPoolName, resourceName, err)
	}

	agentPool = containerservice.AgentPool{}
	agentPool.Name = agentPoolName

	// agentPool.Tags = tags
	// agentPool.NodeLabels = labels
	agentPool.ManagedClusterAgentPoolProfileProperties = &agentPoolProperties

	future, err := agentPoolClient.CreateOrUpdate(ctx, resourceGroupName, resourceName, *agentPoolName, agentPool)

	if err != nil {
		return agentPool, fmt.Errorf("cannot update create or update agentpool: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, agentPoolClient.Client)
	if err != nil {
		fmt.Print(err.Error())
		return agentPool, fmt.Errorf("cannot update create or update agentpool future response: %v", err)
	}

	// Should watch --
	// 	// ProvisioningState - READ-ONLY; The current deployment or provisioning state, which only appears in the response.
	// 	ProvisioningState *string `json:"provisioningState,omitempty"`

	return future.Result(agentPoolClient)
}

func getSubscriptionID() string {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		fmt.Printf("expected to have an environment variable named: AZURE_SUBSCRIPTION_ID")
		os.Exit(1)
	}
	return subscriptionID
}

// Execute executes a specific version of the command
func Execute(version string) {
	ctx := context.Background()

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

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

	agentPool, err := GetAgentPool(ctx, resourceGroupName, clusterName, "ap")

	log.Debug(err)
	if err != nil {
		fmt.Printf("Error creating agent pool: %v", err.Error())
	}

	_ = aksCluster
	_ = agentPool

	// Here's all the steps we need to build

	// Parse the command line flags

	// // create an authorizer from env vars or Azure Managed Service Idenity
	// authorizer, err := auth.NewAuthorizerFromEnvironment()
	// if err == nil {
	// 	vnetClient.Authorizer = authorizer
	// }

}

// // call the VirtualNetworks CreateOrUpdate API
// vnetClient.CreateOrUpdate(context.Background(),
// ...

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

// namespace, err := k8sutil.GetWatchNamespace()
// if err != nil {
// 	log.Errorf("Failed to get watch namespace. Error %v.", err)
// 	os.Exit(1)
// }

// // Get a config to talk to the apiserver
// cfg, err := config.GetConfig()
// if err != nil {
// 	log.Errorf("Error: %v.", err)
// 	os.Exit(1)
// }

// // Create a new Cmd to provide shared dependencies and start components
// mgr, err := manager.New(cfg, manager.Options{
// 	// Watch all namespace
// 	Namespace:          "",
// 	MapperProvider:     restmapper.NewDynamicRESTMapper,
// 	MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
// })
// if err != nil {
// 	log.Errorf("Error: %v.", err)
// 	os.Exit(1)
// }

// log.Info("Registering Components.")

// // Setup Scheme for all resources
// if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
// 	log.Errorf("Error: %v.", err)
// 	os.Exit(1)
// }

// // Setup all Controllers
// if err := controller.AddToManager(mgr); err != nil {
// 	log.Errorf("Error: %v.", err)
// 	os.Exit(1)
// }

// if err = serveCRMetrics(cfg); err != nil {
// 	log.Errorf("Could not generate and serve custom resource metrics. Error: %v.", err.Error())
// }

// // Add to the below struct any other metrics ports you want to expose.
// servicePorts := []v1.ServicePort{
// 	{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
// 	{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
// }
// // Create Service object to expose the metrics port(s).
// service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
// if err != nil {
// 	log.Errorf("Could not create metrics Service. Error: %v.", err.Error())
// }

// // CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
// // necessary to configure Prometheus to scrape metrics from this operator.
// services := []*v1.Service{service}
// _, err = metrics.CreateServiceMonitors(cfg, namespace, services)
// if err != nil {
// 	log.Errorf("Could not create ServiceMonitor object. Error: %v.", err.Error())
// 	// If this operator is deployed to a cluster without the prometheus-operator running, it will return
// 	// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
// 	if err == metrics.ErrServiceMonitorNotPresent {
// 		log.Errorf("Install prometheus-operator in your cluster to create ServiceMonitor objects. Error: %v.", err.Error())
// 	}
// }

// log.Infof("Starting the Cmd.")

// // Start the Cmd
// if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
// 	log.Errorf("Manager exited non-zero. Error: %v.", err)
// 	os.Exit(1)
// }
// }
