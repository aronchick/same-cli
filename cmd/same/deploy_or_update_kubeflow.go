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
	return fmt.Errorf("method 'DeployorUpdateKubeflow' has not yet been implemented")
}
