package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"

	"os/exec"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randString(length int) string {
	return randStringWithCharset(length, charset)
}

// DeployorUpdateKubeflow takes a sameConfig and provisions or updates Kubeflow
func DeployorUpdateKubeflow(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI
	// porter explain --tag ghcr.io/squillace/aks-kubeflow:v0.3.1
	// create an SP per the docs: az ad sp create-for-rbac -n "kubeconfig-read" --role "Azure Kubernetes Service Cluster User Role" --scopes <cluster resource id>

	// 		Args:   []string{"ad sp create-for-rbac", "-n", "kubeconfig-read", "--role", "\"Azure Kubernetes Service Cluster User Role\"", "--scopes", "daron-kf-cluster"},

	reg, err := regexp.Compile(`\/subscriptions\/[^\/]+\/resourcegroups\/[^\/]+`)
	if err != nil {
		return fmt.Errorf("unable to parse regex: %v", err.Error())
	}

	scope := reg.FindString(*aksCluster.ID)

	cmd := &exec.Cmd{
		Path:   "/usr/bin/az",
		Args:   []string{fmt.Sprintf("ad sp create-for-rbac -n kubeconfig-read-%v --role \"Azure Kubernetes Service Cluster User Role\" --scopes '%v'", randString(5), scope)},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	// run `cmd` in background
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start the cmd: %v", err)
	}

	foo := 0
	// do something else
	for i := 1; i < 300000; i++ {
		foo++
	}

	fmt.Println("Made it through the waiting")

	// wait `cmd` until it finishes
	err = cmd.Wait()
	fmt.Println(cmd.String())
	fmt.Println(cmd.Args)
	if err != nil {
		return fmt.Errorf("error while waiting: %v", err)
	}

	// porter creds generate --tag ghcr.io/squillace/aks-kubeflow:v0.3.1
	// for each step, choose "specific value" and enter the value from the SP creation
	// porter creds list will show you your creds
	// then install:
	// ` porter install -c kubeflow --tag ghcr.io/squillace/aks-kubeflow:v0.3.1 --param AZURE_RESOURCE_GROUP=winlin --param CLUSTER_NAME=kubeflow`
	return fmt.Errorf("method 'DeployorUpdateKubeflow' has not yet been implemented")
}
