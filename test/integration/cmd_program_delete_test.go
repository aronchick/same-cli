package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/google/uuid"
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
type ProgramDeleteSuite struct {
	suite.Suite
	rootCmd            *cobra.Command
	pipelineID         string
	pipelineName       string
	logBuf             *gbytes.Buffer
	dc                 infra.DependencyCheckers
	tmpConfigDirectory string
	origDir            string
}

// Before all suite
func (suite *ProgramDeleteSuite) SetupAllSuite() {
	os.Setenv("TEST_PASS", "1")

	if os.Getenv("TEST_K3S") == "true" {
		running, err := utils.GetUtils(suite.rootCmd, []string{}).IsK3sRunning()
		if err != nil || !running {
			assert.Fail(suite.T(), "k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
			suite.T().Skip()
		}
	}

	dc := infra.GetDependencyCheckers(suite.rootCmd, []string{})
	if err := dc.CheckDependenciesInstalled(); err != nil {
		log.Warnf("Failed one or more dependencies - skipping this test: %v", err.Error())
		suite.T().Skip()
	}
	suite.logBuf = gbytes.NewBuffer()
	suite.origDir, _ = os.Getwd()
}

func (suite *ProgramDeleteSuite) SetupTest() {
	os.Setenv("TEST_PASS", "1")
	suite.rootCmd = cmd.RootCmd
	viper.Reset()

	if os.Getenv("KUBECONFIG") == "" {
		os.Setenv("KUBECONFIG", os.ExpandEnv("$HOME/.kube/config"))
	}

	suite.dc = &infra.LiveDependencyCheckers{}
	suite.dc.SetCmd(suite.rootCmd)
	suite.dc.SetCmdArgs([]string{})
	
	if ok, err := suite.dc.CanConnectToKubernetes(); !ok && (err != nil) {
		assert.Fail(suite.T(), `Cannot run tests because we cannot connect to a live cluster. Test your cluster with:  kubectl version`)
		suite.T().Skip()
	}

	if ok, _ := utils.IsKFPReady(suite.rootCmd); !ok {
		assert.Fail(suite.T(), "KFP does not appear to be ready, this may cause tests to fail.")
		suite.T().Skip()
	}

	suite.tmpConfigDirectory = utils.GetTmpConfigDirectory("DELETE")
	tmpConfigFile, _ := utils.GetTmpConfigFile("DELETE", suite.tmpConfigDirectory, "../testdata/samefiles/deletepipeline.yaml")

	_, err := utils.CopyFilesInDir("../testdata/pipelines", suite.tmpConfigDirectory, false)
	if err != nil {
		assert.Fail(suite.T(), "could not copy pipeline files into temp directory: %v", err.Error())
		suite.T().Skip()
	}

	// if err = os.Chdir(suite.tmpConfigDirectory); err != nil {
	// 	log.Warnf("could not switch to the temp directory for execution: %v", suite.tmpConfigDirectory)
	// }

	runID, _ := uuid.NewRandom()
	commandString := fmt.Sprintf("same program run -f %v -e DELETE_SUITE-test-experiment -r %v", tmpConfigFile, runID.String())
	out, err := exec.Command("/bin/bash", "-c", commandString).CombinedOutput()
	if err != nil {
		message := fmt.Errorf("could not start delete test suite because we could not create a sample pipeline: %v", err)
		suite.rootCmd.Print(message)
		log.Warn(message.Error())
		suite.T().Skip()
	}

	r := regexp.MustCompile(`Name:\s+([^\n]+)\nVersionID:\s+([^\n]+)`)
	outputString := string(out)
	rs := r.FindStringSubmatch(outputString)
	if len(rs) < 2 {
		assert.Fail(suite.T(), "cmd_program_delete_test: during setup, could not find name and ID in the returned upload string: %v", outputString)
		suite.T().Skip()
	}
	suite.rootCmd.Printf("%#v\n", rs[1])
	suite.rootCmd.Printf("%#v\n", rs[2])
	suite.pipelineName = rs[1]
	suite.pipelineID = rs[2]

}

func (suite *ProgramDeleteSuite) TearDownAllSuite() {
	// os.RemoveAll(suite.tmpConfigDirectory)
	log.Warnf("Directory: %v", suite.tmpConfigDirectory)
}
func (suite *ProgramDeleteSuite) Test_DeletePipeline() {
	os.Setenv("TEST_PASS", "1")

	// https://github.com/azure-octo/same-cli/issues/91
	assert.True(suite.T(), true)
}

func TestProgramDeleteSuite(t *testing.T) {
	suite.Run(t, new(ProgramDeleteSuite))
}
