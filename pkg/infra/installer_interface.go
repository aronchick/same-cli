package infra

import "github.com/spf13/cobra"

type InstallerInterface interface {
	InstallK3s(*cobra.Command) (string, error)
	DetectK3s(string) (string, error)
	PostInstallK3sRunning(*cobra.Command) error
	InstallKFP(*cobra.Command) error
	GetKubectlCmd() string
	SetKubectlCmd(string)
	GetCmdArgs() []string
	SetCmdArgs([]string)
}
