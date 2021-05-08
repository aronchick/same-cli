package integration_test

import (
	"io/ioutil"
	"os"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
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
	cb := codeBlocks["same-step-1"]
	assert.Equal(suite.T(), cb.Cache_Value, "P0D", "Expected to set a missing cache value to P0D. Actual: %v", cb.Cache_Value)

}

func (suite *ProgramCompileSuite) Test_SettingCacheValue_WithCache() {
	os.Setenv("TEST_PASS", "1")
	c := utils.GetCompileFunctions()

	foundSteps, _ := c.FindAllSteps(ONE_STEP_WITH_CACHE)
	codeBlocks, _ := c.CombineCodeSlicesToSteps(foundSteps)
	cb := codeBlocks["same-step-1"]
	assert.Equal(suite.T(), cb.Cache_Value, "P20D", "Expected to set a missing cache value to P20D. Actual: %v", cb.Cache_Value)

}

func (suite *ProgramCompileSuite) TearDownAllSuite() {
	// os.RemoveAll(suite.tmpConfigDirectory)
	log.Warnf("Directory: %v", suite.tmpDirectory)
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
# + tags=["same-step-1"]
import tensorflow
`

	ONE_STEP_WITH_CACHE = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same-step-1", "cache=P20D"]
import tensorflow
`

	TWO_STEPS = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same-step-1"]
import tensorflow

# +
# + tags=["same-step-2"]
import pytorch
`

	TWO_STEPS_COMBINE = `
# ---

# + tags=["parameters"]
foo = "bar"

# +
# + tags=["same-step-1"]
import tensorflow

# +
# + tags=["same-step-1"]
import numpy

# +
# + tags=["same-step-2"]
import pytorch
`

	TWO_STEPS_COMBINE_NO_PARAMS = `
# +
# + tags=["same-step-1"]
import tensorflow

# +
# + tags=["same-step-1"]
import numpy

# +
# + tags=["same-step-2"]
import pytorch
`
)
