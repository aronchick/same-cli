package utils

import "github.com/azure-octo/same-cli/cmd/sameconfig/loaders"

type CompileMock struct {
}

func (c *CompileMock) WriteSupportFiles(workingDirectory string, directoriesToWriteTo []string) error {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.WriteSupportFiles(workingDirectory, directoriesToWriteTo)
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

func (c *CompileMock) CreateRootFile(target string, aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.CreateRootFile(target, aggregatedSteps, sameConfigFile)
}

func (c *CompileMock) WriteStepFiles(target string, compiledDir string, aggregatedSteps map[string]CodeBlock) (map[string]map[string]string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.WriteStepFiles(target, compiledDir, aggregatedSteps)
}

func (c *CompileMock) ConvertNotebook(jupytextExecutablePath string, notebookFilePath string) (string, error) {
	// Placeholder until we mock
	cl := &CompileLive{}
	return cl.ConvertNotebook(jupytextExecutablePath, notebookFilePath)
}
