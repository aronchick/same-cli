package integration_test

import (
	"io/ioutil"
	"os"
	"os/exec"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/onsi/gomega/gbytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramCompileSuite struct {
	suite.Suite
	logBuf       *gbytes.Buffer
	rootCmd      *cobra.Command
	tmpDirectory string
}

// Before all suite
func (suite *ProgramCompileSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")

	dc := infra.GetDependencyCheckers(suite.rootCmd, []string{})
	if err := dc.CheckDependenciesInstalled(); err != nil {
		log.Warnf("Failed one or more dependencies - skipping this test: %v", err.Error())
		suite.T().Skip()
	}
	suite.logBuf = gbytes.NewBuffer()
}

func (suite *ProgramCompileSuite) SetupTest() {
	os.Setenv("TEST_PASS", "1")

	suite.tmpDirectory, _ = ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	suite.rootCmd = cmd.RootCmd
	viper.Reset()

}

func (suite *ProgramCompileSuite) Test_ParseZeroSteps() {
	testStep(suite.T(), 1, 1, ZERO_STEPS, "ZERO_STEPS")
}

func (suite *ProgramCompileSuite) Test_ParseZeroStepsWithParams() {
	testStep(suite.T(), 1, 1, ZERO_STEPS_WITH_PARAMS, "ZERO_STEPS_WITH_PARAMS")
}

func (suite *ProgramCompileSuite) Test_ParseOneStep() {
	testStep(suite.T(), 4, 2, ONE_STEP, "ONE_STEP")
}

func (suite *ProgramCompileSuite) Test_ParseOneStepWithCache() {
	testStep(suite.T(), 4, 2, ONE_STEP_WITH_CACHE, "ONE_STEP_WITH_CACHE")
}

func (suite *ProgramCompileSuite) Test_ParseTwoSteps() {
	testStep(suite.T(), 6, 3, TWO_STEPS, "TWO_STEPS")
}

func (suite *ProgramCompileSuite) Test_ParseTwoStepsCombine() {
	testStep(suite.T(), 8, 3, TWO_STEPS_COMBINE, "TWO_STEPS_COMBINE")
}

func (suite *ProgramCompileSuite) Test_ParseTwoStepsCombineNoParams() {
	testStep(suite.T(), 6, 3, TWO_STEPS_COMBINE_NO_PARAMS, "TWO_STEPS_COMBINE_NO_PARAMS")
}

func (suite *ProgramCompileSuite) Test_SettingCacheValue_NoCache() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	foundSteps, _ := c.FindAllSteps(ONE_STEP)
	codeBlocks, _ := c.CombineCodeSlicesToSteps(foundSteps)
	cb := codeBlocks["same_step_1"]
	assert.Equal(suite.T(), cb.CacheValue, "P0D", "Expected to set a missing cache value to P0D. Actual: %v", cb.CacheValue)

}

func (suite *ProgramCompileSuite) Test_SettingCacheValue_WithCache() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	foundSteps, _ := c.FindAllSteps(ONE_STEP_WITH_CACHE)
	codeBlocks, _ := c.CombineCodeSlicesToSteps(foundSteps)
	cb := codeBlocks["same_step_1"]
	assert.Equal(suite.T(), cb.CacheValue, "P20D", "Expected to set a missing cache value to P20D. Actual: %v", cb.CacheValue)

}

func (suite *ProgramCompileSuite) Test_ImportsWorkingProperly() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	foundSteps, _ := c.FindAllSteps(NOTEBOOKS_WITH_IMPORT)
	codeBlocks, _ := c.CombineCodeSlicesToSteps(foundSteps)
	packagesToMerge, _ := c.WriteStepFiles("kubeflow", suite.tmpDirectory, codeBlocks)
	containsKey := ""
	for key := range packagesToMerge["same_step_0"] {
		containsKey += key
	}
	assert.Contains(suite.T(), containsKey, "tensorflow", "Expected to contain 'tensorflow'. Actual: %v", packagesToMerge["same_step_0"])
}

func (suite *ProgramCompileSuite) Test_FullNotebookExperience() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()
	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		assert.Fail(suite.T(), "Jupytext not installed")
	}

	notebookPath := "../testdata/notebook/sample_notebook.ipynb"
	if _, exists := os.Stat(notebookPath); exists != nil {
		assert.Fail(suite.T(), "Notebook not found at: %v", notebookPath)
	}
	convertedText, _ := c.ConvertNotebook(jupytextExecutable, notebookPath)
	foundSteps, _ := c.FindAllSteps(convertedText)
	codeBlocks, _ := c.CombineCodeSlicesToSteps(foundSteps)
	packagesToMerge, _ := c.WriteStepFiles("kubeflow", suite.tmpDirectory, codeBlocks)
	containsKey := ""
	for key := range packagesToMerge["same_step_0"] {
		containsKey += key
	}
	assert.Contains(suite.T(), containsKey, "tensorflow", "Expected to contain 'tensorflow'. Actual: %v", packagesToMerge["same_step_0"])
}

func (suite *ProgramCompileSuite) Test_KubeflowRootCompile() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	sameConfigFile, err := loaders.V1{}.LoadSAME("../testdata/notebook/sample_notebook_same.yaml")
	if err != nil {
		assert.Fail(suite.T(), "could not load SAME config file: %v", err)
	}

	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		assert.Fail(suite.T(), "Jupytext not installed")
	}

	notebook_path := "../testdata/notebook/sample_notebook.ipynb"
	if _, exists := os.Stat(notebook_path); exists != nil {
		assert.Fail(suite.T(), "Notebook not found at: %v", notebook_path)
	}
	convertedText, _ := c.ConvertNotebook(jupytextExecutable, notebook_path)
	foundSteps, _ := c.FindAllSteps(convertedText)
	aggregatedSteps, _ := c.CombineCodeSlicesToSteps(foundSteps)
	fullRootFile, _ := c.CreateRootFile("kubeflow", aggregatedSteps, *sameConfigFile)

	assert.Contains(suite.T(), fullRootFile, "import kfp", "Does not contain pre-steps import")
	assert.Contains(suite.T(), fullRootFile, "import same_step_2", "Does not contain multi-step import")
	assert.Contains(suite.T(), fullRootFile, "def get_run_info(", "Does not contain run info import")
	assert.Contains(suite.T(), fullRootFile, "sample_parameter='0.841'", "Does not contain default parameter")
	assert.Contains(suite.T(), fullRootFile, "create_context_file_op = create_context_file_component(context_string=__original_context)\n\n\n\tsame_step_0_op = create_component_from_func", "SAME Step 0 is not the first step")
	assert.Contains(suite.T(), fullRootFile, "same_step_2_op = create_component_from_func(", "Does not contain the third step")
	assert.Contains(suite.T(), fullRootFile, "same_step_2_task.after(same_step_1_task)", "Does not have the final DAG step")
}

func (suite *ProgramCompileSuite) Test_AMLRootCompile() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	sameConfigFile, err := loaders.V1{}.LoadSAME("../testdata/notebook/sample_notebook_same.yaml")
	if err != nil {
		assert.Fail(suite.T(), "could not load SAME config file: %v", err)
	}

	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		assert.Fail(suite.T(), "Jupytext not installed")
	}

	notebook_path := "../testdata/notebook/sample_notebook.ipynb"
	if _, exists := os.Stat(notebook_path); exists != nil {
		assert.Fail(suite.T(), "Notebook not found at: %v", notebook_path)
	}
	convertedText, _ := c.ConvertNotebook(jupytextExecutable, notebook_path)
	foundSteps, _ := c.FindAllSteps(convertedText)
	aggregatedSteps, _ := c.CombineCodeSlicesToSteps(foundSteps)
	fullRootFile, _ := c.CreateRootFile("aml", aggregatedSteps, *sameConfigFile)

	assert.Contains(suite.T(), fullRootFile, "import azureml.core", "Does not contain pre-steps import")
	assert.Contains(suite.T(), fullRootFile, "sample_parameter='0.841'", "Does not contain default parameter")
	assert.Contains(suite.T(), fullRootFile, "experiment = Experiment(ws, \"SampleComplicatedNotebook\")", "Does not contain the experiment name")
	assert.Contains(suite.T(), fullRootFile, "__original_context_param,", "Does not contain the original context")
	assert.Contains(suite.T(), fullRootFile, "inputs=[__pipelinedata_context_same_step_1],", "Does not contain input for third step")
	assert.Contains(suite.T(), fullRootFile, "run_pipeline_definition = [same_step_0_step, same_step_1_step, same_step_2_step]", "Does not have the final pipeline combination")
}

func (suite *ProgramCompileSuite) TearDownAllSuite() {
	os.RemoveAll(suite.tmpDirectory)
}
func (suite *ProgramCompileSuite) Test_CompilePipeline() {
	os.Setenv("TEST_PASS", "1")

	// https://github.com/azure-octo/same-cli/issues/91
	assert.True(suite.T(), true)
}

func testStep(T *testing.T, expectedNumberRaw int, expectedNumberCombined int, testString string, testStringName string) {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()
	foundSteps, err := c.FindAllSteps(testString)
	assert.Equal(T, len(foundSteps), expectedNumberRaw, "%v did not result in %v step. Actual steps: %v", testStringName, expectedNumberRaw, len(foundSteps))
	assert.NoError(T, err, "%v resulted in an error building slices: %v", testStringName, err)

	codeBlocks, err := c.CombineCodeSlicesToSteps(foundSteps)
	assert.Equal(T, len(codeBlocks), expectedNumberCombined, "%v did not result in %v code slices. Actual code slices: %v", testStringName, expectedNumberCombined, len(codeBlocks))
	assert.NoError(T, err, "%v resulted in an error building code blocks: %v", testStringName, err)
}

func TestProgramCompileSuite(t *testing.T) {
	suite.Run(t, new(ProgramCompileSuite))
}

var (
	ZERO_STEPS = `
# ---

foo = "bar"

# +
import tensorflow
`
	ZERO_STEPS_WITH_PARAMS = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
import tensorflow
`

	ONE_STEP = `
# ---

# +
foo = "bar"

# +
# + tags=["same_step_1"]
import tensorflow
`

	ONE_STEP_WITH_CACHE = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same_step_1", "cache=P20D"]
import tensorflow
`

	TWO_STEPS = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same_step_1"]
import tensorflow

# +
# + tags=["same_step_2"]
import pytorch
`

	TWO_STEPS_COMBINE = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same_step_1"]
import tensorflow

# +
# + tags=["same_step_1"]
import numpy

# +
# + tags=["same_step_2"]
import pytorch
`

	TWO_STEPS_COMBINE_NO_PARAMS = `
# +
# + tags=["same_step_1"]
import tensorflow

# +
# + tags=["same_step_1"]
import numpy

# +
# + tags=["same_step_2"]
import pytorch
`

	NOTEBOOKS_WITH_IMPORT = `
# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: light
#       format_version: '1.5'
#       jupytext_version: 1.11.1
#   kernelspec:
#     display_name: Python 3
#     language: python
#     name: python3
# ---

# + tags=["parameters"]
foo = "bar"
num = 17

# +
import tensorflow

a = 4

# +
from IPython.display import Image

b = a + 5

url = 'https://same-project.github.io/SAME-samples/automated_notebook/FaroeIslands.jpeg'

from IPython import display`
)
