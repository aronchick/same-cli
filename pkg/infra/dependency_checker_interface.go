package infra

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DependencyCheckers interface {
	// Setup getter/setters
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)

	// Root method
	CheckDependenciesInstalled() error

	// Kubernetes helpers
	IsKubectlOnPath() (string, error)
	HasClusters() ([]string, error)
	HasContext() (string, error)
	CanConnectToKubernetes() (bool, error)
	HasKubeflowNamespace() (bool, error)
	IsKFPReady() (bool, error)
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
