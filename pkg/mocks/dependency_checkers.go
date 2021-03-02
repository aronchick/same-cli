package mocks

import (
	"fmt"
	"os/user"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type MockInstallers struct {
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs []string
}

func (mi *MockInstallers) InstallK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	return mi.DetectK3s("k3s")
}

func (mi *MockInstallers) StartK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	return mi.DetectK3s("k3s")
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

func (mockDC *MockDependencyCheckers) DetectDockerBin(s string) (string, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "no-docker-path") {
		return "", fmt.Errorf("not find docker in your PATH")
	}

	return "VALID_PATH", nil
}

func (mockDC *MockDependencyCheckers) DetectDockerGroup(s string) (*user.Group, error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "no-docker-group-on-system") {
		return nil, user.UnknownGroupError("NOT_FOUND")
	}

	return &user.Group{Gid: "1001", Name: "docker"}, nil
}

func (mockDC *MockDependencyCheckers) PrintError(s string, err error) (exit bool) {
	message := fmt.Errorf(s, err)
	mockDC.GetCmd().Printf(message.Error())
	log.Fatalf(message.Error())

	return true
}

func (mockDC *MockDependencyCheckers) GetUserGroups(u *user.User) (returnGroups []string, err error) {
	if utils.ContainsString(mockDC.GetCmdArgs(), "cannot-retrieve-groups") {
		return nil, fmt.Errorf("CANNOT RETRIEVE GROUPS")
	} else if utils.ContainsString(mockDC.GetCmdArgs(), "not-in-docker-group") {
		return []string{}, nil
	}

	return []string{"docker"}, nil
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
