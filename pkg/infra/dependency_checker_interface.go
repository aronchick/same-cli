package infra

import (
	"github.com/spf13/cobra"
)

type DependencyCheckers interface {
	CheckDependenciesInstalled(*cobra.Command) error
	HasValidAzureToken(*cobra.Command) (bool, error)
	IsClusterWithKubeflowCreated(*cobra.Command) (bool, error)
	IsK3sRunning(*cobra.Command) (bool, error)
	CreateAKSwithKubeflow(*cobra.Command) error
	IsStorageConfigured(*cobra.Command) (bool, error)
	ConfigureStorage(*cobra.Command) error
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
	WriteCurrentContextToConfig() string
}
