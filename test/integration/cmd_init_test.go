package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/mocks"
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
	dc            infra.DependencyCheckers
}

// before each test
func (suite *InitSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	suite.fatal = false
	suite.dc = &infra.LiveDependencyCheckers{}
	suite.dc.SetCmd(suite.rootCmd)
	suite.dc.SetCmdArgs([]string{})
	if ok, err := suite.dc.CanConnectToKubernetes(); !ok && (err != nil) {
		assert.Fail(suite.T(), "Cannot run tests because we cannot connect to a live cluster. Test this with: kubectl version")
	}
	viper.Reset()
	os.Setenv("TEST_PASS", "1")
}

func (suite *InitSuite) TearDownTest() {

}

func (suite *InitSuite) Test_KFPInstallFailed() {
	os.Setenv("TEST_PASS", "1")
	out := execute_target(suite, mocks.DEPENDENCY_CHECKER_KFP_INSTALL_FAILED_PROBE)
	assert.Contains(suite.T(), string(out), mocks.DEPENDENCY_CHECKER_KFP_INSTALL_FAILED_RESULT)
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}

func execute_target(suite *InitSuite, additionalFlag string) (out string) {
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	viper.SetEnvPrefix("same") // will be uppercased automatically

	tmpFile, _ := ioutil.TempFile(os.TempDir(), "SAME-TEST-RUN-CONFIG-*.yaml")
	defer os.Remove(tmpFile.Name())

	text, _ := ioutil.ReadFile("../testdata/config/notarget.yaml")
	if _, err := tmpFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "init", "--config", tmpFile.Name(), "--", additionalFlag)

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	return out

}
