package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
<<<<<<< HEAD
=======
	"github.com/azure-octo/same-cli/pkg/utils"
>>>>>>> 275781737c8fd9f42dccf298a0078b2f645f9bd9
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

<<<<<<< HEAD
=======
	"github.com/stretchr/testify/assert"
>>>>>>> 275781737c8fd9f42dccf298a0078b2f645f9bd9
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type InstallDockerSuite struct {
	suite.Suite
	rootCmd *cobra.Command
	fatal   bool
}

// before each test
func (suite *InstallDockerSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.fatal = false
	viper.Reset()
	log.SetOutput(ioutil.Discard)
	os.Setenv("TEST_PASS", "1")
}

func (suite *InstallDockerSuite) TearDownTest() {

}

<<<<<<< HEAD
// TODO: Commenting out because we've got to figure out how to test under sudo
// func (suite *InstallDockerSuite) Test_RunDefault() {
// 	viper.Reset()
// 	defer func() { log.StandardLogger().ExitFunc = nil }()
// 	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

// 	viper.SetEnvPrefix("same") // will be uppercased automatically

// 	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "install_docker", "--config", "../testdata/config/notarget.yaml", "--", "--unittestmode", "")

// 	// Putting empty assignments here for debugging in the future
// 	_ = command
// 	_ = err

// 	assert.Equal(suite.T(), false, suite.fatal)
// 	assert.Contains(suite.T(), string(out), "user not in the 'docker' group")
// }
=======
func (suite *InstallDockerSuite) Test_RunDefault() {
	viper.Reset()
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	viper.SetEnvPrefix("same") // will be uppercased automatically

	command, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "install_docker", "--config", "../testdata/config/notarget.yaml", "--", "--unittestmode", "")

	// Putting empty assignments here for debugging in the future
	_ = command
	_ = err

	assert.Equal(suite.T(), false, suite.fatal)
	assert.Contains(suite.T(), string(out), "user not in the 'docker' group")
}
>>>>>>> 275781737c8fd9f42dccf298a0078b2f645f9bd9

func TestInstallDockerSuite(t *testing.T) {
	suite.Run(t, new(InstallDockerSuite))
}
