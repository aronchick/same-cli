package integration_test

import (
	"fmt"
	"os"
	"regexp"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/google/uuid"
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
	rootCmd      *cobra.Command
	pipelineID   string
	pipelineName string
	logBuf       *gbytes.Buffer
	dc           infra.DependencyCheckers
}

// Before all suite
func (suite *ProgramDeleteSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")

	if os.Getenv("TEST_K3S") == "true" {
		running, err := utils.GetUtils().IsK3sRunning(suite.rootCmd)
		if err != nil || !running {
			log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
		}
	}

	dc := infra.GetDependencyCheckers(suite.rootCmd, []string{})
	if err := dc.CheckDependenciesInstalled(suite.rootCmd); err != nil {
		log.Warnf("Failed one or more dependencies - skipping this test: %v", err.Error())
		suite.T().Skip()
	}
	suite.logBuf = gbytes.NewBuffer()

}

// before each test
func (suite *ProgramDeleteSuite) SetupTest() {
	os.Setenv("TEST_PASS", "1")
	suite.rootCmd = cmd.RootCmd

	suite.dc = &infra.LiveDependencyCheckers{}
	if ok, err := suite.dc.CanConnectToKubernetes(suite.rootCmd); !ok && (err != nil) {
		assert.Fail(suite.T(), `Cannot run tests because we cannot connect to a live cluster. Test your cluster with:  kubectl version`)
	}

	if ok, _ := utils.IsKFPReady(suite.rootCmd); !ok {
		log.Warn("KFP does not appear to be ready, this may cause tests to fail.")
	}

	runID, _ := uuid.NewRandom()
	c, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/samefiles/deletepipeline.yaml", "-e", "test-experiment", "-r", runID.String())
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
