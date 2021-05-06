package utils

import "github.com/azure-octo/same-cli/cmd/sameconfig/loaders"

type CompileMock struct {
}

func (c *CompileMock) FindAllSteps(convertedText string) (steps [][]string, code_slices []string, err error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.FindAllSteps(convertedText)
}

func (c *CompileMock) CombineCodeSlicesToSteps(stepsFound [][]string, codeSlices []string) (CodeBlocks, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.CombineCodeSlicesToSteps(stepsFound, codeSlices)
}

func (c *CompileMock) CreateRootFile(aggregatedSteps CodeBlocks, sameConfigFile loaders.SameConfig) (string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.CreateRootFile(aggregatedSteps, sameConfigFile)
}

func (c *CompileMock) WriteStepFiles(compiledDir string, aggregatedSteps CodeBlocks) error {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.WriteStepFiles(compiledDir, aggregatedSteps)
}
