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
	suite.remoteSAMEURL = "https://github.com/SAME-Project/Sample-SAME-Data-Science"
	suite.fatal = false
	viper.Reset()
	log.SetOutput(ioutil.Discard)
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

func (suite *InitSuite) Test_BadTarget() {
	out := execute_target(suite, "UNKNOWN", true, "")
	assert.Contains(suite.T(), string(out), "Setup target 'unknown' not understood")
}

func (suite *InitSuite) Test_AKSTarget() {
	out := execute_target(suite, "aks", false, "")
	assert.Contains(suite.T(), string(out), "Executing AKS setup.")
}

func (suite *InitSuite) Test_LocalTarget() {
	out := execute_target(suite, "local", false, "")
	assert.Contains(suite.T(), string(out), "Executing local setup")
}
func (suite *InitSuite) Test_NoDocker() {
	out := execute_target(suite, "local", true, "no-docker-path")
	assert.Contains(suite.T(), string(out), "not find docker in your PATH")
}

func (suite *InitSuite) Test_NotDockerGroupOnSystem() {
	out := execute_target(suite, "local", true, "no-docker-group-on-system")
	assert.Contains(suite.T(), string(out), "could not find the group")
}

func (suite *InitSuite) Test_CouldNotRetrieveGroups() {
	out := execute_target(suite, "local", true, "cannot-retrieve-groups")
	assert.Contains(suite.T(), string(out), "could not retrieve a list of groups")
}

func (suite *InitSuite) Test_NotInDockerGroup() {
	out := execute_target(suite, "local", true, "not-in-docker-group")
	assert.Contains(suite.T(), string(out), "user not in the 'docker' group")
}

func (suite *InitSuite) Test_K3sInstallFailed() {
	out := execute_target(suite, "local", true, "k3s-not-detected")
	assert.Contains(suite.T(), string(out), "K3S NOT DETECTED")
}

func (suite *InitSuite) Test_KFPLocalInstallFailed() {
	out := execute_target(suite, "local", true, "kfp-install-failed")
	assert.Contains(suite.T(), string(out), "INSTALL KFP FAILED")
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}

func execute_target(suite *InitSuite, target string, fatal bool, additionalFlag string) (out string) {
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

	assert.Equal(suite.T(), fatal, suite.fatal)
	return out

}
