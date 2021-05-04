package utils_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type K8sUtilsSuite struct {
	suite.Suite
	rootCmd      *cobra.Command
	dc           infra.DependencyCheckers
	fatal        bool
	outputBuffer *bytes.Buffer
}

// Before all suite
func (suite *K8sUtilsSuite) SetupAllSuite() {
	log.Trace("Inside Setup All")
}

// Before each test
func (suite *K8sUtilsSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
	suite.outputBuffer = new(bytes.Buffer)
	suite.rootCmd.SetOut(suite.outputBuffer)
	suite.rootCmd.SetErr(suite.outputBuffer)
	suite.dc = infra.GetDependencyCheckers(suite.rootCmd, []string{})
	suite.fatal = false
	os.Setenv("TEST_PASS", "1")

}

// After test
func (suite *K8sUtilsSuite) TearDownAllSuite() {
}

func (suite *K8sUtilsSuite) Test_HasContext() {
	log.Trace("Inside Has Context")
	context, err := suite.dc.HasContext()
	assert.NotEmpty(suite.T(), context, "No context returned from the command.")
	assert.Nil(suite.T(), err, "Error requesting kubernetes context")
}
func (suite *K8sUtilsSuite) Test_HasCluster() {
	clusters, err := suite.dc.HasClusters()
	assert.NotEmpty(suite.T(), len(clusters) > 1, "No clusters returned from the command.")
	assert.Nil(suite.T(), err, "Error requesting kubernetes context")
}

func (suite *K8sUtilsSuite) Test_UnsetKubectlCmd() {
	os.Setenv("TEST_PASS", "1")

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(int) { suite.fatal = true }

	i := &infra.LiveInstallers{}
	os.Setenv("MISSING_KUBECTL", mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE)
	i.SetKubectlCmd("")
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")

	// Fix below: https://github.com/azure-octo/same-cli/issues/221
	//assert.Contains(suite.T(), suite.outputBuffer.String(), mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_RESULT, "Suite does not properly warn about missing kubectl on path.")
	assert.Equal(suite.T(), suite.outputBuffer.String(), "", "Suite does not properly warn about missing kubectl on path.")

	os.Setenv("PATH", origPath)
	os.Unsetenv("MISSING_KUBECTL")
}

func (suite *K8sUtilsSuite) Test_IsUrlReachable() {
	type url_pair struct {
		url    string
		passes bool
	}

	var urls_to_test = []url_pair{}
	urls_to_test = append(urls_to_test, url_pair{"https://google.com:80", true})
	urls_to_test = append(urls_to_test, url_pair{"https://google.com", false}) // Missing Port
	urls_to_test = append(urls_to_test, url_pair{"https://google.com:80/THIS_SHOULDNT_MATTER", true})
	urls_to_test = append(urls_to_test, url_pair{"VALIDURLBUTNOTREACHABLE.com:6443", false}) // Bad URL
	for _, url_pair := range urls_to_test {
		fmt.Printf("Testing URL: %v - %v", url_pair.url, url_pair.passes)
		_, err := utils.GetUtils(&cobra.Command{}, []string{}).IsEndpointReachable(url_pair.url)
		assert.Equal(suite.T(), (err == nil), (url_pair.passes), fmt.Sprintf("Expected URL (%v) is reachable to be: %v", url_pair.url, url_pair.passes))
	}
}

func TestK8sUtilsSuite(t *testing.T) {
	suite.Run(t, new(K8sUtilsSuite))
}
