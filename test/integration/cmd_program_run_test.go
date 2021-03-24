package integration_test

import (
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/onsi/gomega/gbytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramRunSuite struct {
	suite.Suite
	rootCmd       *cobra.Command
	remoteSAMEURL string
	logBuf        *gbytes.Buffer
}

// Before all suite
func (suite *ProgramRunSuite) SetupAllSuite() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/goodpipeline.yaml")
	if out != "" {
		log.Printf("not sure if this is a bad thing, there's an output from creating the pipeline during setup: %v", string(out))
	}

	if os.Getenv("GITHUB_ACTIONS") != "" {
		running, err := utils.GetUtils().K3sRunning(suite.rootCmd)
		if err != nil || !running {
			log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
		}
	}

	suite.logBuf = gbytes.NewBuffer()
}

// Before each test
func (suite *ProgramRunSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	if ok, _ := utils.KFPReady(suite.rootCmd); !ok {
		log.Warn("KFP does not appear to be ready, this may cause tests to fail.")
	}

}

// After test
func (suite *ProgramRunSuite) TearDownAllSuite() {
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "delete", "-i", "")
	if out != "" {
		log.Printf("not sure if this is a bad thing, there's an output from deleting the pipeline during teardown: %v", string(out))
	}

}

func TestProgramRunSuite(t *testing.T) {
	suite.Run(t, new(ProgramRunSuite))
}
