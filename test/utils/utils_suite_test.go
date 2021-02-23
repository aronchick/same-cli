package utils_test

import (
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type UtilsSuite struct {
	suite.Suite
}

// Before all suite
func (suite *UtilsSuite) SetupAllSuite() {
}

// Before each test
func (suite *UtilsSuite) SetupTest() {
	log.SetOutput(ioutil.Discard)
}

// After test
func (suite *UtilsSuite) TearDownAllSuite() {
}

func TestUtilsSuite(t *testing.T) {
	suite.Run(t, new(UtilsSuite))
}
