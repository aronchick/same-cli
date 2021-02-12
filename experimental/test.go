package main

// import (
// 	"bufio"
// 	"fmt"
// 	"os/exec"
// )

// func main() {
// 	testLogin := `
// 	#!/bin/bash
// 	set -e
// 	export CURRENT_LOGIN=` + "`" + `az account show -o json | jq '\''"\(.name) : \(.id)"'\''` + "`" + `
// 	echo "You are logged in with the following credentials: $CURRENT_LOGIN"
// 	echo "If this is not correct, please execute:"
// 	echo "az account list -o json | jq '\''.[] | \"\(.name) : \(.id)\"'\''"
// 	echo "az account set --subscription REPLACE_WITH_YOUR_SUBSCRIPTION_ID"
// 	`

// 	// export CURRENT_LOGIN=$(az account show -ojson | jq '"\(.name) : \(.id)"')
// 	// echo "You are logged in with the following credentials: $CURRENT_LOGIN"
// 	// echo "If this is not correct, please execute:"
// 	// echo "az account list -o json | jq '.[] | \"\(.name) : \(.id)\"'"
// 	// echo "az account set --subscription REPLACE_WITH_YOUR_SUBSCRIPTION_ID"
// 	// `
// 	// testLogin = `
// 	// #!/bin/bash
// 	// set -e
// 	// echo "Installing Blob Storage Driver"
// 	// `

// 	// testLoginReplaced := strings.ReplaceAll(testLogin, "BACKTICK", "`")
// 	fmt.Println(testLogin)

// 	err := executeInlineBashScript(testLogin, "Your account does not appear to be logged into Azure. Please execute `az login` to authorize this account.")
// 	if err != nil {
// 		fmt.Printf("Error: %v", err)
// 	}
// }

// func executeInlineBashScript(SCRIPT string, errorMessage string) error {
// 	scriptCMD := exec.Command("/bin/bash", "-c", fmt.Sprintf("echo '%s' | bash -s --", SCRIPT))
// 	outPipe, err := scriptCMD.StdoutPipe()
// 	errPipe, _ := scriptCMD.StderrPipe()
// 	if err != nil {
// 		fmt.Printf("Could not create the commmand with the following message: %v\n", err)
// 		return err
// 	}
// 	err = scriptCMD.Start()

// 	if err != nil {
// 		fmt.Printf("Could not start the command with the following message: %v\n", err)
// 		return err
// 	}
// 	errScanner := bufio.NewScanner(errPipe)
// 	scanner := bufio.NewScanner(outPipe)
// 	for scanner.Scan() {
// 		m := scanner.Text()
// 		fmt.Println(m)
// 	}
// 	err = scriptCMD.Wait()

// 	if err != nil {
// 		for errScanner.Scan() {
// 			m := errScanner.Text()
// 			println(m)
// 		}
// 		fmt.Printf("Error while waiting the commmand with the following message: %v\n", err)
// 		return err
// 	}
// 	return nil
// }
