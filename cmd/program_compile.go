package cmd

/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

var compileProgramCmd = &cobra.Command{
	Use:   "compile",
	Short: "TEMPORARY FUNCTION: Compiles a notebook to a SAME program. Just exploring until we see how this might work.",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(); err != nil {
			return fmt.Errorf("Failed during dependency checks: %v", err)
		}

		// Load config file. Explicit parameters take precedent over config file.
		u := utils.GetUtils(cmd, args)
		sameConfigFilePath, err := u.GetConfigFilePath(filePath)
		if err != nil {
			log.Errorf("could not resolve SAME config file path: %v", err)
			return err
		}

		sameConfigFile, err := loaders.V1{}.LoadSAME(sameConfigFilePath)
		if err != nil {
			log.Errorf("could not load SAME config file: %v", err)
			return err
		}

		if sameConfigFile.Spec.ConfigFilePath == "" {
			sameConfigFile.Spec.ConfigFilePath = filePath
		}

		params, _ := cmd.Flags().GetStringSlice("run-param")

		runParams := make(map[string]string)

		if len(sameConfigFile.Spec.Run.Parameters) > 0 {
			runParams = sameConfigFile.Spec.Run.Parameters
		}

		// override the explicitly set run parameters
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				println(fmt.Sprintf("Invalid param format %q. Expect: key=value", param))
			}
			runParams[parts[0]] = parts[1]
		}

		err = compileFile(*sameConfigFile)
		if err != nil {
			return err
		}
		return nil
	},
}

type CodeBlock struct {
	step_identifier     string
	code                string
	parameters          map[string]string
	packages_to_install []string
}

type CodeBlocks map[string]*CodeBlock

func checkExecutableAndFile(sameConfigFile loaders.SameConfig) (string, string, error) {
	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		return "", "", fmt.Errorf("could not find 'jupytext'. Please run 'python -m pip install jupytext'. You may also need to add it to your path by executing: export PATH=$PATH:$HOME/.local/bin")
	}

	notebookRootDir := filepath.Dir(sameConfigFile.Spec.ConfigFilePath)
	notebookFilePath, err := utils.ResolveLocalFilePath(filepath.Join(notebookRootDir, sameConfigFile.Spec.Pipeline.Package))
	if err != nil {
		return "", "", fmt.Errorf("could not find pipeline definition specified in SAME program: %v", notebookFilePath)
	}

	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return "", "", fmt.Errorf("Could not get cwd: %v", err)
	// }
	return jupytextExecutable, notebookFilePath, nil

}

func convertNotebook(jupytextExecutablePath string, notebookFilePath string) (string, error) {
	log.Infof("Using notebook from here: %v\n", notebookFilePath)
	notebookFile, err := os.Open(notebookFilePath)
	if err != nil {
		return "", fmt.Errorf("error reading from notebook file: %v", notebookFilePath)
	}

	scriptCmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%v --to py", jupytextExecutablePath))
	scriptStdin, err := scriptCmd.StdinPipe()

	if err != nil {
		return "", fmt.Errorf("Error building Stdin pipe for notebook file: %v", err.Error())
	}

	b, _ := ioutil.ReadAll(notebookFile)

	go func() {
		defer scriptStdin.Close()
		_, _ = io.WriteString(scriptStdin, string(b))
	}()

	out, err := scriptCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error executing notebook conversion: %v", err.Error())
	}

	if err != nil {
		return "", fmt.Errorf(`
could not convert the file: %v
full error message: %v`, notebookFilePath, string(out))
	}

	return string(out), nil
}

func getTemporaryCompileDirectory() (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory to compile in: %v", err)
	}

	return dir, nil
}

func findAllSteps(convertedText string) (steps [][]string, code_slices []string, err error) {
	// Need to enable multiline for beginning of the line checking - (?m)
	re_text := "(?m)^# SAME-step-([0-9]+)\\w*$"
	re := regexp.MustCompile(re_text)
	stepsFound := re.FindAllStringSubmatch(convertedText, -1)
	if stepsFound == nil {
		return nil, nil, fmt.Errorf("no steps matched in the file for %v", re_text)
	}

	code_blocks_slices := re.Split(convertedText, -1)
	fmt.Printf("Found %v code blocks in %v steps.\n", len(code_blocks_slices), len(stepsFound))

	return stepsFound, code_blocks_slices, nil
}

func combineCodeSlicesToSteps(stepsFound [][]string, codeSlices []string) (CodeBlocks, error) {
	aggregatedSteps := make(CodeBlocks)
	for i, j := range stepsFound {
		if len(j) > 2 {
			return nil, fmt.Errorf("more than one match in this string, not clear how we got here: %v", j)
		} else if len(j) <= 1 {
			return nil, fmt.Errorf("zero matches in this array, not clear how we got here: %v", j)
		}

		thisStep := j[1]
		logrus.Tracef("Current step: %v\n", thisStep)
		logrus.Tracef("Current slice: %v\n", codeSlices[i])

		if aggregatedSteps[thisStep] == nil {
			aggregatedSteps[thisStep] = &CodeBlock{}
		}

		aggregatedSteps[thisStep].code += codeSlices[i]
		aggregatedSteps[thisStep].step_identifier = thisStep

		import_regex := regexp.MustCompile(`(?mi)^\s*(?:from|import)\s+(\w+(?:\s*,\s*\w+)*)`)
		all_imports := import_regex.FindAllStringSubmatch(aggregatedSteps[thisStep].code, -2)

		logrus.Tracef("Code: %v", aggregatedSteps[thisStep].code)
		if len(all_imports) > 1 {
			logrus.Tracef("Packages:")
			for i := range all_imports {
				aggregatedSteps[thisStep].packages_to_install = append(aggregatedSteps[thisStep].packages_to_install, all_imports[i][1])
				logrus.Tracef("- \t%v\n", all_imports[i][1])
			}

		} else {
			logrus.Tracef("No packages to install for step: %v\n", aggregatedSteps[thisStep].step_identifier)
		}
	}

	return aggregatedSteps, nil
}

func writeSameConfigFile(compiledDir string, sameConfigFile loaders.SameConfig) error {
	sameConfigFileYaml, err := yaml.Marshal(&sameConfigFile.Spec)
	if err != nil {
		return fmt.Errorf("error marshaling same config file: %v", err.Error())
	}
	err = os.WriteFile(path.Join(compiledDir, "same.yaml"), []byte(sameConfigFileYaml), 0700)
	if err != nil {
		return fmt.Errorf("error writing root.py file: %v", err.Error())
	}

	if err != nil {
		return fmt.Errorf("error writing same.yaml file to %v: %v", compiledDir, err)
	}

	return nil
}

func createRootFile(aggregatedSteps CodeBlocks) (string, error) {
	// Create the root file
	rootFile_pre_imports := `
import kfp
import kfp.dsl as dsl
from kfp.components import func_to_container_op, InputPath, OutputPath
import kfp.compiler as compiler
from kfp.dsl.types import Dict as KFPDict, List as KFPList

`
	import_section := ""
	for i := range aggregatedSteps {
		import_section += fmt.Sprintf("import step_%v\n", aggregatedSteps[i].step_identifier)
	}

	root_pre_code := `
@dsl.pipeline(name="Compilation of pipelines",)
def root():
		`
	all_code := ""
	previous_step := ""
	for i := range aggregatedSteps {
		package_string := ""
		if len(aggregatedSteps[i].packages_to_install) > 0 {
			package_string = fmt.Sprintf("\"%v\"", strings.Join(aggregatedSteps[i].packages_to_install[:], "\",\""))
		}

		all_code += fmt.Sprintf(`
	step_%v_op = func_to_container_op(
		func=step_%v.main,
		base_image="python:3.9-slim-buster",
		packages_to_install=[%v],
	)
	step_%v_task = step_%v_op()
		`, aggregatedSteps[i].step_identifier, aggregatedSteps[i].step_identifier, package_string, aggregatedSteps[i].step_identifier, aggregatedSteps[i].step_identifier)
		if previous_step != "" {
			all_code += fmt.Sprintf(`
	step_%v_task.after(step_%v_task)
		`, aggregatedSteps[i].step_identifier, previous_step)
		}
		previous_step = aggregatedSteps[i].step_identifier
	}
	return rootFile_pre_imports + import_section + root_pre_code + all_code, nil

}

func writeRootFile(compiledDir string, rootFileContents string) error {
	file_to_write := path.Join(compiledDir, "root.py")
	logrus.Tracef("File: %v\n", file_to_write)

	err = os.WriteFile(file_to_write, []byte(rootFileContents), 0700)
	if err != nil {
		return fmt.Errorf("Error writing root.py file: %v", err.Error())
	}

	return nil
}

func writeStepFiles(compiledDir string, aggregatedSteps CodeBlocks) error {
	for i := range aggregatedSteps {
		parameter_string := ""
		if len(aggregatedSteps[i].parameters) > 0 {
			for _, key := range aggregatedSteps[i].parameters {
				if parameter_string != "" {
					parameter_string += ","
				}
				parameter_string += key + "=" + aggregatedSteps[i].parameters[key]
			}

		}

		step_to_write := compiledDir + fmt.Sprintf("/step_%v.py", aggregatedSteps[i].step_identifier)
		code_to_write := fmt.Sprintf(`
def main(%v):

`, parameter_string)

		scanner := bufio.NewScanner(strings.NewReader(aggregatedSteps[i].code))
		for scanner.Scan() {
			code_to_write += fmt.Sprintf("\t" + scanner.Text() + "\n")
		}

		err = os.WriteFile(step_to_write, []byte(code_to_write), 0700)
		if err != nil {
			return fmt.Errorf("Error writing step %v: %v", step_to_write, err.Error())
		}
	}

	return nil

}

func compileFile(sameConfigFile loaders.SameConfig) (err error) {
	jupytextExecutablePath, notebookFilePath, err := checkExecutableAndFile(sameConfigFile)
	if err != nil {
		return err
	}

	convertedText, err := convertNotebook(jupytextExecutablePath, notebookFilePath)
	if err != nil {
		return err
	}

	stepsFound, codeSlices, err := findAllSteps(convertedText)
	if err != nil {
		return err
	}

	aggregatedSteps, err := combineCodeSlicesToSteps(stepsFound, codeSlices)
	if err != nil {
		return err
	}

	rootFileContents, err := createRootFile(aggregatedSteps)
	if err != nil {
		return err
	}

	compiledDir, err := getTemporaryCompileDirectory()
	if err != nil {
		return err
	}

	err = writeRootFile(compiledDir, rootFileContents)
	if err != nil {
		return err
	}

	sameConfigFile.Spec.Pipeline.Package = "root.py"
	err = writeSameConfigFile(compiledDir, sameConfigFile)
	if err != nil {
		return nil
	}

	err = writeStepFiles(compiledDir, aggregatedSteps)
	if err != nil {
		return nil
	}

	fmt.Printf("Compilation complete! In order to upload, go to this directory (%v) and execute 'same program run'.\n", compiledDir)
	return nil

}

func init() {
	programCmd.AddCommand(compileProgramCmd)

	compileProgramCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")
}
