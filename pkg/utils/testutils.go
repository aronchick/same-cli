package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func ExecuteCommandC(t *testing.T, root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
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

	log.Tracef("Command to execute: same %v", root.CalledAs())

	c, err = root.ExecuteC()
	return c, buf.String(), err
}

func PrintErrorAndReturnExit(cmd *cobra.Command, s string, err error) (exit bool) {
	message := fmt.Sprintf(s, err.Error())
	cmd.Println(message)
	log.Fatalf(message)

	return os.Getenv("TEST_PASS") != ""
}

func IsSudoer() bool {
	sudoerCmd := exec.Command("/bin/bash", "-c", "timeout 2 sudo id && echo Access granted || echo Access denied")
	output, err := sudoerCmd.CombinedOutput()
	if strings.Contains(string(output), "Access granted") && err == nil {
		return true
	}

	return false
}
