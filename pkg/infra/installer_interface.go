package infra

import "github.com/spf13/cobra"

type InstallerInterface interface {
	InstallK3s(*cobra.Command) (string, error)
	StartK3s(*cobra.Command) (string, error)
	DetectK3s(string) (string, error)
	InstallKFP(*cobra.Command) error
	GetKubectlCmd() string
	SetKubectlCmd(string)
}
