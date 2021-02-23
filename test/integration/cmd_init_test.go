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
type InitSuite struct {
	suite.Suite
	rootCmd       *cobra.Command
	remoteSAMEURL string
	fatal         bool
}

// before each test
func (suite *InitSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/Sample-SAME-Data-Science"
	suite.fatal = false
	os.Setenv("TEST_PASS", "1")
}

func (suite *InitSuite) TearDownTest() {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *InitSuite) Test_EmptyConfig() {
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init", "--config", "../testdata/config/emptyconfig.yaml")

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	assert.Equal(suite.T(), true, suite.fatal)
	assert.Contains(suite.T(), string(out), "Nil file or empty load config settings")
}

func (suite *InitSuite) Test_NoTargetSet() {
	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init", "--config", "../testdata/config/notarget.yaml")

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	assert.Equal(suite.T(), false, suite.fatal)
	assert.Contains(suite.T(), string(out), "No 'target' set for deployment")
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}
