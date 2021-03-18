package integration_test

import (
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/spf13/cobra"

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

// before each test
func (suite *RootSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
}

// COMMENTING OUT TEST UNTIL UTILS.MOCKS COMPLETE
// func (suite *RootSuite) Test_NoConfigDir() {
// 	viper.Reset()
// 	defer func() { log.StandardLogger().ExitFunc = nil }()
// 	var fatal bool
// 	log.StandardLogger().ExitFunc = func(int) { fatal = true }

// 	origHome := os.Getenv("HOME")

// 	// Set to empty HOME
// 	os.Setenv("HOME", "/tmp")
// 	fatal = false
// 	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init")

// 	// Putting empty assignments here for debugging in the future
// 	_ = command
// 	_ = err

// 	assert.Equal(suite.T(), true, fatal)
// 	assert.Contains(suite.T(), string(out), "No config file found")
// 	os.Setenv("HOME", origHome)

// }

// COMMENTING OUT TEST UNTIL UTILS.MOCKS COMPLETE
// func (suite *RootSuite) Test_BadConfig() {
// 	viper.Reset()
// 	defer func() { log.StandardLogger().ExitFunc = nil }()
// 	var fatal bool
// 	log.StandardLogger().ExitFunc = func(int) { fatal = true }

// 	fatal = false
// 	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "--config", "/tmp/badfile.yaml", "init")

// 	// Putting empty assignments here for debugging in the future
// 	_ = command
// 	_ = err

// 	assert.Equal(suite.T(), true, fatal)
// 	assert.Contains(suite.T(), string(out), "No config file found")
// }

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}
