package mocks

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type MockInstallers struct {
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs []string
}

func (mi *MockInstallers) InstallK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	// TODO: Should have a real failure to install k3s message
	k3sCommand, err = mi.DetectK3s("k3s")
	return k3sCommand, err
}

func (mi *MockInstallers) StartK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	// TODO: Should have a real failure to start k3s message
	k3sCommand, err = mi.DetectK3s("k3s")
	return k3sCommand, err
}
func (mi *MockInstallers) DetectK3s(s string) (string, error) {
	if utils.ContainsString(mi._cmdArgs, "k3s-not-detected") {
		return "", fmt.Errorf("K3S NOT DETECTED")
	}

	return "VALID", nil
}

func (mi *MockInstallers) SetCmdArgs(args []string) {
	mi._cmdArgs = args
}

type MockDependencyCheckers struct {
	_cmd            *cobra.Command
	_kubectlCommand string
	_installers     utils.InstallerInterface
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
	mockDC._installers.SetCmdArgs(args)
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

func (mockDC *MockDependencyCheckers) SetInstallers(mi utils.InstallerInterface) {
	mockDC._installers = mi
}

func (mockDC *MockDependencyCheckers) GetInstallers() utils.InstallerInterface {
	return mockDC._installers
}

func (mockDC *MockDependencyCheckers) PrintError(s string, err error) (exit bool) {
	return utils.PrintError(s, err)
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

func (mockDC *MockDependencyCheckers) InstallKFP() (err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "kfp-install-failed") {
		return fmt.Errorf("INSTALL KFP FAILED")
	}

	return nil
}

func (mockDC *MockDependencyCheckers) WriteCurrentContextToConfig() string {
	//TODO: Build mock
	return ""
}
