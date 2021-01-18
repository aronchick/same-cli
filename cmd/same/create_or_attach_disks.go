package cmd

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

// CreateOrAttachDisks takes a sameConfig and provisions the disk against the cloud. It hands back a handle if it already detects it.
func CreateOrAttachDisks(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI
	return nil
}
