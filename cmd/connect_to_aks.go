package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

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

func getSubscriptionID() string {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		fmt.Printf("expected to have an environment variable named: AZURE_SUBSCRIPTION_ID")
		os.Exit(1)
	}
	return subscriptionID
}
