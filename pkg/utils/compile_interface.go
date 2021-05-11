package utils

import (
	"os"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	log "github.com/sirupsen/logrus"
)

type CompileInterface interface {
	FindAllSteps(string) ([]FoundStep, error)
	CombineCodeSlicesToSteps([]FoundStep) (map[string]CodeBlock, error)
	CreateRootFile(map[string]CodeBlock, loaders.SameConfig) (string, error)
	ConvertNotebook(string, string) (string, error)
	WriteStepFiles(string, map[string]CodeBlock) error
}

type CodeBlock struct {
	Step_Identifier     string
	Code                string
	Parameters          map[string]string
	Packages_To_Install map[string]string
	Tags                map[string]string
	Cache_Value         string
}

type FoundStep struct {
	index       int
	step_name   string
	tags        []string
	code_slice  string
	cache_value string
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
