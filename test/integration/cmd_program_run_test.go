package integration_test

import (
	"fmt"
	"os"
	"regexp"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramRunSuite struct {
	suite.Suite
	rootCmd            *cobra.Command
	dc                 infra.DependencyCheckers
	tmpConfigDirectory string
	fatal              bool
	runID              string
}

// Before all suite
func (suite *ProgramRunSuite) SetupAllSuite() {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		running, err := utils.GetUtils().IsK3sRunning(suite.rootCmd)
		if err != nil || !running {
			log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
		}
	}

	os.Setenv("TEST_PASS", "1")

	suite.dc = &infra.LiveDependencyCheckers{}
	if ok, err := suite.dc.CanConnectToKubernetes(suite.rootCmd); !ok && (err != nil) {
		assert.Fail(suite.T(), `Cannot run tests because we cannot connect to a live cluster. Test your cluster with:  kubectl version`)
	}

	if err := suite.dc.CheckDependenciesInstalled(suite.rootCmd); err != nil {
		log.Warnf("Failed one or more dependencies - skipping this test: %v", err.Error())
		suite.T().Skip()
	}

	if ok, _ := utils.IsKFPReady(suite.rootCmd); !ok {
		log.Warn("KFP does not appear to be ready, this may cause tests to fail.")
	}

}

// Before each test
func (suite *ProgramRunSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	utils.ClearCobraArgs(suite.rootCmd)
	suite.fatal = false

	if os.Getenv("KUBECONFIG") == "" {
		os.Setenv("KUBECONFIG", os.ExpandEnv("$HOME/.kube/config"))
	}

	suite.tmpConfigDirectory = utils.GetTmpConfigDirectory("RUN")

	random, _ := uuid.NewRandom()
	suite.runID = random.String()

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }
}

func (suite *ProgramRunSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tmpConfigDirectory)
}

func (suite *ProgramRunSuite) TearDownAllSuite() {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ProgramRunSuite) Test_ExecuteWithNoCreate() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program")
	assert.Contains(suite.T(), string(out), "same program [command]")
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateAndNoArgs() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "--config", configFileName, "--", "NO-ARGS-TEST")
	assert.Regexp(suite.T(), regexp.MustCompile(`required flag\(s\).+?"experiment-name".+? not set`), string(out))
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithFileAndNoKubectl() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/same.yaml", "-e", "Test_ExecuteWithCreateWithFileAndNoKubectl-experiment", "-r", suite.runID, "--config", configFileName, "--", mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE)
	assert.Contains(suite.T(), string(out), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithNoKubeconfig() {
	os.Setenv("TEST_PASS", "1")
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/dev/null/baddir")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/same.yaml", "-e", "Test_ExecuteWithCreateWithNoKubeconfig-experiment", "-r", suite.runID, "--config", configFileName)
	assert.Contains(suite.T(), string(out), "could not set kubeconfig default context")

	if origKubeconfig != "" {
		_ = os.Setenv("KUBECONFIG", origKubeconfig)
	} else {
		_ = os.Unsetenv("KUBECONFIG")
	}
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithBadFile() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "/dev/null/same.yaml", "-e", "Test_ExecuteWithCreateWithBadFile-experiment", "-r", suite.runID, "--config", configFileName)
	assert.Contains(suite.T(), string(out), "could not find sameFile", "Attempting to start same run with a bad file did not fail (as expected).")
}

func (suite *ProgramRunSuite) Test_GetRemoteBadURL() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	// Use a URL with a control character
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "https://\n", "-e", "Test_GetRemoteBadURL-experiment", "-r", suite.runID, "--config", configFileName)
	assert.Contains(suite.T(), string(out), "unable to parse")
}

func (suite *ProgramRunSuite) Test_GetRemoteNoSAME() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")

	// The URL 'https://github.com/dapr/dapr' does not have a 'same.yaml' file in it, so it should fail
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "https://github.com/dapr/dapr", "-e", "Test_GetRemoteNoSAME-experiment", "-r", suite.runID, "--config", configFileName)
	assert.Contains(suite.T(), string(out), "could not download SAME")

}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEWithBadPipelineFile() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	sameFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/samefiles/badpipelinedirectory.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", sameFileName, "-e", "Test_GetRemoteSAMEWithBadPipelineFile-experiment", "-r", suite.runID, "--config", configFileName, "--", "BAD_PIPELINE_TEST")
	assert.Contains(suite.T(), string(out), "could not find pipeline definition specified in SAME program")
}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEWithBadPipelineDirectory() {
	os.Setenv("TEST_PASS", "1")
	configFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	sameFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/samefiles/badpipelinefile.yaml")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", sameFileName, "-e", "Test_GetRemoteSAMEWithBadPipelineDirectory-experiment", "-r", suite.runID, "--config", configFileName)
	assert.Contains(suite.T(), string(out), "could not find pipeline definition specified in SAME program")
}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEGoodPipeline() {
	os.Setenv("TEST_PASS", "1")
	configFileName, err := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/config/notarget.yaml")
	if err != nil {
		log.Warnf("Error: %v", err)
	}
	sameFileName, _ := utils.GetTmpConfigFile("RUN", suite.tmpConfigDirectory, "../testdata/samefiles/goodpipeline.yaml")

	log.Warnf("Config File: %v", configFileName)
	log.Warnf("Config File: %v", sameFileName)

	_, _ = utils.CopyFilesInDir("../testdata/pipelines", suite.tmpConfigDirectory, false)
	origDir, _ := os.Getwd()
	_ = os.Chdir(suite.tmpConfigDirectory)
	_, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", sameFileName, "-e", "Test_GetRemoteSAMEGoodPipeline-experiment", "-r", suite.runID, "--config", configFileName)

	_ = out
	assert.NoError(suite.T(), err, fmt.Sprintf("Error found (non expected): %v", err))
	_ = os.Chdir(origDir)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProgramRunSuite(t *testing.T) {
	suite.Run(t, new(ProgramRunSuite))
}
