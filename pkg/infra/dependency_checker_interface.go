package infra

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DependencyCheckers interface {
	CheckDependenciesInstalled(*cobra.Command) error
	IsKubectlOnPath(*cobra.Command) (string, error)
	HasValidAzureToken(*cobra.Command) (bool, error)
	CanConnectToKubernetes(*cobra.Command) (bool, error)
	HasKubeflowNamespace(*cobra.Command) (bool, error)
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

func GetDependencyCheckers(cmd *cobra.Command, args []string) DependencyCheckers {
	logrus.Tracef("Current TEST_PASS value: %v", os.Getenv("TEST_PASS"))
	var i DependencyCheckers = &LiveDependencyCheckers{}
	if os.Getenv("TEST_PASS") != "" {
		i = &MockDependencyCheckers{}
	}
	i.SetCmd(cmd)
	i.SetCmdArgs(args)
	return i
}
