package utils

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/mocks"
	gogetter "github.com/hashicorp/go-getter"
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

func (u *UtilsMock) IsK3sRunning() (bool, error) {
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

func (u *UtilsMock) Detect(absFilePath string, cwd string, detectors []gogetter.Detector) (string, error) {
	if ContainsString(u.GetCmdArgs(), mocks.UTILS_TEST_BAD_CONFIG_FILE_DETECT_PROBE) {
		return "", fmt.Errorf(mocks.UTILS_TEST_BAD_CONFIG_FILE_DETECT_RESULT)
	}
	return "VALID_FILE_PATH", nil
}

func (u *UtilsMock) GetFile(tempSameFilename string, corrected_url string) error {
	if ContainsString(u.GetCmdArgs(), mocks.UTILS_TEST_BAD_RETRIEVE_SAME_FILE_PROBE) {
		return fmt.Errorf(mocks.UTILS_TEST_BAD_RETRIEVE_SAME_FILE_RESULT)
	}
	return nil
}

func (u *UtilsMock) GetConfigFilePath(s string) (string, error) {
	// Temporary until we start mocking
	if ContainsString(u.GetCmdArgs(), mocks.UTILS_TEST_BAD_CONFIG_FILE_DETECT_PROBE) {
		return "", fmt.Errorf(mocks.UTILS_TEST_BAD_CONFIG_FILE_DETECT_RESULT)
	} else if ContainsString(u.GetCmdArgs(), mocks.UTILS_TEST_BAD_RETRIEVE_SAME_FILE_PROBE) {
		return "", fmt.Errorf(mocks.UTILS_TEST_BAD_RETRIEVE_SAME_FILE_RESULT)
	}

	// Need to fall back to live if haven't caught errors
	ul := &UtilsLive{}
	return ul.GetConfigFilePath(s)
}

func (u *UtilsMock) IsK3sHealthy() (string, error) {
	// Temporary until we start mocking
	ul := &UtilsLive{}
	return ul.IsK3sHealthy()
}

func (u *UtilsMock) IsRemoteFilePath(s string) (bool, error) {
	// Temporary until we start mocking
	ul := &UtilsLive{}
	return ul.IsRemoteFilePath(s)
}

func (u *UtilsMock) IsEndpointReachable(s string) (bool, error) {
	// Temporary until we start mocking
	ul := &UtilsLive{}
	return ul.IsEndpointReachable(s)
}
