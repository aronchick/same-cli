package integration_test

import (
	"fmt"
	"io/ioutil"
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

	"github.com/onsi/gomega/gbytes"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramRunSuite struct {
	suite.Suite
	rootCmd        *cobra.Command
	remoteSAMEURL  string
	logBuf         *gbytes.Buffer
	dc             infra.DependencyCheckers
	kubectlCommand string
	configFile     *os.File
	fatal          bool
	runID          string
}

// Before all suite
func (suite *ProgramRunSuite) SetupAllSuite() {
	suite.logBuf = gbytes.NewBuffer()
	suite.configFile, _ = ioutil.TempFile(os.TempDir(), "SAME-TEST-RUN-CONFIG-*.yaml")

	text, _ := ioutil.ReadFile("../testdata/config/notarget.yaml")
	if _, err := suite.configFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	if os.Getenv("GITHUB_ACTIONS") != "" {
		running, err := utils.GetUtils().IsK3sRunning(suite.rootCmd)
		if err != nil || !running {
			log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
		}
	}

	os.Setenv("TEST_PASS", "1")
}

// Before each test
func (suite *ProgramRunSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	suite.kubectlCommand = "kubectl"
	suite.fatal = false

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

	random, _ := uuid.NewRandom()
	suite.runID = random.String()
}

func (suite *ProgramRunSuite) TearDownAllSuite() {
	defer os.Remove(suite.configFile.Name())
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
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "--config", "../testdata/config/notarget.yaml")
	assert.Regexp(suite.T(), regexp.MustCompile(`required flag\(s\).+?"experiment-name".+? not set`), string(out))
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithFileAndNoKubectl() {
	os.Setenv("TEST_PASS", "1")
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/same.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml", "--", mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE)
	assert.Contains(suite.T(), string(out), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithNoKubeconfig() {
	os.Setenv("TEST_PASS", "1")
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/dev/null/baddir")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/same.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not set kubeconfig default context")

	if origKubeconfig != "" {
		_ = os.Setenv("KUBECONFIG", origKubeconfig)
	} else {
		_ = os.Unsetenv("KUBECONFIG")
	}
}

func (suite *ProgramRunSuite) Test_ExecuteWithCreateWithBadFile() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "/dev/null/same.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not find sameFile", "Attempting to start same run with a bad file did not fail (as expected).")
}

func (suite *ProgramRunSuite) Test_GetRemoteBadURL() {
	os.Setenv("TEST_PASS", "1")
	// Use a URL with a control character
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "https://\n", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "unable to parse")
}

func (suite *ProgramRunSuite) Test_GetRemoteNoSAME() {
	os.Setenv("TEST_PASS", "1")
	// The URL 'https://github.com/dapr/dapr' does not have a 'same.yaml' file in it, so it should fail
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "https://github.com/dapr/dapr", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not download SAME")

}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEWithBadPipelineFile() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/samefiles/badpipelinedirectory.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not find pipeline definition specified in SAME program")
}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEWithBadPipelineDirectory() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/samefiles/badpipelinefile.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not find pipeline definition specified in SAME program")
}

func (suite *ProgramRunSuite) Test_GetRemoteSAMEGoodPipeline() {
	os.Setenv("TEST_PASS", "1")

	_, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "run", "-f", "../testdata/samefiles/goodpipeline.yaml", "-e", "test-experiment", "-r", suite.runID, "--config", "../testdata/config/notarget.yaml")

	_ = out
	assert.NoError(suite.T(), err, fmt.Sprintf("Error found (non expected): %v", err))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProgramRunSuite(t *testing.T) {
	suite.Run(t, new(ProgramRunSuite))
}
