package integration_test

import (
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type RootSuite struct {
	suite.Suite
	rootCmd       *cobra.Command
	remoteSAMEURL string
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *RootSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/Sample-SAME-Data-Science"
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *RootSuite) Test_NoConfigDir() {
	origHome := os.Getenv("HOME")

	// Set to empty HOME
	os.Setenv("HOME", "/tmp")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "")
	assert.Contains(suite.T(), string(out), "same [command]")
	os.Setenv("HOME", origHome)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}
