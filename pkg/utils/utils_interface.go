package utils

import (
	"os"

	gogetter "github.com/hashicorp/go-getter"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

type UtilsInterface interface {
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)

	// File convenience methods
	GetFile(string, string) error
	Detect(string, string, []gogetter.Detector) (string, error)
	IsRemoteFilePath(string) (bool, error)
	GetConfigFilePath(string) (string, error)

	// Kubernetes interaction methods
	IsEndpointReachable(string) (bool, error)
	IsK3sHealthy() (string, error)
	DetectK3s() (string, error)
	IsK3sRunning() (bool, error)
}

func GetUtils(cmd *cobra.Command, args []string) UtilsInterface {
	log.Tracef("Current TEST_PASS value: %v", os.Getenv("TEST_PASS"))
	log.Tracef("Current GITHUB_ACTIONS value: %v", os.Getenv("GITHUB_ACTIONS"))

	var u UtilsInterface = &UtilsLive{}
	if os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("TEST_PASS") != "" {
		// We're in a GITHUB_ACTION run or a test pass.
		// Should probably combine these somehow and have a way to override during testing if we want to force live testing
		// during a run.
		u = &UtilsMock{}
	}

	u.SetCmd(cmd)
	u.SetCmdArgs(args)
	return u
}
