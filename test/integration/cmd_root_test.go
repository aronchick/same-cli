package integration_test

import (
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	origHome := os.Getenv("HOME")

	// Set to empty HOME
	os.Setenv("HOME", "/tmp")
	fatal = false
	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "create")

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	assert.Equal(suite.T(), true, fatal)
	assert.Contains(suite.T(), string(out), "Nil file or empty load config settings")
	os.Setenv("HOME", origHome)

}

func (suite *RootSuite) Test_BadConfig() {
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "--config", "/tmp/badfile.yaml", "create")

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	assert.Equal(suite.T(), true, fatal)
	assert.Contains(suite.T(), string(out), "Nil file or empty load config settings")
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}
