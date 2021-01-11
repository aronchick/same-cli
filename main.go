package main

import (
	"flag"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
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

func main() {
	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	printVersion()

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
}
