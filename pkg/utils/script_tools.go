package utils

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func ExecuteInlineBashScript(cmd *cobra.Command, SCRIPT string, errorMessage string) error {
	scriptCMD := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' | bash -s --", SCRIPT))
	outPipe, err := scriptCMD.StdoutPipe()
	errPipe, _ := scriptCMD.StderrPipe()
	if err != nil {
		cmd.Println(errorMessage)
		return err
	}
	err = scriptCMD.Start()

	if err != nil {
		cmd.Println(errorMessage)
		return err
	}
	errScanner := bufio.NewScanner(errPipe)
	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		m := scanner.Text()
		cmd.Println(m)
	}
	err = scriptCMD.Wait()

	if err != nil {
		for errScanner.Scan() {
			m := errScanner.Text()
			cmd.Println(m)
		}
		cmd.Println(errorMessage)
		return err
	}
	return nil
}
