package infra

import "github.com/spf13/cobra"

type InstallerInterface interface {
	InstallK3s() (string, error)
	PostInstallK3sRunning() error
	InstallKFP() error
	GetKubectlCmd() string
	SetKubectlCmd(string)
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
}
