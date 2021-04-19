package infra

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DependencyCheckers interface {
	CheckDependenciesInstalled() error
	IsKubectlOnPath() (string, error)
	HasValidAzureToken() (bool, error)
	CanConnectToKubernetes() (bool, error)
	HasKubeflowNamespace() (bool, error)
	IsClusterWithKubeflowCreated() (bool, error)
	IsK3sRunning() (bool, error)
	CreateAKSwithKubeflow() error
	IsStorageConfigured() (bool, error)
	ConfigureStorage() error
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
	WriteCurrentContextToConfig() string
}

func GetDependencyCheckers(cmd *cobra.Command, args []string) DependencyCheckers {
	log.Tracef("Current TEST_PASS value: %v", os.Getenv("TEST_PASS"))
	var dc DependencyCheckers = &LiveDependencyCheckers{}
	if os.Getenv("TEST_PASS") != "" {
		dc = &MockDependencyCheckers{}
	}
	dc.SetCmd(cmd)
	dc.SetCmdArgs(args)
	return dc
}
