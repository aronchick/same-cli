package cmd

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

// DeployOrUpdateAPipeline takes a sameConfig and communicates with Kubeflow to deploy a piplene
func DeployOrUpdateAPipeline(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI
	return fmt.Errorf("method 'DeployOrUpdateAPipeline' has not yet been implemented")
}
