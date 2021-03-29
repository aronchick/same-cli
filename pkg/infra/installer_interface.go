package infra

import "github.com/spf13/cobra"

type InstallerInterface interface {
	InstallK3s(*cobra.Command) (string, error)
	PostInstallK3sRunning(*cobra.Command) error
	InstallKFP(*cobra.Command) error
	GetKubectlCmd(*cobra.Command) string
	SetKubectlCmd(string)
	GetCmdArgs() []string
	SetCmdArgs([]string)
}
