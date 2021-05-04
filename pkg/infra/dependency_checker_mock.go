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

func (mockDC *MockDependencyCheckers) CheckDependenciesInstalled() (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "dependencies-missing") {
		return fmt.Errorf("DEPENDENCIES MISSING")
	} else if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE) {
		return fmt.Errorf(mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
	}

	return nil
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

func (mockDC *MockDependencyCheckers) HasContext() (currentContext string, err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_MISSING_CONTEXT_PROBE) {
		return "", fmt.Errorf(mocks.DEPENDENCY_CHECKER_MISSING_CONTEXT_RESULT)
	}
	return "VALID_CONTEXT", nil
}

func (mockDC *MockDependencyCheckers) HasClusters() (clusters []string, err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_MISSING_CLUSTERS_PROBE) {
		return []string{}, fmt.Errorf(mocks.DEPENDENCY_CHECKER_MISSING_CLUSTERS_RESULT)
	}
	return []string{"VALID_CLUSTER"}, nil
}

func (mockDC *MockDependencyCheckers) IsKFPReady() (running bool, err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_KFP_NOT_READY_PROBE) {
		return false, fmt.Errorf(mocks.DEPENDENCY_CHECKER_KFP_NOT_READY_RESULT)
	}
	return true, nil
}
