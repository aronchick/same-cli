package integration_test

import (
	"bytes"
	"os"

	"testing"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProgramCreateSuite struct {
	suite.Suite
	rootCmd *cobra.Command
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ProgramCreateSuite) SetupTest() {
	suite.rootCmd = cmd.RootCmd
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ProgramCreateSuite) Test_ExecuteWithNoCreate() {
	_, out, _ := executeCommandC(suite.T(), suite.rootCmd, "program")
	assert.Contains(suite.T(), string(out), "same program [command]")
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateAndNoArgs() {
	_, out, _ := executeCommandC(suite.T(), suite.rootCmd, "program", "create")
	assert.Contains(suite.T(), string(out), "required flag(s) \"file\"")
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithFileAndNoKubectl() {
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	_, out, _ := executeCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "same.yaml")
	assert.Contains(suite.T(), string(out), "Error: the 'kubectl' binary is not on your PATH")
	os.Setenv("PATH", origPath)
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithNoKubeconfig() {
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/dev/null/baddir")
	_, out, _ := executeCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "same.yaml")
	assert.Contains(suite.T(), string(out), "Could not set kubeconfig default context")
	if origKubeconfig != "" {
		_ = os.Setenv("KUBECONFIG", origKubeconfig)
	} else {
		_ = os.Unsetenv("KUBECONFIG")
	}
}

func (suite *ProgramCreateSuite) Test_ExecuteWithCreateWithPadFile() {
	_, out, _ := executeCommandC(suite.T(), suite.rootCmd, "program", "create", "-f", "/dev/null/same.yaml")
	assert.Contains(suite.T(), string(out), "could not find sameFile")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProgramCreateSuite(t *testing.T) {
	suite.Run(t, new(ProgramCreateSuite))
}

func executeCommandC(t *testing.T, root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	// Need to check if we're running in debug mode for VSCode
	// Empty them if they exist
	if (len(os.Args) > 2) && (os.Args[1] == "-test.run") {
		os.Args[1] = ""
		os.Args[2] = ""
	}

	c, err = root.ExecuteC()
	return c, buf.String(), err
}

// var _ = Describe("same program", func() {

// 	var rootCmd = cmd.RootCmd
// 	BeforeSuite(func() {
// 		log.SetOutput(ioutil.Discard)
// 		ts, err := cmdtest.Read("testdata/integration")
// 	})

// 	BeforeEach(func() {
// 	})

// 	Context("create", func() {

// 		It("Should run without arguments", func() {
// 			execute_and_read(*rootCmd, []string{}, "same [command]")
// 		})

// })
