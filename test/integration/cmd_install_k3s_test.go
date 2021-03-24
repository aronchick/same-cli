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
	// log.SetLevel(log.TraceLevel)
	// log.SetOutput(ioutil.Discard)
	os.Setenv("TEST_PASS", "1")

	_, err := utils.GetUtils().DetectK3s()
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

func TestInstallK3sSuite(t *testing.T) {
	suite.Run(t, new(InstallK3sSuite))
}
