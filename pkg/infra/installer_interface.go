package infra

import (
	"os"

	"github.com/spf13/cobra"
)

type InstallerInterface interface {
	// Instantiation getter/setters
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
	GetKubectlCmd() (string, error)
	SetKubectlCmd(string)

	// KFP helpers
	InstallKFP() error
}

func GetInstallers(cmd *cobra.Command, args []string) InstallerInterface {
	var i InstallerInterface = &LiveInstallers{}
	if os.Getenv("TEST_PASS") != "" {
		i = &MockInstallers{}
	}
	i.SetCmd(cmd)
	i.SetCmdArgs(args)
	return i
}
