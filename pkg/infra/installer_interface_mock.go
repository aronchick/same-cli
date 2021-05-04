package infra

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type MockInstallers struct {
	_cmd *cobra.Command
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs        []string
	_kubectlCommand string
}

func (mi *MockInstallers) InstallKFP() (err error) {
	if utils.ContainsString(mi.GetCmdArgs(), mocks.DEPENDENCY_CHECKER_KFP_INSTALL_FAILED_PROBE) {
		return fmt.Errorf(mocks.DEPENDENCY_CHECKER_KFP_INSTALL_FAILED_RESULT)
	}

	return nil
}

func (mi *MockInstallers) GetCmd() *cobra.Command {
	return mi._cmd
}

func (mi *MockInstallers) SetCmd(cmd *cobra.Command) {
	mi._cmd = cmd
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

func (mi *MockInstallers) GetKubectlCmd() (string, error) {
	return mi._kubectlCommand, nil
}
