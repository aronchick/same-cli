package infra

import (
	"fmt"
	"os"

	"github.com/azure-octo/same-cli/pkg/mocks"
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

func (mockDC *MockDependencyCheckers) HasValidAzureToken() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "invalid-azure-token") {
		return false, fmt.Errorf("INVALID AZURE TOKEN")
	}
	return true, nil
}

func (mockDC *MockDependencyCheckers) IsStorageConfigured() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "is-storage-configuration-failed") {
		return false, fmt.Errorf("IS STORAGE CONFIGURATION FAILED")
	}
	return true, nil
}

func (mockDC *MockDependencyCheckers) ConfigureStorage() (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "storage-configuration-failed") {
		return fmt.Errorf("STORAGE CONFIGURATION FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) CreateAKSwithKubeflow() (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "create-aks-with-kubeflow-failed") {
		return fmt.Errorf("CREATE AKS WITH KUBEFLOW FAILED")
	}
	return nil
}

func (mockDC *MockDependencyCheckers) CheckDependenciesInstalled() (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "dependencies-missing") {
		return fmt.Errorf("DEPENDENCIES MISSING")
	} else if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE) {
		return fmt.Errorf(mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
	}

	return nil
}

func (mockDC *MockDependencyCheckers) IsClusterWithKubeflowCreated() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "is-cluster-with-kubeflow-created-failed") {
		return false, fmt.Errorf("IS CLUSTER WITH KUBEFLOW CREATED FAILED")
	}
	return true, nil
}
func (mockDC *MockDependencyCheckers) IsK3sRunning() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "k3s-is-not-running") {
		return false, fmt.Errorf("K3S NOT RUNNING")
	}
	return true, nil
}

func (mockDC *MockDependencyCheckers) IsKubectlOnPath() (string, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE) ||
		os.Getenv("MISSING_KUBECTL") != "" {
		return "", fmt.Errorf(mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
	}
	return "kubectl", nil
}

func (mockDC *MockDependencyCheckers) CanConnectToKubernetes() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_CANNOT_CONNECT_TO_K8S_PROBE) {
		return false, fmt.Errorf(mocks.DEPENDENCY_CHECKER_CANNOT_CONNECT_TO_K8S_RESULT)
	} else if utils.ContainsString(mockDC.GetCmdArgs(), mocks.MOCK_CONNECT_TO_KUBERNETES_CLUSTER) {
		return true, nil
	}

	// Fall back to connecting to a real Kubernetes cluster
	var li = &LiveDependencyCheckers{}
	li.SetCmd(mockDC.GetCmd())
	li.SetCmdArgs(mockDC.GetCmdArgs())
	return li.CanConnectToKubernetes()
}

func (mockDC *MockDependencyCheckers) HasKubeflowNamespace() (bool, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_MISSING_KUBEFLOW_NAMESPACE_PROBE) {
		return false, fmt.Errorf(mocks.DEPENDENCY_CHECKER_MISSING_KUBEFLOW_NAMESPACE_RESULT)
	}
	return true, nil
}

func (mockDC *MockDependencyCheckers) WriteCurrentContextToConfig() string {
	//TODO: Build mock
	return ""
}
