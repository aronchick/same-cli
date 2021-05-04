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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

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
	packages_to_install []string
}

type CodeBlocks map[string]*CodeBlock

func compileFile(sameConfigFile loaders.SameConfig) (err error) {
	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		return fmt.Errorf("could not find 'nbconvert'. Please run 'python -m pip install nbconvert'. You may also need to add it to your path by executing: export PATH=$PATH:$HOME/.local/bin")
	}

	notebookRootDir := filepath.Dir(sameConfigFile.Spec.ConfigFilePath)
	notebookFilePath, err := utils.ResolveLocalFilePath(filepath.Join(notebookRootDir, sameConfigFile.Spec.Pipeline.Package))
	if err != nil {
		return fmt.Errorf("could not find pipeline definition specified in SAME program: %v", notebookFilePath)
	}

	fmt.Printf("Current filepath: %v\n", notebookFilePath)
	notebookFile, err := os.Open(notebookFilePath)
	if err != nil {
		return fmt.Errorf("error reading from notebook file: %v", notebookFilePath)
	}

	scriptCmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%v --to py", jupytextExecutable))
	scriptStdin, err := scriptCmd.StdinPipe()

	if err != nil {
		return fmt.Errorf("Error building Stdin pipe for notebook file: %v", err.Error())
	}

	b, _ := ioutil.ReadAll(notebookFile)

	go func() {
		defer scriptStdin.Close()
		_, _ = io.WriteString(scriptStdin, string(b))
	}()

	out, err := scriptCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error executing notebook conversion: %v", err.Error())
	}

	if err != nil {
		err = fmt.Errorf(`
could not convert the file: %v
full error message: %v`, notebookFilePath, string(out))
		return err
	}

	convertedText := string(out)

	// dir, err := ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	// if err != nil {
	// 	log.Warn(err)
	// 	return
	// }

	// fmt.Printf("TempDir: %v\n", dir)

	// Need to enable multiline for beginning of the line checking - (?m)
	re_text := "(?m)^# SAME-step-([0-9]+)\\w*$"
	re := regexp.MustCompile(re_text)

	all_named_steps := re.FindAllStringSubmatch(convertedText, -1)
	if all_named_steps == nil {
		return fmt.Errorf("no steps matched in the entire file for %v", re_text)
	}

	// fmt.Printf("Steps: %v", all_named_steps)

	code_blocks_slices := re.Split(convertedText, -1)
	// fmt.Printf("Code block slices: %v", len(code_blocks_slices))
	code_blocks := make(CodeBlocks)

	logrus.Infof("Found %v steps to combine.", len(all_named_steps))

	for i, j := range all_named_steps {
		if len(j) > 2 {
			return fmt.Errorf("more than one match in this string, not clear how we got here: %v", j)
		} else if len(j) <= 1 {
			return fmt.Errorf("zero matches in this array, not clear how we got here: %v", j)
		}

		logrus.Tracef("Current step: %v\n", j[1])
		logrus.Tracef("Current slice: %v\n", code_blocks_slices[i])
		if code_blocks[j[1]] == nil {
			code_blocks[j[1]] = &CodeBlock{}
		}
		code_blocks[j[1]].code += code_blocks_slices[i]
		code_blocks[j[1]].step_identifier = j[1]

		import_regex := regexp.MustCompile(`(?mi)^\s*(?:from|import)\s+(\w+(?:\s*,\s*\w+)*)`)
		all_imports := import_regex.FindAllStringSubmatch(code_blocks[j[1]].code, -2)

		// fmt.Printf("Code: %v", code_blocks_slices[i])
		if len(all_imports) > 1 {
			// fmt.Printf("Code: %v", code_blocks_slices[i])
			// fmt.Printf("Match: %v", all_imports[1])
			logrus.Tracef("Packages:")
			for i := range all_imports {
				code_blocks[j[1]].packages_to_install = append(code_blocks[j[1]].packages_to_install, all_imports[i][1])
				logrus.Tracef("- \t%v\n", all_imports[i][1])
			}

		} else {
			logrus.Tracef("No Matches\n")
		}
	}

	// Create the root file
	rootFile_pre_imports := `
import kfp
import kfp.dsl as dsl
from kfp.components import func_to_container_op, InputPath, OutputPath
import kfp.compiler as compiler
from kfp.dsl.types import Dict as KFPDict, List as KFPList
`
	import_section := ""
	for i := range code_blocks {
		import_section += fmt.Sprintf("import step_%v\n", code_blocks[i].step_identifier)
	}

	root_pre_code := `
@dsl.pipeline(
name="Compilation of pipelines",
)
def root():
`
	all_code := ""
	previous_step := ""
	for i := range code_blocks {
		package_string := ""
		if len(code_blocks[i].packages_to_install) > 0 {
			package_string = fmt.Sprintf("\"%v\"", strings.Join(code_blocks[i].packages_to_install[:], "\",\""))
		}

		all_code += fmt.Sprintf(`
	step_%v_op = func_to_container_op(
		func=step_%v.step_%v,
		base_image="python:3.9-slim-buster",
		packages_to_install=[%v
		],
	)
	step_%v_task = step_%v_op()
`, code_blocks[i].step_identifier, code_blocks[i].step_identifier, code_blocks[i].step_identifier, package_string, code_blocks[i].step_identifier, code_blocks[i].step_identifier)
		if previous_step != "" {
			all_code += fmt.Sprintf(`
	step_%v_task.after(step_%v_task)
`, code_blocks[i].step_identifier, previous_step)
		}
		previous_step = code_blocks[i].step_identifier
	}

	compiledDir := "/tmp/working_same/compiled"
	_ = RemoveContents(compiledDir)

	_ = Copy("/tmp/working_same/same.yaml", compiledDir+"/same.yaml")

	file_to_write := compiledDir + "/root.py"
	logrus.Tracef("File: %v\n", file_to_write)

	err = os.WriteFile(file_to_write, []byte(rootFile_pre_imports+import_section+root_pre_code+all_code), 0700)
	if err != nil {
		return fmt.Errorf("Error writing root.py file: %v", err.Error())
	}

	for i := range code_blocks {
		step_to_write := compiledDir + fmt.Sprintf("/step_%v.py", code_blocks[i].step_identifier)
		code_to_write := fmt.Sprintf(`
def step_%v():

`, code_blocks[i].step_identifier)

		scanner := bufio.NewScanner(strings.NewReader(code_blocks[i].code))
		for scanner.Scan() {
			code_to_write += fmt.Sprintf("\t" + scanner.Text() + "\n")
		}

		err = os.WriteFile(step_to_write, []byte(code_to_write), 0700)
		if err != nil {
			return fmt.Errorf("Error writing step %v: %v", step_to_write, err.Error())
		}
	}

	fmt.Printf("Compilation complete! In order to upload, go to this directory (%v) and execute 'same program run'. Make sure your same.yaml is pointing at root.py\n", compiledDir)
	return nil

}

func init() {
	programCmd.AddCommand(compileProgramCmd)

	compileProgramCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")

	compileProgramCmd.Flags().StringSliceP("run-param", "p", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
