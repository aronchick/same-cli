package integration_test

import (
	"os"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramUtilsSuite struct {
	suite.Suite
	rootCmd *cobra.Command
}

// Before all suite
func (suite *ProgramUtilsSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")
}

// Before each test
func (suite *ProgramUtilsSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
}

func (suite *ProgramUtilsSuite) TearDownAllSuite() {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ProgramUtilsSuite) Test_EmptyPipelineParameters() {
	os.Setenv("TEST_PASS", "1")
	_, _ = cmd.FindPipelineByName("")
	//assert.Contains(suite.T(), string(out), "same program [command]")
}

func TestProgramUtilsSuite(t *testing.T) {
	suite.Run(t, new(ProgramUtilsSuite))
}
