package infra

import (
	"github.com/spf13/cobra"
)

type DependencyCheckers interface {
	CheckDependenciesInstalled(*cobra.Command) error
	HasValidAzureToken(*cobra.Command) error
	IsClusterWithKubeflowCreated(*cobra.Command) error
	CreateAKSwithKubeflow(*cobra.Command) error
	IsStorageConfigured(*cobra.Command) error
	ConfigureStorage(*cobra.Command) error
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
	WriteCurrentContextToConfig() string
}
