package main

import (
	"flag"
	"runtime"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	// az "github.com/Azure/go-autorest/autorest/azure/auth"
)

// Version
var (
	Version string = "0.0.1"
)

func printVersion() {
	log.Infof("Go Version: %s", runtime.Version())
	log.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Infof("same version: %v", Version)
}

func getAKSClient() (containerservice.ManagedClustersClient, error) {
	aksClient := containerservice.NewManagedClustersClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	aksClient.Authorizer = auth
	aksClient.AddToUserAgent(config.UserAgent())
	aksClient.PollingDuration = time.Hour * 1
	return aksClient, nil
}

// GetAKS returns an existing AKS cluster given a resource group name and resource name
func GetAKS(ctx context.Context, resourceGroupName, resourceName string) (c containerservice.ManagedCluster, err error) {
	aksClient, err := getAKSClient()
	if err != nil {
		return c, fmt.Errorf("cannot get AKS client: %v", err)
	}

	c, err = aksClient.Get(ctx, resourceGroupName, resourceName)
	if err != nil {
		return c, fmt.Errorf("cannot get AKS managed cluster %v from resource group %v: %v", resourceName, resourceGroupName, err)
	}

	return c, nil
}

func main() {
	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

	// Here's all the steps we need to build

	// Parse the command line flags

	// // create a VirtualNetworks client
	// vnetClient := network.NewVirtualNetworksClient("<subscriptionID>")

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
