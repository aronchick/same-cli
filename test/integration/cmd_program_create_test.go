package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/onsi/gomega/gbytes"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramCreateSuite struct {
	suite.Suite
	rootCmd        *cobra.Command
	remoteSAMEURL  string
	logBuf         *gbytes.Buffer
	kubectlCommand string
	configFile     *os.File
	fatal          bool
}

// Before all suite
func (suite *ProgramCreateSuite) SetupAllSuite() {
	suite.logBuf = gbytes.NewBuffer()
	suite.configFile, _ = ioutil.TempFile(os.TempDir(), "SAME-TEST-RUN-CONFIG-*.yaml")

	text, _ := ioutil.ReadFile("../testdata/config/notarget.yaml")
	if _, err := suite.configFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	if os.Getenv("TEST_K3S") == "true" {
		running, err := utils.GetUtils().K3sRunning(suite.rootCmd)
		if err != nil || !running {
			log.Fatal("k3s does not appear to be installed, required for testing. Please run 'sudo same installK3s'")
		}
	}

	os.Setenv("TEST_PASS", "1")
}

// Before each test
func (suite *ProgramCreateSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.remoteSAMEURL = "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo"
	suite.kubectlCommand = "kubectl"
	suite.fatal = false

	if ok, _ := utils.KFPReady(suite.rootCmd); !ok {
		log.Warn("KFP does not appear to be ready, this may cause tests to fail.")
	}
}

func (suite *ProgramCreateSuite) TearDownAllSuite() {
	defer os.Remove(suite.configFile.Name())
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ProgramCreateSuite) Test_ExecuteWithNoCreate() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program")
	assert.Contains(suite.T(), string(out), "same program [command]")
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateAndNoArgs() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "required flag(s) \"file\"")

}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithFileAndNoKubectl() {
	os.Setenv("TEST_PASS", "1")
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/same.yaml", "--config", "../testdata/config/notarget.yaml", "--", mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE)
	assert.Contains(suite.T(), string(out), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT)
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithNoKubeconfig() {
	os.Setenv("TEST_PASS", "1")
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/dev/null/baddir")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/same.yaml", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "Could not set kubeconfig default context")

	if origKubeconfig != "" {
		_ = os.Setenv("KUBECONFIG", origKubeconfig)
	} else {
		_ = os.Unsetenv("KUBECONFIG")
	}
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithBadFile() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "/dev/null/same.yaml", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not find sameFile", "Attempting to start same create with a bad file did not fail (as expected).")

}

func (suite *ProgramCreateSuite) Test_GetRemoteBadURL() {
	os.Setenv("TEST_PASS", "1")
	// Use a URL with a control character
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "https://\n", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "unable to parse")
}

func (suite *ProgramCreateSuite) Test_GetRemoteNoSAME() {
	os.Setenv("TEST_PASS", "1")
	// The URL 'https://github.com/dapr/dapr' does not have a 'same.yaml' file in it, so it should fail
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "https://github.com/dapr/dapr", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "could not download SAME")

}

func (suite *ProgramCreateSuite) Test_GetRemoteSAMEWithBadPipelineFile() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/badpipelinedirectory.yaml", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "/dev/null/bad_pipeline.tgz: not a directory")
}

func (suite *ProgramCreateSuite) Test_GetRemoteSAMEWithBadPipelineDirectory() {
	os.Setenv("TEST_PASS", "1")
	_, out, _ := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/badpipelinefile.yaml", "--config", "../testdata/config/notarget.yaml")
	assert.Contains(suite.T(), string(out), "/tmp/bad_pipeline.tgz: no such file or directory")
}

func (suite *ProgramCreateSuite) Test_GetRemoteSAMEGoodPipeline() {
	os.Setenv("TEST_PASS", "1")

	_, out, err := utils.ExecuteCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "../testdata/samefiles/goodpipeline.yaml", "--config", "../testdata/config/notarget.yaml")

	_ = out
	assert.NoError(suite.T(), err, fmt.Sprintf("Error found (non expected): %v", err))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProgramCreateSuite(t *testing.T) {
	suite.Run(t, new(ProgramCreateSuite))
}
