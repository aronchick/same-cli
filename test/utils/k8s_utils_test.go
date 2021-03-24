package utils_test

import (
	"testing"

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
	rootCmd *cobra.Command
}

// Before all suite
func (suite *K8sUtilsSuite) SetupAllSuite() {
	log.Trace("Inside Setup All")
}

// Before each test
func (suite *K8sUtilsSuite) SetupTest() {
}

// After test
func (suite *K8sUtilsSuite) TearDownAllSuite() {
}

func (suite *K8sUtilsSuite) Test_HasContext() {
	log.Trace("Inside Has Context")
	context, err := utils.HasContext(suite.rootCmd)
	assert.NotEmpty(suite.T(), context, "No context returned from the command.")
	assert.Nil(suite.T(), err, "Error requesting kubernetes context")
}
func (suite *K8sUtilsSuite) Test_HasCluster() {
	clusters, err := utils.HasClusters(suite.rootCmd)
	assert.NotEmpty(suite.T(), len(clusters) > 1, "No clusters returned from the command.")
	assert.Nil(suite.T(), err, "Error requesting kubernetes context")
}

// COMMENTING OUT TEST UNTIL UTILS.MOCKS COMPLETE - may not be necessary at all
// func (suite *K8sUtilsSuite) Test_K3sRunning() {
// 	if os.Getenv("GITHUB_ACTIONS") != "" {
// 		suite.T().Skip()
// 	}

// 	running, err := utils.GetUtils().K3sRunning(suite.rootCmd)
// 	assert.True(suite.T(), running, "K3s is not running.")
// 	assert.Nil(suite.T(), err, "Error requesting testing for k3s cluster: %v", err)
// }

func TestK8sUtilsSuite(t *testing.T) {
	suite.Run(t, new(K8sUtilsSuite))
}
