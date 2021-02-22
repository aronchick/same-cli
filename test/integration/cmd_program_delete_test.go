package integration_test

import (
	"fmt"
	"regexp"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/utils"
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
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ProgramDeleteSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd

	c, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/deletepipeline.yaml")
	_ = c
	_ = out

	if err != nil {
		message := fmt.Errorf("could not start delete test suite because we could not create a sample pipeline: %v", err)
		suite.rootCmd.Print(message)
		log.Fatalf(message.Error())
	}

	r := regexp.MustCompile(`Name:\s+([^\n]+)\nID:\s+([^\n]+)`)
	rs := r.FindStringSubmatch(string(out))
	if len(rs) < 2 {
		log.Fatalf("cmd_program_delete_test: during setup, could not find name and ID in the returned upload string: %v", out)
	}
	fmt.Printf("%#v\n", rs[1])
	fmt.Printf("%#v\n", rs[2])
	suite.pipelineName = rs[1]
	suite.pipelineID = rs[2]
}

func (suite *ProgramDeleteSuite) Test_DeletePipeline() {
	_, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "delete", "-i", suite.pipelineID)
	assert.Contains(suite.T(), string(out), "Successfully deleted pipeline ID")
	assert.NoError(suite.T(), err, fmt.Sprintf("Error found (non expected): %v", err))
}

func TestProgramDeleteSuite(t *testing.T) {
	suite.Run(t, new(ProgramDeleteSuite))
}
