package integration_test

import (
	"fmt"
	"io/ioutil"
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
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/Sample-SAME-Data-Science"
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/goodpipeline.yaml")
	if out != "" {
		log.Printf("not sure if this is a bad thing, there's an output from creating the pipeline during setup: %v", string(out))
	}
	suite.logBuf = gbytes.NewBuffer()
}

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
	suite.rootCmd.Printf("%#v\n", rs[1])
	suite.rootCmd.Printf("%#v\n", rs[2])
	suite.pipelineName = rs[1]
	suite.pipelineID = rs[2]

	// suite.logBuf = gbytes.NewBuffer()
	// log.SetOutput(suite.logBuf)
	// defer func() {
	// 	log.SetOutput(os.Stderr)
	// }()

	log.SetOutput(ioutil.Discard)
}

func (suite *ProgramDeleteSuite) Test_DeletePipeline() {
	_, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "delete", "-i", suite.pipelineID)
	assert.Contains(suite.T(), string(out), "Successfully deleted pipeline ID")
	assert.NoError(suite.T(), err, fmt.Sprintf("Error found (non expected): %v", err))
}

func TestProgramDeleteSuite(t *testing.T) {
	suite.Run(t, new(ProgramDeleteSuite))
}
