package utils

import (
	"os"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	log "github.com/sirupsen/logrus"
)

type CompileInterface interface {
	FindAllSteps(string) ([][]string, []string, error)
	CombineCodeSlicesToSteps([][]string, []string) (CodeBlocks, error)
	CreateRootFile(CodeBlocks, loaders.SameConfig) (string, error)
	WriteStepFiles(string, CodeBlocks) error
}

type CodeBlock struct {
	step_identifier     string
	code                string
	parameters          map[string]string
	packages_to_install []string
}

type CodeBlocks map[string]*CodeBlock

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
