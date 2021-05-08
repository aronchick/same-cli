package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type CompileLive struct {
}

func (c *CompileLive) FindAllSteps(convertedText string) (foundSteps []FoundStep, err error) {
	// Need to enable multiline for beginning of the line checking - (?m)
	// Looking for something of the format:
	// # - ...
	// or
	// # + tags=[...]
	re_text := `(?m)^\s*# (?:\+|\-) ?(.*?)$`
	re_steps := regexp.MustCompile(re_text)
	stepsFound := re_steps.FindAllStringSubmatch(convertedText, -1)

	tagsFound := make([][]string, len(stepsFound))
	namedStepsFound := false
	for i, thisStep := range stepsFound {
		tagsFound[i] = ParseTagsForStep(thisStep[1])
		for _, tag := range tagsFound[i] {
			if strings.HasPrefix(tag, "same-step-") {
				namedStepsFound = true
			}
		}
	}

	if !namedStepsFound {
		log.Tracef("no steps found in the file - treating the entire file as a single step.")
		foundStep := FoundStep{}
		foundStep.code_slice = convertedText
		foundStep.index = 0
		foundStep.step_name = "same-step-0"
		foundStep.tags = nil

		return []FoundStep{foundStep}, nil
	}

	log.Trace("Found at least one step with a 'same-step-#' format, breaking up the file")

	code_blocks_slices := re_steps.Split(convertedText, -1)
	foundSteps = make([]FoundStep, 0)
	current_step_name := "same-step-0"
	current_index := 0
	log.Tracef("Raw steps found: %v", len(stepsFound))
	log.Tracef("Code slices found: %v", len(code_blocks_slices))
	log.Tracef("Raw tag blocks found: %v", len(tagsFound))
	for i := range stepsFound {

		if (i == 0) && (code_blocks_slices[0] == "") {
			// When splitting cells, you can often have a zero cell
			// at the start, so skipping it
			code_blocks_slices = code_blocks_slices[1:]
		}

		cacheValue := ""
		genericTags := make([]string, 0)

		// Drop tags into one  of three categories (should be more extensible in the future)
		for _, tag := range tagsFound[i] {
			if strings.HasPrefix(tag, "same-step-") {
				current_step_name = tag
				current_index, _ = strconv.Atoi(strings.Split(tag, "-")[2])
			} else if strings.HasPrefix(tag, "cache=") {
				cacheValue = strings.Split(tag, "=")[1]
			} else {
				genericTags = append(genericTags, tag)
			}
		}
		thisFoundStep := FoundStep{}
		thisFoundStep.step_name = current_step_name
		thisFoundStep.cache_value = cacheValue
		thisFoundStep.tags = genericTags
		thisFoundStep.index = current_index
		thisFoundStep.code_slice = code_blocks_slices[i]
		foundSteps = append(foundSteps, thisFoundStep)

	}

	return foundSteps, nil
}

func ParseTagsForStep(s string) []string {
	re_tags_text := `tags=\[([^\]]*)\]`
	re_tags := regexp.MustCompile(re_tags_text)
	tags_found := re_tags.FindAllStringSubmatch(s, -1)
	log.Tracef(" - Tags found: %v\n", len(tags_found))
	if len(tags_found) > 0 {
		all_tags := strings.Split(tags_found[0][1], ",")
		returned_tags := make([]string, len(all_tags))
		for _, this_tag := range all_tags {
			this_tag = strings.TrimSpace(this_tag)
			if this_tag[0] == '"' {
				this_tag = this_tag[1:]
			}
			if end := len(this_tag) - 1; this_tag[end] == '"' {
				this_tag = this_tag[:end]
			}
			returned_tags = append(returned_tags, this_tag)
		}
		return returned_tags
	}

	return nil

}

func (c *CompileLive) CombineCodeSlicesToSteps(foundSteps []FoundStep) (map[string]CodeBlock, error) {
	aggregatedSteps := make(map[string]CodeBlock)
	for _, foundStep := range foundSteps {

		logrus.Tracef("Current step: %v\n", foundStep.step_name)
		logrus.Tracef("Current slice: %v\n", foundStep.code_slice)

		thisCodeBlock := CodeBlock{}
		if _, exists := aggregatedSteps[foundStep.step_name]; exists {
			thisCodeBlock = aggregatedSteps[foundStep.step_name]
		}

		thisCodeBlock.Code += foundStep.code_slice
		thisCodeBlock.Step_Identifier = foundStep.step_name
		thisCodeBlock.Cache_Value = "P0D"
		if foundStep.cache_value != "" {
			thisCodeBlock.Cache_Value = foundStep.cache_value
		}

		import_regex := regexp.MustCompile(`(?mi)^\s*(?:from|import)\s+(\w+(?:\s*,\s*\w+)*)`)
		all_imports := import_regex.FindAllStringSubmatch(aggregatedSteps[foundStep.step_name].Code, -2)

		log.Tracef("Code: %v", aggregatedSteps[foundStep.step_name].Code)
		if len(all_imports) > 1 {
			log.Tracef("Packages:")
			for i := range all_imports {
				thisCodeBlock.Packages_To_Install = append(aggregatedSteps[foundStep.step_name].Packages_To_Install, all_imports[i][1])
				logrus.Tracef("- \t%v\n", all_imports[i][1])
			}

		} else {
			log.Tracef("No packages to install for step: %v\n", aggregatedSteps[foundStep.step_name].Step_Identifier)
		}
		aggregatedSteps[foundStep.step_name] = thisCodeBlock
	}

	return aggregatedSteps, nil
}

func (c *CompileLive) CreateRootFile(aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
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
		import_section += fmt.Sprintf("import step_%v\n", aggregatedSteps[i].Step_Identifier)
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
		if len(aggregatedSteps[i].Packages_To_Install) > 0 {
			package_string = fmt.Sprintf("\"%v\"", strings.Join(aggregatedSteps[i].Packages_To_Install[:], "\",\""))
		}

		all_code += fmt.Sprintf(`
	step_%v_op = func_to_container_op(
		func=step_%v.main,
		base_image="python:3.9-slim-buster",
		packages_to_install=[%v],
	)
	step_%v_task = step_%v_op()
	step_%v_task.execution_options.caching_strategy.max_cache_staleness = "%v"
		`, aggregatedSteps[i].Step_Identifier, aggregatedSteps[i].Step_Identifier, package_string, aggregatedSteps[i].Step_Identifier, aggregatedSteps[i].Step_Identifier, aggregatedSteps[i].Step_Identifier, aggregatedSteps[i].Cache_Value)
		if previous_step != "" {
			all_code += fmt.Sprintf(`
	step_%v_task.after(step_%v_task)
		`, aggregatedSteps[i].Step_Identifier, previous_step)
		}
		previous_step = aggregatedSteps[i].Step_Identifier
	}
	return rootFile_pre_imports + import_section + root_pre_code + all_code, nil

}

func (c *CompileLive) WriteStepFiles(compiledDir string, aggregatedSteps map[string]CodeBlock) error {
	for i := range aggregatedSteps {
		parameter_string, _ := JoinMapKeysValues(aggregatedSteps[i].Parameters)

		step_to_write := compiledDir + fmt.Sprintf("/step_%v.py", aggregatedSteps[i].Step_Identifier)
		code_to_write := fmt.Sprintf(`
def main(%v):

`, parameter_string)

		scanner := bufio.NewScanner(strings.NewReader(aggregatedSteps[i].Code))
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
