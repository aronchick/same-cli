package utils

import (
	"os"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	log "github.com/sirupsen/logrus"
)

type CompileInterface interface {
	ConfirmPackages(loaders.SameConfig) (map[string]string, error)
	FindAllSteps(string) ([]FoundStep, error)
	CombineCodeSlicesToSteps([]FoundStep) (map[string]CodeBlock, error)
	CreateRootFile(string, map[string]CodeBlock, loaders.SameConfig) (string, error)
	ConvertNotebook(string, string) (string, error)
	WriteStepFiles(string, string, map[string]CodeBlock) (map[string]map[string]string, error)
	WriteSupportFiles(string, []string) error
}

type CodeBlock struct {
	StepIdentifier    string
	Code              string
	Parameters        map[string]string
	PackagesToInstall map[string]string
	Tags              map[string]string
	CacheValue        string
	EnvironmentName   string
}

type FoundStep struct {
	Index           int
	StepName        string
	Tags            []string
	CodeSlice       string
	CacheValue      string
	EnvironmentName string
}

type RootFile struct {
	StepImports             []string
	RootParameterString     string
	Steps                   map[string]RootStep
	ExperimentName          string
	StepString              string
	GlobalPackagesToInstall string
}

type RootStep struct {
	Name                string
	PackageString       string
	ContextVariableName string
	CacheValue          string
	PreviousStep        string
}

func GetCompileFunctions() CompileInterface {
	log.Tracef("Current TEST_PASS value: %v", os.Getenv("TEST_PASS"))
	log.Tracef("Current GITHUB_ACTIONS value: %v", os.Getenv("GITHUB_ACTIONS"))

	var c CompileInterface = &CompileLive{}
	if os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("TEST_PASS") != "" {
		// We're in a GITHUB_ACTION run or a test pass.
		// Should probably combine these somehow and have a way to override during testing if we want to force live testing
		// during a run.
		c = &CompileMock{}
	}

	return c
}
