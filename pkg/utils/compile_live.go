package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/sirupsen/logrus"
)

type CompileLive struct {
}

func (c *CompileLive) FindAllSteps(convertedText string) (steps [][]string, code_slices []string, err error) {
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

func (c *CompileLive) CombineCodeSlicesToSteps(stepsFound [][]string, codeSlices []string) (CodeBlocks, error) {
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

func (c *CompileLive) CreateRootFile(aggregatedSteps CodeBlocks, sameConfigFile loaders.SameConfig) (string, error) {
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

	rootParameterString := ""
	if len(sameConfigFile.Spec.Run.Parameters) > 0 {
		rootParameters := make(map[string]string, len(sameConfigFile.Spec.Run.Parameters))
		for k, v := range sameConfigFile.Spec.Run.Parameters {
			rootParameters[k] = v
		}
		rootParameterString, _ = JoinMapKeysValues(rootParameters)
	}

	root_pre_code := fmt.Sprintf(`
@dsl.pipeline(name="Compilation of pipelines",)
def root(%v):
		`, rootParameterString)
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

func (c *CompileLive) WriteStepFiles(compiledDir string, aggregatedSteps CodeBlocks) error {
	for i := range aggregatedSteps {
		parameter_string, _ := JoinMapKeysValues(aggregatedSteps[i].parameters)

		step_to_write := compiledDir + fmt.Sprintf("/step_%v.py", aggregatedSteps[i].step_identifier)
		code_to_write := fmt.Sprintf(`
def main(%v):

`, parameter_string)

		scanner := bufio.NewScanner(strings.NewReader(aggregatedSteps[i].code))
		for scanner.Scan() {
			code_to_write += fmt.Sprintf("\t" + scanner.Text() + "\n")
		}

		err := os.WriteFile(step_to_write, []byte(code_to_write), 0700)
		if err != nil {
			return fmt.Errorf("Error writing step %v: %v", step_to_write, err.Error())
		}
	}

	return nil

}
