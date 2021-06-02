package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	pongo2 "github.com/flosch/pongo2/v4"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/internal/box"
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
			if strings.HasPrefix(tag, "same_step_") {
				namedStepsFound = true
			}
		}
	}

	if !namedStepsFound {
		log.Tracef("no steps found in the file - treating the entire file as a single step.")
		foundStep := FoundStep{}
		foundStep.CodeSlice = convertedText
		foundStep.Index = 0
		foundStep.StepName = "same_step_0"
		foundStep.Tags = nil

		return []FoundStep{foundStep}, nil
	}

	log.Trace("Found at least one step with a 'same_step_#' format, breaking up the file")

	code_blocks_slices := re_steps.Split(convertedText, -1)
	foundSteps = make([]FoundStep, 0)
	current_step_name := "same_step_0"
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
			if strings.HasPrefix(tag, "same_step_") {
				current_step_name = tag
				current_index, _ = strconv.Atoi(strings.Split(tag, "_")[2])
			} else if strings.HasPrefix(tag, "cache=") {
				cacheValue = strings.Split(tag, "=")[1]
			} else {
				genericTags = append(genericTags, tag)
			}
		}
		thisFoundStep := FoundStep{}
		thisFoundStep.StepName = current_step_name
		thisFoundStep.CacheValue = cacheValue
		thisFoundStep.Tags = genericTags
		thisFoundStep.Index = current_index
		thisFoundStep.CodeSlice = code_blocks_slices[i]
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
	for i, foundStep := range foundSteps {

		log.Tracef("Current step: %v\n", foundStep.StepName)
		log.Tracef("Current slice: %v\n", foundStep.CodeSlice)

		thisCodeBlock := CodeBlock{}
		if _, exists := aggregatedSteps[foundStep.StepName]; exists {
			thisCodeBlock = aggregatedSteps[foundStep.StepName]
		}

		thisCodeBlock.Code += foundStep.CodeSlice
		thisCodeBlock.StepIdentifier = foundStep.StepName
		thisCodeBlock.CacheValue = "P0D"
		if foundStep.CacheValue != "" {
			thisCodeBlock.CacheValue = foundStep.CacheValue
		}

		import_regex := regexp.MustCompile(`(?mi)^\s*(?:from|import)\s+(\w+(?:\s*,\s*\w+)*)`)
		all_imports := import_regex.FindAllStringSubmatch(thisCodeBlock.Code, -2)

		log.Tracef("Code: %v", aggregatedSteps[foundStep.StepName].Code)
		if len(all_imports) >= 1 {
			log.Tracef("Packages:")
			if thisCodeBlock.PackagesToInstall == nil {
				thisCodeBlock.PackagesToInstall = make(map[string]string)
			}
			for i := range all_imports {
				// TODO: Parse for versions eventually
				thisCodeBlock.PackagesToInstall[all_imports[i][1]] = ""
				log.Tracef("- \t%v\n", all_imports[i][1])
			}

		} else {
			log.Tracef("No packages to install for found step #: %v\n", i)
		}
		aggregatedSteps[foundStep.StepName] = thisCodeBlock
	}

	return aggregatedSteps, nil
}

func (c *CompileLive) CreateRootFile(target string, aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {

	if !ContainsString([]string{"kubeflow", "aml"}, target) {
		return "", fmt.Errorf("unknown compilation target: %v", target)
	}

	rootParameterString := ""

	if len(sameConfigFile.Spec.Run.Parameters) > 0 {
		rootParameters := make(map[string]string, len(sameConfigFile.Spec.Run.Parameters))
		for k, untyped_v := range sameConfigFile.Spec.Run.Parameters {
			switch untyped_v.(type) {
			case int, int8, uint8, int16, uint16, int32, uint32, int64, uint64, uint, uintptr, float32, float64, bool, string:
				rootParameters[k] = fmt.Sprintf("%v", untyped_v)
			default:
				log.Warnf("We only support numeric, bool and strings as default parameters (no dicts or lists). We're setting the default value for '%v' to ''.", k)
				rootParameters[k] = ""
			}

		}
		rootParameterString, _ = JoinMapKeysValues(rootParameters)
	}

	previous_step := ""
	steps_left_to_parse_map := make(map[string]string)
	allSteps := map[string]map[string]string{}

	// Copying this to a new variable so that we can delete them
	for _, thisCodeBlock := range aggregatedSteps {
		steps_left_to_parse_map[thisCodeBlock.StepIdentifier] = thisCodeBlock.StepIdentifier
	}

	steps_to_parse := make([]string, 0, len(steps_left_to_parse_map))
	for key := range steps_left_to_parse_map {
		steps_to_parse = append(steps_to_parse, key)
	}
	sort.Strings(steps_to_parse)

	// Unfortunately, every early step's package includes also need to be included in later
	// steps. This is become some objects (like IPython.image) require module imports.
	// There's probably a more elegant way to handle this.
	globalPackagesSlice := make(map[string]string)
	globalPackagesString := ""
	for i := 0; i < len(steps_to_parse); i++ {
		thisCodeBlock := CodeBlock{}
		thisCodeBlock.StepIdentifier = steps_to_parse[i]

		// Another hacky work-around - we're just building every package into every container
		// ...definitely should be more efficient (only build in what we need per container)
		packageString := ""
		for k := range thisCodeBlock.PackagesToInstall {
			globalPackagesSlice[k] = ""
		}

		for k := range globalPackagesSlice {
			packageString += fmt.Sprintf("'%v',", k)
			globalPackagesString += fmt.Sprintf("'%v',", k)
		}

		allSteps[thisCodeBlock.StepIdentifier] = map[string]string{
			"Name":          thisCodeBlock.StepIdentifier,
			"PackageString": packageString,
			"CacheValue":    thisCodeBlock.CacheValue,
			"PreviousStep":  previous_step,
		}

		previous_step = thisCodeBlock.StepIdentifier

	}

	experimentName := removeIllegalExperimentNameCharacters(sameConfigFile.Spec.Metadata.Name)
	step_string := ""
	for _, step := range steps_to_parse {
		if step_string != "" {
			step_string += ", "
		}
		step_string += fmt.Sprintf("%v_step", step)
	}

	rootFileContext := pongo2.Context{
		"RootParameterString":  rootParameterString,
		"GlobalPackagesString": globalPackagesString,
		"Steps":                allSteps,
		"StepString":           step_string,
		"ExperimentName":       experimentName,
	}

	var root_file_bytes []byte
	switch target {
	case "kubeflow":
		root_file_bytes = box.Get("/kfp/root.tmpl")
	case "aml":
		root_file_bytes = box.Get("/aml/root.tmpl")
	default:
		return "", fmt.Errorf("unknown compilation target: %v", target)
	}

	tmpl := pongo2.Must(pongo2.FromBytes(root_file_bytes))

	rootFileString, err := tmpl.Execute(rootFileContext)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %v", err)
	}

	return rootFileString, nil

}

func (c *CompileLive) WriteStepFiles(target string, compiledDir string, aggregatedSteps map[string]CodeBlock) error {
	for i := range aggregatedSteps {
		parameter_string, _ := JoinMapKeysValues(aggregatedSteps[i].Parameters)
		if parameter_string != "" {
			parameter_string = "," + parameter_string
		}

		// Prepend an empty locals as the default
		parameter_string = "__context='gAR9lC4=', __run_info='gAR9lC4=', __metadata_url=''" + parameter_string

		step_to_write := ""
		if target == "kubeflow" {
			step_to_write = filepath.Join(compiledDir, fmt.Sprintf("%v.py", aggregatedSteps[i].StepIdentifier))
		} else if target == "aml" {
			// AML requires each step to be in its own directory, with the same name as the python file
			stepDirectoryName := filepath.Join(compiledDir, aggregatedSteps[i].StepIdentifier)
			_, err := os.Stat(stepDirectoryName)
			if os.IsNotExist(err) {
				errDir := os.MkdirAll(stepDirectoryName, 0700)
				if errDir != nil {
					return fmt.Errorf("error creating step directory for %v: %v", stepDirectoryName, err)
				}

			}

			step_to_write = filepath.Join(stepDirectoryName, fmt.Sprintf("%v.py", aggregatedSteps[i].StepIdentifier))
		} else {
			return fmt.Errorf("unknown target: %v", target)
		}

		code_to_write := fmt.Sprintf(`
import argparse as __argparse
from multiprocessing import context
import pathlib
from typing import NamedTuple
from azureml.core import Run
from pprint import pprint as __pp
import os
from pathlib import Path as __Path
from azureml.pipeline.core import (
	PipelineData as __PipelineData,
	PipelineParameter as __PipelineParameter,
)
import dill
from base64 import (
	urlsafe_b64encode as __urlsafe_b64encode,
	urlsafe_b64decode as __urlsafe_b64decode,
)

def main(%v) -> NamedTuple('FuncOutput',[('context', str),]):
	import dill
	import base64
	from base64 import urlsafe_b64encode, urlsafe_b64decode
	from copy import copy as __copy
	from types import ModuleType as __ModuleType
	from pprint import pprint as __pp
	import datetime as __datetime
	import requests

	__run_info_dict = dill.loads(urlsafe_b64decode(__run_info))
	__base64_decode = urlsafe_b64decode(__context)
	__context_import_dict = dill.loads(__base64_decode)

	__variables_to_mount = {}
	__loc = {}

	for __k in __context_import_dict:
		__variables_to_mount[__k] = dill.loads(__context_import_dict[__k])

	__json_data = {
		"experiment_id": __run_info_dict["experiment_id"],
		"run_id": __run_info_dict["run_id"],
		"step_id": "%v",
		"metadata_type": "input",
		"metadata_value": __context,
		"metadata_time": __datetime.datetime.now().isoformat(),
	}

	print(f"Metadata url: {__metadata_url}")
	if __metadata_url != '':
		print("Found metadata URL - executing.")
		__pp(__json_data)
		try:
			__r = requests.post(__metadata_url, json=__json_data,)	
			__r.raise_for_status()
		except requests.exceptions.HTTPError as __err:
			print(f"Error: {__err}")

`, parameter_string, aggregatedSteps[i].StepIdentifier)

		scanner := bufio.NewScanner(strings.NewReader(aggregatedSteps[i].Code))
		inner_code_to_execute := `
import dill
import base64
from base64 import urlsafe_b64encode, urlsafe_b64decode
from types import ModuleType as __ModuleType

`
		for scanner.Scan() {
			inner_code_to_execute += fmt.Sprintln(scanner.Text())
		}
		inner_code_to_execute += `
__locals_keys = frozenset(locals().keys())
__globals_keys = frozenset(globals().keys())
__context_export = {}

for val in __globals_keys:
	if not val.startswith("_") and not isinstance(val, __ModuleType):
		__context_export[val] = dill.dumps(globals()[val])

# Locals needs to come after globals in case we made changes
for val in __locals_keys:
	if not val.startswith("_") and not isinstance(val, __ModuleType):
		__context_export[val] = dill.dumps(locals()[val])

__b64_string = str(urlsafe_b64encode(dill.dumps(__context_export)), encoding="ascii")

`

		code_to_write += fmt.Sprintf("\t__inner_code_to_execute = '''%v'''\n", inner_code_to_execute)
		code_to_write += fmt.Sprintf(`
	exec(__inner_code_to_execute, __variables_to_mount, __loc)

	__json_output_data = {
		"experiment_id": __run_info_dict["experiment_id"],
		"run_id": __run_info_dict["run_id"],
		"step_id": "%v",
		"metadata_type": "output",
		"metadata_value": __loc["__b64_string"],
		"metadata_time": __datetime.datetime.now().isoformat(),
	}

	print(f"Metadata url: {__metadata_url}")
	if __metadata_url != '':
		print("Found metadata URL - executing.")
		__pp(__json_data)
		try:
			__r = requests.post(__metadata_url, json=__json_output_data,)	
			__r.raise_for_status()
		except requests.exceptions.HTTPError as err:
			print(f"Error: {err}")

	from collections import namedtuple
	output = namedtuple("FuncOutput", ["context"])
	return output(__loc["__b64_string"])
`, aggregatedSteps[i].StepIdentifier)

		code_to_write += `
if __name__ == "__main__":
	__run = Run.get_context()
	__parser = __argparse.ArgumentParser("cleanse")
	__parser.add_argument("--input_context", type=str, help="Context to run as string")
	__parser.add_argument("--run_info", type=str, help="Run info")
	__parser.add_argument("--output_context_path", type=str, help="Output context path")
	__parser.add_argument("--metadata_url", type=str, help="Metadata URL")

	__args = __parser.parse_args()

	__input_context_string = "gAR9lC4="
	__context_filename = "context.txt"
	if "__pipelinedata_context" in __args.input_context:
		context_full_path = __Path(__args.input_context) / __context_filename
		print(f"reading file: {context_full_path}")
		__input_context_string = context_full_path.read_text()
	elif __args.input_context and __args.input_context.strip():
		__input_context_string = __args.input_context.strip()

	# Need to unpack and do this here, because AML only gives
	# us the run id inside the container. Unpacking and repacking so
	# bulk of the code is unchanged.
	__run_info_dict = dill.loads(__urlsafe_b64decode(__args.run_info))
	__run_info_dict["run_id"] = __run.get_details()["runId"]

	# Returns a tuple, where the zeroth index is the string
	__output_context_tuple = main(
		__context=__input_context_string,
		__run_info=str(
			__urlsafe_b64encode(dill.dumps(__run_info_dict)), encoding="ascii"
		),
		__metadata_url=__args.metadata_url,
	)

	__p = __Path(__args.output_context_path)
	__p.mkdir(parents=True, exist_ok=True)
	__filepath = __p / __context_filename
	with __filepath.open("w+") as __f:
		__f.write(__output_context_tuple[0])
`

		err := os.WriteFile(step_to_write, []byte(code_to_write), 0400)
		if err != nil {
			return fmt.Errorf("Error writing step %v: %v", step_to_write, err.Error())
		}
	}

	return nil

}

func (c *CompileLive) ConvertNotebook(jupytextExecutablePath string, notebookFilePath string) (string, error) {
	log.Infof("Using notebook from here: %v\n", notebookFilePath)
	notebookFile, err := os.Open(notebookFilePath)
	if err != nil {
		return "", fmt.Errorf("program_compile.go: error reading from notebook file: %v", notebookFilePath)
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

func removeIllegalExperimentNameCharacters(s string) string {
	return strings.Map(
		func(r rune) rune {
			if r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		},
		s,
	)
}
