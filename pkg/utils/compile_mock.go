package utils

import "github.com/azure-octo/same-cli/cmd/sameconfig/loaders"

type CompileMock struct {
}

func (c *CompileMock) FindAllSteps(convertedText string) ([]FoundStep, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.FindAllSteps(convertedText)
}

func (c *CompileMock) CombineCodeSlicesToSteps(foundSteps []FoundStep) (map[string]CodeBlock, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.CombineCodeSlicesToSteps(foundSteps)
}

func (c *CompileMock) CreateRootFile(aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.CreateRootFile(aggregatedSteps, sameConfigFile)
}

func (c *CompileMock) WriteStepFiles(compiledDir string, aggregatedSteps map[string]CodeBlock) error {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.WriteStepFiles(compiledDir, aggregatedSteps)
}

func (c *CompileMock) ConvertNotebook(jupytextExecutablePath string, notebookFilePath string) (string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.ConvertNotebook(jupytextExecutablePath, notebookFilePath)
}
