package infra

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type LiveInstallers struct {
	_cmdArgs        []string
	_kubectlCommand string
}

func (li *LiveInstallers) InstallK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	log.SetLevel(log.TraceLevel)
	log.Tracef("Inside Installer: %v", cmd.Short)
	executingUser, err := user.LookupId(os.Getenv("SUDO_UID"))

	if err != nil {
		log.Tracef("Error: %v", err)
		u, _ := user.Current()
		log.Tracef("Current UID: %v", u.Username)
		log.Tracef("Current SUDO_UID: %v", os.Getenv("SUDO_UID"))
		log.Fatalf("We only support running this command under sudo. Please reexecute.")
	}

	// For some reason, go doesn't want to look up the username based
	// on the UID. So, just grabbing the SUDO_USER for now.
	executingUserName := os.Getenv("SUDO_USER")

	_, err = exec.LookPath("systemd")
	if err != nil {
		log.Fatalf("We only currently support this command on systems running systemd. Please read more about installing k3s on your machine here: https://rancher.com/docs/k3s/latest/en/advanced/")
	}

	log.Tracef("Executing User Home: %v", executingUser.HomeDir)
	k3sCommand, detectErr := li.DetectK3s("k3s")
	log.Tracef("k3sCommand: %v", k3sCommand)
	log.Tracef("Error: %v", detectErr)

	if (k3sCommand == "") || (detectErr != nil && strings.Contains(detectErr.Error(), "file not found")) {
		log.Tracef("SUDO_USER: %v", os.Getenv("SUDO_USER"))
		log.Tracef("Executing User Name: %v", executingUserName)

		uid, err1 := strconv.Atoi(executingUser.Uid)
		gid, err2 := strconv.Atoi(executingUser.Gid)
		if err1 != nil || err2 != nil {
			log.Fatalf("'%v' or '%v' could not be converted to int from strconv.Atoi: \n1: %v\n2: %v", executingUser.Uid, executingUser.Gid, err1, err2)
		}

		kubeDir := path.Join(executingUser.HomeDir, ".kube")
		if _, err := os.Stat(kubeDir); os.IsNotExist(err) {
			logrus.Tracef("%v does not exist, creating it now.", kubeDir)
			mkDirErr := os.Mkdir(kubeDir, 0700)
			if mkDirErr != nil {
				log.Fatalf("Unable to create the .kube directory at: %v", mkDirErr)
			}
			chownErr := os.Chown(kubeDir, uid, gid)
			if chownErr != nil {
				log.Fatalf("Unable to change %v to be owned by the current user.", kubeDir)
			}
		}

		userKubeConfigLocation := fmt.Sprintf("%v/.kube/config", executingUser.HomeDir)

		// defer func() { _ = os.Remove(tmpK3sConfig.Name()) }()
		// defer func() {
		// 	backupFileName := fmt.Sprintf("%v.bak", userKubeConfigLocation)
		// 	_ = os.Remove(backupFileName)
		// }()

		k3sDownloadAndInstallURL := "curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644"
		k3sDownloadAndInstallScript := fmt.Sprintf(`
#!/bin/bash
set -e
%v
`, k3sDownloadAndInstallURL)

		cmd.Printf("About to execute the following:\n%v\n", k3sDownloadAndInstallScript)

		// I wonder if the below code is right - you could remove the if statement, but then it's slightly less
		// readable, assuming that a user has to deduce that returning err could be nil
		if err := utils.ExecuteInlineBashScript(cmd, k3sDownloadAndInstallScript, "K3s package failed to download and install."); err != nil {
			return "", err
		}

		kubeConfigs := []string{}
		_, err = os.Stat(userKubeConfigLocation)
		log.Tracef("SUDO_USER: %v", os.Getenv("SUDO_USER"))
		if err == nil || errors.Is(err, os.ErrExist) {
			backupConfigCommand := fmt.Sprintf("cp -rf %v %v.bak", userKubeConfigLocation, userKubeConfigLocation)
			log.Tracef("backupConfigCmd:\n%v\n", backupConfigCommand)

			backupConfigScript := fmt.Sprintf(`
	#!/bin/bash
	set -e
	runuser -l %v -c "%v"
			`, executingUserName, backupConfigCommand)

			cmd.Printf("About to execute the following:\n%v\n", backupConfigScript) // I wonder if the below code is right - you could remove the if statement, but then it's slightly less
			// readable, assuming that a user has to deduce that returning err could be nil
			if err := utils.ExecuteInlineBashScript(cmd, backupConfigScript, "Kubeconfig failed to backup."); err != nil {
				return "", err
			}

			kubeConfigs = append(kubeConfigs, userKubeConfigLocation)
		}

		// Create our Temp File:  This will create a filename like /tmp/prefix-123456
		// We can use a pattern of "pre-*.txt" to get an extension like: /tmp/pre-123456.txt
		tmpFile, err := ioutil.TempFile(os.TempDir(), "K3S_CONFIG_TEMP-")
		if err != nil {
			log.Fatal("Cannot create temporary file for merging", err)
		}

		// Remember to clean up the file afterwards
		defer os.Remove(tmpFile.Name())

		log.Tracef("Created File: %v", tmpFile.Name())
		_ = tmpFile.Chmod(0666)

		k3s_default_config, err := ioutil.ReadFile("/etc/rancher/k3s/k3s.yaml")

		if err != nil {
			log.Fatal("Unable to read from /etc/rancher/k3s/k3s.yaml", err)
		}
		if _, err = tmpFile.Write(k3s_default_config); err != nil {
			log.Fatal("Failed to write to temporary file", err)
		}

		log.Tracef("Wrote to: %v", tmpFile.Name())

		// Close the file
		if err := tmpFile.Close(); err != nil {
			log.Fatal(err)
		}

		kubeConfigs = append(kubeConfigs, tmpFile.Name())
		mergeAndFlattenCommand := fmt.Sprintf("KUBECONFIG=%v kubectl config view --flatten > %v/.kube/config", strings.Join(kubeConfigs, ":"), executingUser.HomeDir)
		log.Tracef("mergeAndFlattenCommand:\n%v\n", mergeAndFlattenCommand)

		k3sMergeScript := fmt.Sprintf(`
#!/bin/bash
set -e
runuser -l %v -c "%v"
export KUBECONFIG=$HOME/.kube/config
		`, executingUserName, mergeAndFlattenCommand)

		cmd.Printf("About to execute the following:\n%v\n", k3sMergeScript)
		// I wonder if the below code is right - you could remove the if statement, but then it's slightly less
		// readable, assuming that a user has to deduce that returning err could be nil
		if err := utils.ExecuteInlineBashScript(cmd, k3sMergeScript, "K3s failed merge configs."); err != nil {
			return "", err
		}
	} else if detectErr != nil {
		return "", fmt.Errorf("error looking for K3s in PATH: %v", err)
	}

	log.Trace("Finished installing, detecting again.")

	return "k3s", li.PostInstallK3sRunning(cmd)
}

func (i *LiveInstallers) PostInstallK3sRunning(cmd *cobra.Command) (err error) {

	cmd.Println("Waiting up to 120 seconds for k3s to become ready...")
	elapsedTime := 0
	for {
		cmd.Printf("%v...", elapsedTime)
		if isRunning, _ := utils.GetUtils().K3sRunning(cmd); isRunning {
			cmd.Println("k3s is running locally.")
			return nil
		}
		if elapsedTime >= 120 {
			message := "k3s did not start in a timely manner."
			cmd.Println(message)
			return fmt.Errorf(message)

		}
		time.Sleep(5 * time.Second)

		// Printed after sleep is over
		elapsedTime += 5

	}

}

func (i *LiveInstallers) DetectK3s(s string) (string, error) {
	log.Infof("Executing detect on: %v", s)
	return exec.LookPath(s)
}

func (i *LiveInstallers) InstallKFP(cmd *cobra.Command) (err error) {

	log.Tracef("Inside InstallKFP()")
	kubectlCommand := i.GetKubectlCmd()
	log.Tracef("kubectlCommand: %v\n", kubectlCommand)
	kfpInstall := fmt.Sprintf(`
	#!/bin/bash
	set -e
	export PIPELINE_VERSION=1.4.1
	export KUBECTL_COMMAND=%v
	$KUBECTL_COMMAND create namespace kubeflow || true
	$KUBECTL_COMMAND config set-context --current --namespace=kubeflow || true 
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/cluster-scoped-resources?ref=$PIPELINE_VERSION"
	$KUBECTL_COMMAND wait --for condition=established --timeout=60s crd/applications.app.k8s.io
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/env/platform-agnostic-pns?ref=$PIPELINE_VERSION"
	`, kubectlCommand)

	log.Tracef("About to execute: %v\n", kfpInstall)
	if err := utils.ExecuteInlineBashScript(cmd, kfpInstall, "KFP failed to install."); err != nil {
		log.Tracef("Error executing: %v\n", err.Error())
		return err
	}

	return nil
}

func (i *LiveInstallers) SetKubectlCmd(s string) {
	i._kubectlCommand = s
}

func (i *LiveInstallers) GetKubectlCmd() string {
	return i._kubectlCommand
}

func (i *LiveInstallers) GetCmdArgs() []string {
	return i._cmdArgs
}

func (i *LiveInstallers) SetCmdArgs(args []string) {
	i._cmdArgs = args
}
