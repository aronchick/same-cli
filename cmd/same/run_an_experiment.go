package cmd

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

// RunAnExperiment takes a sameConfig and communicates with Kubeflow to run a piplene
func RunAnExperiment(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI
	return fmt.Errorf("method 'RunAnExperiment' has not yet been implemented")
}
