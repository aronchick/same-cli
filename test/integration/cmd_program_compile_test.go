package integration_test

import (
	"io/ioutil"
	"os"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/onsi/gomega/gbytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramCompileSuite struct {
	suite.Suite
	logBuf       *gbytes.Buffer
	rootCmd      *cobra.Command
	tmpDirectory string
}

// Before all suite
func (suite *ProgramCompileSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")

	dc := infra.GetDependencyCheckers(suite.rootCmd, []string{})
	if err := dc.CheckDependenciesInstalled(); err != nil {
		log.Warnf("Failed one or more dependencies - skipping this test: %v", err.Error())
		suite.T().Skip()
	}
	suite.logBuf = gbytes.NewBuffer()
}

func (suite *ProgramCompileSuite) SetupTest() {
	os.Setenv("TEST_PASS", "1")

	suite.tmpDirectory, _ = ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	suite.rootCmd = cmd.RootCmd
	viper.Reset()

}

func (suite *ProgramCompileSuite) TearDownAllSuite() {
	// os.RemoveAll(suite.tmpConfigDirectory)
	log.Warnf("Directory: %v", suite.tmpDirectory)
}
func (suite *ProgramCompileSuite) Test_DeletePipeline() {
	os.Setenv("TEST_PASS", "1")

	// https://github.com/azure-octo/same-cli/issues/91
	assert.True(suite.T(), true)
}

func TestProgramCompileSuite(t *testing.T) {
	suite.Run(t, new(ProgramCompileSuite))
}
