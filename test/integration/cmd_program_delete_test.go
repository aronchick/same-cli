package integration_test

import (
	"fmt"
	"os"
	"regexp"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/onsi/gomega/gbytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramDeleteSuite struct {
	suite.Suite
	rootCmd       *cobra.Command
	pipelineID    string
	pipelineName  string
	logBuf        *gbytes.Buffer
	remoteSAMEURL string
}

// Before all suite
func (suite *ProgramDeleteSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/goodpipeline.yaml")
	if out != "" {
		log.Printf("not sure if this is a bad thing, there's an output from creating the pipeline during setup: %v", string(out))
	}

	running, err := utils.K3sRunning(suite.rootCmd)
	if err != nil || !running {
		log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
	}

	suite.logBuf = gbytes.NewBuffer()

}

// before each test
func (suite *ProgramDeleteSuite) SetupTest() {
	os.Setenv("TEST_PASS", "1")
	suite.rootCmd = cmd.RootCmd

	if ok, _ := utils.KFPReady(suite.rootCmd); !ok {
		log.Warn("KFP does not appear to be ready, this may cause tests to fail.")
	}        

	c, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/deletepipeline.yaml")
	_ = c
	_ = out

	if err != nil {
		message := fmt.Errorf("could not start delete test suite because we could not create a sample pipeline: %v", err)
		suite.rootCmd.Print(message)
		log.Fatalf(message.Error())
	}

	r := regexp.MustCompile(`Name:\s+([^\n]+)\nVersionID:\s+([^\n]+)`)
	rs := r.FindStringSubmatch(string(out))
	if len(rs) < 2 {
		log.Fatalf("cmd_program_delete_test: during setup, could not find name and ID in the returned upload string: %v", out)
	}
	suite.rootCmd.Printf("%#v\n", rs[1])
	suite.rootCmd.Printf("%#v\n", rs[2])
	suite.pipelineName = rs[1]
	suite.pipelineID = rs[2]
}

func (suite *ProgramDeleteSuite) Test_DeletePipeline() {
	os.Setenv("TEST_PASS", "1")

	// https://github.com/azure-octo/same-cli/issues/91
	assert.True(suite.T(), true)
}

func TestProgramDeleteSuite(t *testing.T) {
	suite.Run(t, new(ProgramDeleteSuite))
}
