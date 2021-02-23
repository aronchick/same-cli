package utils_test

import (
	"testing"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type nilfn func(t assert.TestingT, object interface{}, msgAndArgs ...interface{}) bool

func isGoodURL(t *testing.T, thisUrl string, expectedValue bool, fn nilfn) {
	isRemote, err := utils.IsRemoteFilePath(thisUrl)
	if expectedValue {
		assert.True(t, isRemote)
	} else {
		assert.False(t, isRemote)
	}

	// Use the function to test if nil or not
	fn(t, err)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *UtilsSuite) Test_RemoteGithub() {
	isGoodURL(suite.T(), "http://github.com", true, assert.Nil)
}

func (suite *UtilsSuite) Test_GitHubURLOrgNoRepo() {
	isGoodURL(suite.T(), "http://github.com/contoso", true, assert.Nil)
}
func (suite *UtilsSuite) Test_GitHubURLOrgRepo() {
	isGoodURL(suite.T(), "http://github.com/contoso/sameple-repo", true, assert.Nil)
}
func (suite *UtilsSuite) Test_GitHubURLOrgRepoFile() {
	isGoodURL(suite.T(), "http://github.com/contoso/sameple-repo/same.yaml", true, assert.Nil)
}
func (suite *UtilsSuite) Test_NoSchema() {
	// "a URL should start with a scheme (e.g. http://, https:// or git://"
	isGoodURL(suite.T(), "github.com/contoso/sameple-repo/same.yaml", false, assert.Nil)
}
func (suite *UtilsSuite) Test_BadlyFormed() {
	isGoodURL(suite.T(), "github/contoso/sample-repo/same.yaml", false, assert.Nil)
}
func (suite *UtilsSuite) Test_LocalRelative() {
	isGoodURL(suite.T(), "../abc.txt", false, assert.Nil)
}
func (suite *UtilsSuite) Test_LocalAbsolute() {
	isGoodURL(suite.T(), "/ab/c.txt", false, assert.Nil)
}
func (suite *UtilsSuite) Test_LocalNoPath() {
	isGoodURL(suite.T(), "c.txt", false, assert.Nil)
}

func (suite *UtilsSuite) Test_FileIsNotNil() {
	isGoodURL(suite.T(), "", false, assert.NotNil)
}

func (suite *UtilsSuite) Test_FileIsParseable() {
	// Put a control character in the URL
	isGoodURL(suite.T(), "\n", false, assert.NotNil)
}
