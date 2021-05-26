package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
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
		foundStep.code_slice = convertedText
		foundStep.index = 0
		foundStep.step_name = "same_step_0"
		foundStep.tags = nil

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
	for i, foundStep := range foundSteps {

		log.Tracef("Current step: %v\n", foundStep.step_name)
		log.Tracef("Current slice: %v\n", foundStep.code_slice)

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
		all_imports := import_regex.FindAllStringSubmatch(thisCodeBlock.Code, -2)

		log.Tracef("Code: %v", aggregatedSteps[foundStep.step_name].Code)
		if len(all_imports) >= 1 {
			log.Tracef("Packages:")
			if thisCodeBlock.Packages_To_Install == nil {
				thisCodeBlock.Packages_To_Install = make(map[string]string)
			}
			for i := range all_imports {
				// TODO: Parse for versions eventually
				thisCodeBlock.Packages_To_Install[all_imports[i][1]] = ""
				log.Tracef("- \t%v\n", all_imports[i][1])
			}

		} else {
			log.Tracef("No packages to install for found step #: %v\n", i)
		}
		aggregatedSteps[foundStep.step_name] = thisCodeBlock
	}

	return aggregatedSteps, nil
}

func (c *CompileLive) CreateRootFile(target string, aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
	switch target {
	case "kubeflow":
		return compileRootFileKubeflow(aggregatedSteps, sameConfigFile)
	case "aml":
		return compileRootFileAML(aggregatedSteps, sameConfigFile)
	default:
		return "", fmt.Errorf("unknown compilation target: %v", target)
	}
}

func compileRootFileKubeflow(aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
	// Create the root file
	rootFile_pre_imports := `
import kfp
import kfp.dsl as dsl
from kfp.components import func_to_container_op, InputPath, OutputPath
import kfp.compiler as compiler
from kfp.dsl.types import Dict as KFPDict, List as KFPList
from typing import NamedTuple

`
	import_section := ""
	for i := range aggregatedSteps {
		import_section += fmt.Sprintf("import %v\n", aggregatedSteps[i].Step_Identifier)
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

	run_info_component := `
def get_run_info(
	run_id: str,
	mlflow_run_id: str = '',
	mlflow_tracking_url: str = '',
) -> NamedTuple("RunInfoOutput", [("run_info", str),]):
	"""Example of getting run info for current pipeline run"""
	import kfp
	import json
	import dill
	import base64
	import datetime
	from dateutil.tz import tzlocal
	from pprint import pprint as pp

	print(f"Current run ID is {run_id}.")
	client = kfp.Client(host="http://ml-pipeline:8888")
	run_info = client.get_run(run_id=run_id)
	# Hide verbose info
	run_info.run.pipeline_spec.workflow_manifest = None

	from collections import namedtuple

	pp(run_info.run)

	run_info_dict = {
		"run_id": run_info.run.id,
		"name": run_info.run.name,
		"created_at": run_info.run.created_at.isoformat(),
		"pipeline_id": run_info.run.pipeline_spec.pipeline_id,
	}
	for r in run_info.run.resource_references:
		run_info_dict[r.key.type.lower()] = r.key.id

	output = namedtuple("RunInfoOutput", ["run_info"])
	return output(
		str(base64.urlsafe_b64encode(dill.dumps(run_info_dict)), encoding="ascii")
	)

get_run_info_component = kfp.components.create_component_from_func(
	func=get_run_info,
	packages_to_install=[
		"kfp",
		"dill",
	],
)
`

	root_pre_code := fmt.Sprintf(`
@dsl.pipeline(name="Compilation of pipelines",)
def root(%v, context='', metadata_url=''):
	# The below is base64 encoding of an empty locals() output
	if context == '':
		_original_context = "gAR9lC4="
	else:
		_original_context = context

	'''kfp.dsl.RUN_ID_PLACEHOOLDER inside a parameter will be populated with KFP Run ID at runtime.'''
	run_info_op = get_run_info_component(run_id=kfp.dsl.RUN_ID_PLACEHOLDER)

		`, rootParameterString)
	all_code := ""
	previous_step := ""
	context_variable := ""
	number_of_raw_steps := len(aggregatedSteps)
	steps_left_to_parse := make(map[string]string)

	for _, thisCodeBlock := range aggregatedSteps {
		steps_left_to_parse[thisCodeBlock.Step_Identifier] = thisCodeBlock.Step_Identifier
	}

	// Unfortunately, every early step's package includes also need to be included in later
	// steps. This is become some objects (like IPython.image) require module imports.
	// There's probably a more elegant way to handle this.
	packages_to_install_global := make(map[string]string)
	packages_to_install_global["dill"] = ""
	for i := 0; i < number_of_raw_steps; i++ {
		thisCodeBlock := CodeBlock{}
		for _, test_step_identifier := range steps_left_to_parse {
			if thisCodeBlock.Step_Identifier == "" || test_step_identifier <= thisCodeBlock.Step_Identifier {
				thisCodeBlock = aggregatedSteps[test_step_identifier]
			}
		}

		if thisCodeBlock.Step_Identifier == "" {
			return "", fmt.Errorf("compile_live.go: got to the end of searching and did not assign a code block. Not sure how.")
		}
		delete(steps_left_to_parse, thisCodeBlock.Step_Identifier)

		package_string := ""
		for k := range thisCodeBlock.Packages_To_Install {
			packages_to_install_global[k] = ""
		}

		for k := range packages_to_install_global {
			package_string += fmt.Sprintf("'%v',", k)
		}

		context_variable = fmt.Sprintf("%v_task.outputs['context']", previous_step)
		if previous_step == "" {
			context_variable = "_original_context"
		}

		all_code += fmt.Sprintf(`
	%v_op = func_to_container_op(
		func=%v.main,
		base_image="python:3.9-slim-buster",
		packages_to_install=[%v],
	)
	%v_task = %v_op(context=%v, run_info=run_info_op.outputs["run_info"], metadata_url=metadata_url)
	%v_task.execution_options.caching_strategy.max_cache_staleness = "%v"
		`,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			package_string,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			context_variable,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Cache_Value)

		if previous_step != "" {
			all_code += fmt.Sprintf(`
	%v_task.after(%v_task)
		`,
				thisCodeBlock.Step_Identifier,
				previous_step)
		}

		previous_step = thisCodeBlock.Step_Identifier
	}
	return rootFile_pre_imports + import_section + run_info_component + root_pre_code + all_code, nil

}

func compileRootFileAML(aggregatedSteps map[string]CodeBlock, sameConfigFile loaders.SameConfig) (string, error) {
	// Create the root file
	rootFile_pre_imports := `
from typing import NamedTuple
import azureml.core

# import dotenv
import dill
import base64

import os
from azureml.core import Workspace
from azureml.core.authentication import ServicePrincipalAuthentication
from azureml.core.compute import ComputeTarget, AmlCompute
from azureml.core.runconfig import RunConfiguration
from azureml.core.conda_dependencies import CondaDependencies
from azureml.core import Environment
from azureml.pipeline.core import Pipeline, PipelineData, PipelineParameter
from azureml.pipeline.steps import PythonScriptStep
from azureml.core import Run, Experiment, Datastore
`
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

	get_aml_workspace := `
def get_aml_workspace(aml_workspace_credentials):
    svc_pr_password = aml_workspace_credentials.get("AML_SP_PASSWORD_VALUE")

    svc_pr = ServicePrincipalAuthentication(
        tenant_id=aml_workspace_credentials.get("AML_SP_TENANT_ID"),
        service_principal_id=aml_workspace_credentials.get("AML_SP_APP_ID"),
        service_principal_password=svc_pr_password,
    )

    return Workspace(
        subscription_id=aml_workspace_credentials.get("WORKSPACE_SUBSCRIPTION_ID"),
        resource_group=aml_workspace_credentials.get("WORKSPACE_RESOURCE_GROUP"),
        workspace_name=aml_workspace_credentials.get("WORKSPACE_NAME"),
        auth=svc_pr,
    )

`

	root_pre_code := fmt.Sprintf(`
def root(
    %v,
    context="",
    metadata_url="",
    aml_workspace_credentials={},
):
	# The below is base64 encoding of an empty locals() output
	if context == '':
		_original_context = "gAR9lC4="
	else:
		_original_context = context


    expected_fields = [
        "AML_SP_PASSWORD_VALUE",
        "AML_SP_TENANT_ID",
        "AML_SP_APP_ID",
        "WORKSPACE_SUBSCRIPTION_ID",
        "WORKSPACE_RESOURCE_GROUP",
        "WORKSPACE_NAME",
        "AML_COMPUTE_NAME",
    ]

    missing_fields = [
        field
        for field in expected_fields
        if not aml_workspace_credentials.get(field, None)
    ]
    if len(missing_fields) > 0:
        raise ValueError(
            f"Missing expected fields in credential dictionary: {','.join(missing_fields)}"
        )

    ws = get_aml_workspace(aml_workspace_credentials)
    experiment = Experiment(ws, "%v")

    run_info_dict = {
        "experiment_id": experiment.id,
        "step_id": "same_step_0",
    }

    output = {}
    output["run_info"] = str(
        base64.urlsafe_b64encode(dill.dumps(run_info_dict)), encoding="ascii"
    )

		`, rootParameterString, sameConfigFile.GetName())

	provision_aml_compute := `
compute_name = aml_workspace_credentials.get("AML_COMPUTE_NAME")
vm_size = "STANDARD_NC6"
if compute_name in ws.compute_targets:
	compute_target = ws.compute_targets[compute_name]
	if compute_target and type(compute_target) is AmlCompute:
		print("Found compute target: " + compute_name)
else:
	print("Creating a new compute target...")
	provisioning_config = AmlCompute.provisioning_configuration(
		vm_size=vm_size, min_nodes=0, max_nodes=4  # STANDARD_NC6 is GPU-enabled
	)
	# create the compute target
	compute_target = ComputeTarget.create(ws, compute_name, provisioning_config)

	# Can poll for a minimum number of nodes and for a specific timeout.
	# If no min node count is provided it will use the scale settings for the cluster
	compute_target.wait_for_completion(
		show_output=True, min_node_count=None, timeout_in_minutes=20
	)

	# For a more detailed view of current cluster status, use the 'status' property
	print(compute_target.status.serialize())
`
	all_code := ""
	previous_step := ""
	context_variable := ""
	number_of_raw_steps := len(aggregatedSteps)
	all_steps := make([]string, 0)
	steps_left_to_parse := make(map[string]string)

	for _, thisCodeBlock := range aggregatedSteps {
		steps_left_to_parse[thisCodeBlock.Step_Identifier] = thisCodeBlock.Step_Identifier
	}

	// Unfortunately, every early step's package includes also need to be included in later
	// steps. This is become some objects (like IPython.image) require module imports.
	// There's probably a more elegant way to handle this.
	packages_to_install_global := make(map[string]string)
	packages_to_install_global["dill"] = ""
	for i := 0; i < number_of_raw_steps; i++ {
		thisCodeBlock := CodeBlock{}
		for _, test_step_identifier := range steps_left_to_parse {
			if thisCodeBlock.Step_Identifier == "" || test_step_identifier <= thisCodeBlock.Step_Identifier {
				thisCodeBlock = aggregatedSteps[test_step_identifier]
			}
		}

		if thisCodeBlock.Step_Identifier == "" {
			return "", fmt.Errorf("compile_live.go: got to the end of searching and did not assign a code block. Not sure how.")
		}

		all_steps = AppendIfMissing(all_steps, thisCodeBlock.Step_Identifier)
		delete(steps_left_to_parse, thisCodeBlock.Step_Identifier)

		for k := range thisCodeBlock.Packages_To_Install {
			packages_to_install_global[k] = ""
		}

		context_variable = fmt.Sprintf("__pipelinedata_context_%v", previous_step)
		if previous_step == "" {
			context_variable = "__original_context_param"
		}

		all_code += fmt.Sprintf(`
entry_point = "%v.py"
__original_context_param = PipelineParameter(
	name="input_string", default_value=__original_context
)
__original_context_param = PipelineParameter(
	name="input_string", default_value=__original_context
)
__pipelinedata_context_%v = PipelineData(
	"__pipelinedata_context_%v", output_mode="mount"
)

%v_step = PythonScriptStep(
	source_directory="%v",
	script_name=entry_point,
	arguments=[
		"--input_context",
		%v,
		"--run_info",
		output["run_info"],
		"--metadata_url",
		metadata_url,
		"--output_context",
		__pipelinedata_context_%v,
	],
	outputs=[__pipelinedata_context_%v],
	compute_target=compute_target,
	runconfig=aml_run_config,
	allow_reuse=False,
    )
`,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier,
			context_variable,
			thisCodeBlock.Step_Identifier,
			thisCodeBlock.Step_Identifier)

		previous_step = thisCodeBlock.Step_Identifier
	}

	step_string := fmt.Sprintf(`"%v"
`, strings.Join(all_steps, `","`))

	package_string := ""
	for k := range packages_to_install_global {
		package_string += fmt.Sprintf("'%v',", k)
	}
	global_packages_to_install := fmt.Sprintf(`
aml_run_config = RunConfiguration()
aml_run_config.target = compute_target
aml_run_config.environment = Environment(name="AML_COMPUTE_ENVIRONMENT")
conda_dep = CondaDependencies()

all_packages = [%v]
for package in all_packages:
	conda_dep.add_pip_package(package)

# Adds dependencies to PythonSection of myenv
aml_run_config.environment.python.conda_dependencies = conda_dep
aml_run_config.environment.python.conda_dependencies.set_python_version("3.8")
`, package_string)

	pipeline_submit := fmt.Sprintf(`
run_pipeline_definition = [%v]

built_pipeline = Pipeline(workspace=ws, steps=[run_pipeline_definition])
pipeline_run = experiment.submit(built_pipeline)
`, step_string)

	main_section := `

if __name__ == "__main__":
    # dotenv.load_dotenv()

    credentials_dict = {
        "AML_SP_PASSWORD_VALUE": os.environ.get("AML_SP_PASSWORD_VALUE"),
        "AML_SP_TENANT_ID": os.environ.get("AML_SP_TENANT_ID"),
        "AML_SP_APP_ID": os.environ.get("AML_SP_APP_ID"),
        "WORKSPACE_SUBSCRIPTION_ID": os.environ.get("WORKSPACE_SUBSCRIPTION_ID"),
        "WORKSPACE_RESOURCE_GROUP": os.environ.get("WORKSPACE_RESOURCE_GROUP"),
        "WORKSPACE_NAME": os.environ.get("WORKSPACE_NAME"),
        "AML_COMPUTE_NAME": os.environ.get("AML_COMPUTE_NAME"),
    }

    # execute only if run as a script
    root(
        context="gAR9lC4=", metadata_url="", aml_workspace_credentials=credentials_dict
    )

`
	return rootFile_pre_imports +
		get_aml_workspace +
		provision_aml_compute +
		global_packages_to_install +
		root_pre_code +
		all_code +
		pipeline_submit +
		main_section, nil
}

func (c *CompileLive) WriteStepFiles(target string, compiledDir string, aggregatedSteps map[string]CodeBlock) error {
	for i := range aggregatedSteps {
		parameter_string, _ := JoinMapKeysValues(aggregatedSteps[i].Parameters)
		if parameter_string != "" {
			parameter_string = "," + parameter_string
		}

		// Prepend an empty locals as the default
		parameter_string = "__context='gAR9lC4=', __run_info={}, __metadata_url=''" + parameter_string

		step_to_write := compiledDir + fmt.Sprintf("/%v.py", aggregatedSteps[i].Step_Identifier)
		code_to_write := fmt.Sprintf(`
import argparse as __argparse
from multiprocessing import context
import pathlib
from typing import NamedTuple
import dotenv
from azureml.core import Run
from pprint import pprint as __pp
import os
from pathlib import Path as __Path
from azureml.pipeline.core import (
    PipelineData as __PipelineData,
    PipelineParameter as __PipelineParameter,

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
		"experiment_id": __run_info_dict["experiment"],
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

`, parameter_string, aggregatedSteps[i].Step_Identifier)

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
		"experiment_id": __run_info_dict["experiment"],
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

	
    return __loc["__b64_string"]
`, aggregatedSteps[i].Step_Identifier)

		code_to_write += `
if __name__ == "__main__":
    dotenv.load_dotenv()

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

    __output_context_string = main(
        __context=__input_context_string,
        __run_info=__args.run_info,
        __metadata_url=__args.metadata_url,
    )

    __p = __Path(__args.output_context_path)
    __p.mkdir(parents=True, exist_ok=True)
    __filepath = __p / __context_filename
    with __filepath.open("w+") as __f:
        __f.write(__output_context_string)
`

		err := os.WriteFile(step_to_write, []byte(code_to_write), 0700)
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
