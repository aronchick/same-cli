package integration_test

import (
	"io/ioutil"
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
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	suite.fatal = false
	viper.Reset()
	// log.SetOutput(ioutil.Discard)
	os.Setenv("TEST_PASS", "1")
}

func (suite *InitSuite) TearDownTest() {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *InitSuite) Test_EmptyConfig() {
	viper.Reset()
	os.Unsetenv("SAME_TARGET")
	os.Setenv("TEST_PASS", "1")
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init", "--config", "../testdata/config/emptyconfig.yaml")

	assert.Contains(suite.T(), string(out), "as a default")
}

func (suite *InitSuite) Test_BadTarget() {
	out := execute_target(suite, "UNKNOWN", "")
	assert.Contains(suite.T(), string(out), "Setup target 'unknown' not understood")
	assert.Equal(suite.T(), true, suite.fatal)
}

func (suite *InitSuite) Test_AKSTarget() {
	out := execute_target(suite, "aks", "")
	assert.Contains(suite.T(), string(out), "Executing AKS setup.")
	assert.Equal(suite.T(), false, suite.fatal)
}

func (suite *InitSuite) Test_LocalTarget() {
	out := execute_target(suite, "local", "")
	assert.Contains(suite.T(), string(out), "Executing local setup")
	assert.Equal(suite.T(), false, suite.fatal)
}

// COMMENTING OUT TEST UNTIL UTILS.MOCKS COMPLETE
// func (suite *InitSuite) Test_K3sInstallFailed() {
// 	os.Setenv("TEST_PASS", "1")
// 	out := execute_target(suite, "local", mocks.INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_PROBE)
// 	assert.Contains(suite.T(), string(out), mocks.INIT_TEST_K3S_STARTED_BUT_SERVICES_FAILED_RESULT, "Testing for failed K3s installation did not work.")
// 	assert.Equal(suite.T(), true, suite.fatal)
// }

func (suite *InitSuite) Test_KFPLocalInstallFailed() {
	os.Setenv("TEST_PASS", "1")
	out := execute_target(suite, "local", "kfp-install-failed")
	assert.Contains(suite.T(), string(out), "INSTALL KFP FAILED")
	assert.Equal(suite.T(), true, suite.fatal)

}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}

func execute_target(suite *InitSuite, target string, additionalFlag string) (out string) {
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	viper.SetEnvPrefix("same") // will be uppercased automatically
	err := viper.BindEnv("target")
	if err != nil {
		assert.Failf(suite.T(), "could not bind viper to 'target': %v ", err.Error())
	}

	os.Setenv("SAME_TARGET", target) // typically done outside of the app

	tmpFile, _ := ioutil.TempFile(os.TempDir(), "SAME-TEST-RUN-CONFIG-*.yaml")
	defer os.Remove(tmpFile.Name())

	text, _ := ioutil.ReadFile("../testdata/config/notarget.yaml")
	if _, err = tmpFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init", "--config", tmpFile.Name(), "--", "--unittestmode", additionalFlag)

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	return out

}
