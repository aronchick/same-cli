package cmd

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

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

// GetAgentPool creates or gets and returns a client for a specific agent pool
func GetAgentPool(ctx context.Context, resourceGroupName, resourceName string, agentPoolNamePrefix string, sameConfig loaders.SameConfig) (agentPool containerservice.AgentPool, err error) {

	// nodePoodName: sameAgentPool
	// cores:
	//     requested: 20
	//     required: 10
	//     minimum_per_machine: 4
	// gpus:
	//     type: V100
	//     per_machine: 1
	// disks:
	//     - name: data_disk
	//       size: 10Gi
	//       volumeMount:
	//         mountPath: "/mnt/data_disk"
	//         name: volume
	//     - name: model_disk
	//       size: 10Gi
	//       volumeMount:
	//         mountPath: "/mnt/model_disk"
	//         name: volume

	experimentSHA := sameConfig.Spec.Resources.NodePoolName
	agentPoolProperties := containerservice.ManagedClusterAgentPoolProfileProperties{}

	numberOfNodes := int32(math.Ceil(float64(sameConfig.Spec.Resources.Cores.Requested) / float64(sameConfig.Spec.Resources.Cores.MinimumPerMachine)))

	agentPoolProperties.Count = to.Int32Ptr(numberOfNodes)
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

	safeAgentPoolString := aksNamingString(fmt.Sprintf("%v_%v", agentPoolNamePrefix, experimentSHA))
	if len(safeAgentPoolString) > 12 {
		safeAgentPoolString = safeAgentPoolString[:12]
	}
	agentPoolName := to.StringPtr(safeAgentPoolString)

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
