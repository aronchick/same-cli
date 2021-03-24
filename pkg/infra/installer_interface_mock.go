package infra

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type MockInstallers struct {
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs        []string
	_kubectlCommand string
}

func (mi *MockInstallers) InstallK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	if utils.ContainsString(mi.GetCmdArgs(), "k3s-install-failed") {
		return "", fmt.Errorf("INSTALL K3S FAILED")
	}

	return "VALID", nil
}

func (mi *MockInstallers) PostInstallK3sRunning(cmd *cobra.Command) error {
	if utils.ContainsString(mi._cmdArgs, mocks.INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_PROBE) {
		return fmt.Errorf(mocks.INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_RESULT)
	}

	return nil
}

func (mi *MockInstallers) InstallKFP(cmd *cobra.Command) (err error) {
	if utils.ContainsString(mi.GetCmdArgs(), "kfp-install-failed") {
		return fmt.Errorf("INSTALL KFP FAILED")
	}

	return nil
}

func (mi *MockInstallers) GetCmdArgs() []string {
	return mi._cmdArgs
}

func (mi *MockInstallers) SetCmdArgs(args []string) {
	mi._cmdArgs = args
}
func (mi *MockInstallers) SetKubectlCmd(s string) {
	mi._kubectlCommand = s
}

func (mi *MockInstallers) GetKubectlCmd() string {
	return mi._kubectlCommand
}
