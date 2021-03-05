package infra

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type MockDependencyCheckers struct {
	_cmd            *cobra.Command
	_kubectlCommand string
	_cmdArgs        []string
}

func (mockDC *MockDependencyCheckers) SetCmd(cmd *cobra.Command) {
	mockDC._cmd = cmd
}

func (mockDC *MockDependencyCheckers) GetCmd() *cobra.Command {
	return mockDC._cmd
}

func (mockDC *MockDependencyCheckers) SetCmdArgs(args []string) {
	mockDC._cmdArgs = args
}

func (mockDC *MockDependencyCheckers) GetCmdArgs() []string {
	return mockDC._cmdArgs
}

func (mockDC *MockDependencyCheckers) SetKubectlCmd(s string) {
	mockDC._kubectlCommand = s
}

func (mockDC *MockDependencyCheckers) GetKubectlCmd() string {
	return mockDC._kubectlCommand
}

func (mockDC *MockDependencyCheckers) HasValidAzureToken(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "invalid-azure-token") {
		return fmt.Errorf("INVALID AZURE TOKEN")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) IsStorageConfigured(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "is-storage-configuration-failed") {
		return fmt.Errorf("IS STORAGE CONFIGURATION FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) ConfigureStorage(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "storage-configuration-failed") {
		return fmt.Errorf("STORAGE CONFIGURATION FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) CreateAKSwithKubeflow(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "create-aks-with-kubeflow-failed") {
		return fmt.Errorf("CREATE AKS WITH KUBEFLOW FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) CheckDependenciesInstalled(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "dependencies-missing") {
		return fmt.Errorf("DEPENDENCIES MISSING")
	}

	return nil
}

func (mockDC *MockDependencyCheckers) IsClusterWithKubeflowCreated(*cobra.Command) (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "is-cluster-with-kubeflow-created-failed") {
		return fmt.Errorf("IS CLUSTER WITH KUBEFLOW CREATED FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) WriteCurrentContextToConfig() string {
	//TODO: Build mock
	return ""
}
