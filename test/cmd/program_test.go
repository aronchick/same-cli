package cmd_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"

	"github.com/azure-octo/same-cli/cmd"
	log "github.com/sirupsen/logrus"

	"github.com/onsi/gomega/gbytes"
)

// func emptyRun(*Command, []string) {}

// func executeCommand(root *Command, args ...string) (output string, err error) {
// 	_, output, err = executeCommandC(root, args...)
// 	return output, err
// }

// func executeCommandWithContext(ctx context.Context, root *Command, args ...string) (output string, err error) {
// 	buf := new(bytes.Buffer)
// 	root.SetOut(buf)
// 	root.SetErr(buf)
// 	root.SetArgs(args)

// 	err = root.ExecuteContext(ctx)

// 	return buf.String(), err
// }

// func executeCommandC(root *Command, args ...string) (c *Command, output string, err error) {
// 	buf := new(bytes.Buffer)
// 	root.SetOut(buf)
// 	root.SetErr(buf)
// 	root.SetArgs(args)

// 	c, err = root.ExecuteC()

// 	return c, buf.String(), err
// }

// func resetCommandLineFlagSet() {
// 	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
// }

// func checkStringContains(t *testing.T, got, expected string) {
// 	if !strings.Contains(got, expected) {
// 		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
// 	}
// }

// func checkStringOmits(t *testing.T, got, expected string) {
// 	if strings.Contains(got, expected) {
// 		t.Errorf("Expected to not contain: \n %v\nGot: %v", expected, got)
// 	}
// }

// const onetwo = "one two"

// func TestSingleCommand(t *testing.T) {
// 	var rootCmdArgs []string
// 	rootCmd := &Command{
// 		Use:  "root",
// 		Args: ExactArgs(2),
// 		Run:  func(_ *Command, args []string) { rootCmdArgs = args },
// 	}
// 	aCmd := &Command{Use: "a", Args: NoArgs, Run: emptyRun}
// 	bCmd := &Command{Use: "b", Args: NoArgs, Run: emptyRun}
// 	rootCmd.AddCommand(aCmd, bCmd)

// 	output, err := executeCommand(rootCmd, "one", "two")
// 	if output != "" {
// 		t.Errorf("Unexpected output: %v", output)
// 	}
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}

// 	got := strings.Join(rootCmdArgs, " ")
// 	if got != onetwo {
// 		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
// 	}
// }

// func TestChildCommand(t *testing.T) {
// 	var child1CmdArgs []string
// 	rootCmd := &Command{Use: "root", Args: NoArgs, Run: emptyRun}
// 	child1Cmd := &Command{
// 		Use:  "child1",
// 		Args: ExactArgs(2),
// 		Run:  func(_ *Command, args []string) { child1CmdArgs = args },
// 	}
// 	child2Cmd := &Command{Use: "child2", Args: NoArgs, Run: emptyRun}
// 	rootCmd.AddCommand(child1Cmd, child2Cmd)

// 	output, err := executeCommand(rootCmd, "child1", "one", "two")
// 	if output != "" {
// 		t.Errorf("Unexpected output: %v", output)
// 	}
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}

// 	got := strings.Join(child1CmdArgs, " ")
// 	if got != onetwo {
// 		t.Errorf("child1CmdArgs expected: %q, got: %q", onetwo, got)
// 	}
// }

var _ = Describe("same program", func() {

	createProgramCmd := cmd.CreateProgramCmd

	BeforeSuite(func() {
		log.SetOutput(ioutil.Discard)
	})

	AfterSuite(func() {
	})

	BeforeEach(func() {
	})

	Context("create", func() {

		It("Should run without arguments", func() {
			out := gbytes.NewBuffer()
			createProgramCmd.SetOut(out)
			createProgramCmd.SetErr(out)
			createProgramCmd.SetArgs([]string{})
			c, err := createProgramCmd.ExecuteC()
			a := c
			b := string(out.Contents())
			d := err
			_ = a
			_ = b
			_ = d
		})

		// It("Should run program without arguments", func() {
		// 	execute_and_read(*rootCmd, []string{"program"}, "same program [command]")
		// })

		// It("Should notify about missing kubectl", func() {
		// 	// Erase everyting in PATH
		// 	os.Setenv("PATH", "")

		// 	execute_and_read(*rootCmd, []string{"program", "create", "-f", ""}, "same program create [flags]")
		// })

		// It("Should read a config file from local disk", func() {
		// 	wd, _ := os.Getwd()
		// 	log.Info(wd)
		// 	execute_and_read(*rootCmd, []string{"program", "create", "-f", string(wd)}, "same program create [flags]")
		// })

		// func Test_ExecuteCommand(t *testing.T) {
		// 	cmd := NewRootCmd()
		// 	b := bytes.NewBufferString("")
		// 	cmd.SetOut(b)
		// 	cmd.SetArgs([]string{"--in", "testisawesome"})
		// 	cmd.Execute()
		// 	out, err := ioutil.ReadAll(b)
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	if string(out) != "testisawesome" {
		// 		t.Fatalf("expected \"%s\" got \"%s\"", "testisawesome", string(out))
		// 	}
		// }
		// It("a GH url with no org or repo", func() {
		// 	Expect(utils.IsRemoteFilePath("http://github.com")).Should(Equal(true))
		// })
		// It("a GH url with org but no repo", func() {
		// 	Expect(utils.IsRemoteFilePath("http://github.com/contoso")).Should(Equal(true))
		// })
		// It("a GH url with org and repo", func() {
		// 	Expect(utils.IsRemoteFilePath("http://github.com/contoso/sameple-repo")).Should(Equal(true))
		// })
		// It("a GH url with org and repo and file", func() {
		// 	Expect(utils.IsRemoteFilePath("http://github.com/contoso/sameple-repo/same.yaml")).Should(Equal(true))
		// })
		// It("a URL should start with a scheme (e.g. http://, https:// or git://", func() {
		// 	badURL("github.com/contoso/sameple-repo/same.yaml")
		// })
		// It("a badly formed URL", func() {
		// 	badURL("github/contoso/sample-repo/same.yaml")
		// })
		// It("a local relative file", func() {
		// 	badURL("../abc.txt")
		// })
		// It("a local absolute file", func() {
		// 	badURL("/ab/c.txt")
		// })
		// It("a file with no path", func() {
		// 	badURL("c.txt")
	})
})
