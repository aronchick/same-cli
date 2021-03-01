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
type InstallK3sSuite struct {
	suite.Suite
	rootCmd *cobra.Command
	fatal   bool
}

// before each test
func (suite *InstallK3sSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.fatal = false
	viper.Reset()
	log.SetLevel(log.TraceLevel)
	// log.SetOutput(ioutil.Discard)
	os.Setenv("TEST_PASS", "1")

	i := utils.Installers{}
	_, err := i.DetectK3s("k3s")
	if err == nil {
		log.Warn("k3s detected, please uninstall with\n./pkg/infra/k3s-uninstall.sh")
	}

}

func (suite *InstallK3sSuite) TearDownTest() {

}

func (suite *InstallK3sSuite) Test_AssertPass() {
	// Just a placeholder test until we figure out what to test for real.
	assert.True(suite.T(), true)
}

// TODO: Commenting out because we've got to figure out how to test under sudo
// func (suite *InstallK3sSuite) Test_RunDefault() {
// 	viper.Reset()
// 	currentUser, _ := user.Current()
// 	os.Setenv("SUDO_UID", currentUser.Uid)
// 	defer func() { log.StandardLogger().ExitFunc = nil }()
// 	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }
// 	viper.SetEnvPrefix("same") // will be uppercased automatically

// 	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "installK3s", "--config", "../testdata/config/notarget.yaml", "--", "--unittestmode", "")

// 	// Putting empty assignments here for debugging in the future
// 	_ = command
// 	_ = err

// 	assert.Equal(suite.T(), false, suite.fatal)
// 	assert.Contains(suite.T(), string(out), "user not in the 'docker' group")
// }

func TestInstallK3sSuite(t *testing.T) {
	suite.Run(t, new(InstallK3sSuite))
}
