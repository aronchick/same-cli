package utils

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/spf13/cobra"
)

type UtilsMock struct {
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs []string
	_cmd     *cobra.Command
}

func (u *UtilsMock) DetectK3s() (string, error) {

	if ContainsString(u.GetCmdArgs(), "k3s-not-detected") {
		return "", fmt.Errorf("K3S NOT DETECTED")
	}

	return "VALID", nil
}

func (u *UtilsMock) IsK3sRunning(cmd *cobra.Command) (bool, error) {
	if ContainsString(u.GetCmdArgs(), mocks.UTILS_TEST_K3S_RUNNING_FAILED_PROBE) {
		return false, fmt.Errorf(mocks.UTILS_TEST_K3S_RUNNING_FAILED_RESULT)
	}
	return true, nil
}

func (u *UtilsMock) GetCmdArgs() []string {
	return u._cmdArgs
}

func (u *UtilsMock) SetCmdArgs(args []string) {
	u._cmdArgs = args
}

func (u *UtilsMock) SetCmd(cmd *cobra.Command) {
	u._cmd = cmd
}

func (u *UtilsMock) GetCmd() *cobra.Command {
	return u._cmd
}
