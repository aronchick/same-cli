package cmd_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	cmd "github.com/azure-octo/same-cli/cmd"
	log "github.com/sirupsen/logrus"
)

func execute_and_read(cmd cobra.Command, args []string, expected_substring string) {

	// The Ginkgo test runner takes over os.Args and fills it with
	// its own flags.  This makes the cobra command arg parsing
	// fail because of unexpected options.  Work around this.

	// Save a copy of os.Args
	// origArgs := os.Args[:]

	// Trim os.Args to only the first arg
	os.Args = append(os.Args[:1], args...) // trim to only the first arg, which is the command itself

	r, w, _ := os.Pipe()
	tmp := os.Stdout
	defer func() {
		os.Stdout = tmp
	}()
	os.Stdout = w
	go func() {
		err := cmd.Execute()
		Expect(err).Should(BeNil())
		w.Close()
	}()

	stdout, _ := ioutil.ReadAll(r)
	// Run the command which parses os.Args

	Expect(string(stdout)).To(ContainSubstring(expected_substring))
}

var _ = Describe("same program", func() {

	var rootCmd = cmd.RootCmd
	BeforeSuite(func() {
		log.SetOutput(ioutil.Discard)
	})

	BeforeEach(func() {
	})

	Context("create", func() {

		It("Should run without arguments", func() {
			execute_and_read(*rootCmd, []string{}, "same [command]")
		})

		It("Should run program without arguments", func() {
			execute_and_read(*rootCmd, []string{"program"}, "same program [command]")
		})

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
