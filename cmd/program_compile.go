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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

func writeRootFile(compiledDir string, rootFileContents string) error {
	file_to_write := path.Join(compiledDir, "root.py")
	logrus.Tracef("File: %v\n", file_to_write)

	err = os.WriteFile(file_to_write, []byte(rootFileContents), 0700)
	if err != nil {
		return fmt.Errorf("Error writing root.py file: %v", err.Error())
	}

	return nil
}

func compileFile(sameConfigFile loaders.SameConfig) (err error) {
	var c = utils.GetCompileFunctions()
	jupytextExecutablePath, notebookFilePath, err := checkExecutableAndFile(sameConfigFile)
	if err != nil {
		return err
	}

	convertedText, err := convertNotebook(jupytextExecutablePath, notebookFilePath)
	if err != nil {
		return err
	}

	stepsFound, codeSlices, err := c.FindAllSteps(convertedText)
	if err != nil {
		return err
	}

	aggregatedSteps, err := c.CombineCodeSlicesToSteps(stepsFound, codeSlices)
	if err != nil {
		return err
	}

	rootFileContents, err := c.CreateRootFile(aggregatedSteps, sameConfigFile)
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

	err = c.WriteStepFiles(compiledDir, aggregatedSteps)
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