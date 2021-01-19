package cmd

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

// DeployorUpdateKubeflow takes a sameConfig and provisions or updates Kubeflow
func DeployorUpdateKubeflow(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI
	// porter explain --tag ghcr.io/squillace/aks-kubeflow:v0.3.1
	// create an SP per the docs: az ad sp create-for-rbac -n "kubeconfig-read" --role "Azure Kubernetes Service Cluster User Role" --scopes <cluster resource id>
	az ad sp create-for-rbac -n "kubeconfig-read" --role "Azure Kubernetes Service Cluster User Role" --scopes <cluster resource id>

	// porter creds generate --tag ghcr.io/squillace/aks-kubeflow:v0.3.1
	// for each step, choose "specific value" and enter the value from the SP creation
	// porter creds list will show you your creds
	// then install:
	// ` porter install -c kubeflow --tag ghcr.io/squillace/aks-kubeflow:v0.3.1 --param AZURE_RESOURCE_GROUP=winlin --param CLUSTER_NAME=kubeflow`
	return fmt.Errorf("method 'DeployorUpdateKubeflow' has not yet been implemented")
}
