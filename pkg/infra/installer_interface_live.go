package infra

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"

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
		log.Tracef("Current UID: %v", os.Getenv("SUDO_UID"))
		log.Tracef("Error: %v", err)
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
	log.Tracef("Executing User Name: %v", executingUserName)
	k3sCommand, detectErr := li.DetectK3s("k3s")
	log.Tracef("k3sCommand: %v", k3sCommand)
	log.Tracef("Error: %v", detectErr)

	if (k3sCommand == "") || (detectErr != nil && strings.Contains(detectErr.Error(), "file not found")) {
		tmpMergeConfig, err := ioutil.TempFile(os.TempDir(), "KUBECONFIG_MERGE")
		if err != nil {
			log.Fatalf("Error creating temp kubeconfig merge file.: %v", err)
		}
		log.Tracef("SUDO_USER: %v", os.Getenv("SUDO_USER"))
		log.Tracef("Executing User Name: %v", executingUserName)

		var openMode os.FileMode = 0644
		err = tmpMergeConfig.Chmod(openMode)
		if err != nil {
			log.Fatalf("Error changing perms of temp file to read write user: %v", err)
		}

		uid, err1 := strconv.Atoi(executingUser.Uid)
		gid, err2 := strconv.Atoi(executingUser.Gid)
		if err1 != nil || err2 != nil {
			log.Fatalf("'%v' or '%v' could not be converted to int from strconv.Atoi: \n1: %v\n2: %v", executingUser.Uid, executingUser.Gid, err1, err2)
		}
		err = os.Chown(tmpMergeConfig.Name(), uid, gid)
		if err != nil {
			log.Fatalf("could not change ownership on file '%v': %v", tmpMergeConfig.Name(), err)
		}
		log.Tracef("Merge file: %v", tmpMergeConfig.Name())
		log.Tracef("Executing User Name: %v", executingUserName)

		tmpK3sConfig, _ := ioutil.TempFile(os.TempDir(), "K3SCONFIG")
		userKubeConfigLocation := fmt.Sprintf("%v/.kube/config", executingUser.HomeDir)

		defer func() { _ = os.Remove(tmpK3sConfig.Name()) }()
		defer func() {
			backupFileName := fmt.Sprintf("%v.bak", userKubeConfigLocation)
			_ = os.Remove(backupFileName)
		}()

		_ = "https://github.com/rancher/k3s/releases/download/v1.19.2%2Bk3s1/k3s"
		k3sDownloadAndInstallURL := "curl -sfL https://get.k3s.io | sh -"
		k3sDownloadAndInstallScript := fmt.Sprintf(`
#!/bin/bash
set -e
%v
yes | cp -rf /etc/rancher/k3s/k3s.yaml %v
chmod 0777 %v
`, k3sDownloadAndInstallURL, tmpK3sConfig.Name(), tmpK3sConfig.Name())

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
	sudo su %v
	%v
			`, executingUserName, backupConfigCommand)

			cmd.Printf("About to execute the following:\n%v\n", backupConfigScript) // I wonder if the below code is right - you could remove the if statement, but then it's slightly less
			// readable, assuming that a user has to deduce that returning err could be nil
			if err := utils.ExecuteInlineBashScript(cmd, backupConfigScript, "Kubeconfig failed to backup."); err != nil {
				return "", err
			}

			kubeConfigs = append(kubeConfigs, userKubeConfigLocation)
		}

		kubeConfigs = append(kubeConfigs, tmpK3sConfig.Name())
		mergeAndFlattenCommand := fmt.Sprintf("KUBECONFIG=%v kubectl config view --flatten > %v/.kube/config", strings.Join(kubeConfigs, ":"), executingUser.HomeDir)
		log.Tracef("mergeAndFlattenCommand:\n%v\n", mergeAndFlattenCommand)

		k3sMergeScript := fmt.Sprintf(`
#!/bin/bash
set -e
sudo su %v
%v 
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

	return li.DetectK3s("k3s")
}

func (li *LiveInstallers) StartK3s(cmd *cobra.Command) (k3sCommand string, err error) {
	k3sCommand, err = li.DetectK3s("k3s")
	log.Infof("finished detecting:\nCommand: %v\nError: %v\n", k3sCommand, err)
	if err != nil {
		return "", fmt.Errorf("error looking for K3s in PATH")
	} else if k3sCommand == "" {
		k3sStartScript := `
		#!/bin/bash
		set -e
		curl -o /tmp/startk3s -sfL https://raw.githubusercontent.com/kf5i/k3ai-plugins/main/plugin_wsl_start/startk3s  
		$SUDO mv /tmp/startk3s /usr/local/bin
		chmod +x /usr/local/bin/startk3s
		/usr/local/bin/startk3s
		waiting_pod_array=( "k8s-app=kube-dns;kube-system" 
							"k8s-app=metrics-server;kube-system"
						  )

		for i in "${waiting_pod_array[@]}"; do 
			echo "$i"; 
			IFS=';' read -ra VALUES <<< "$i"
			wait "${VALUES[0]}" "${VALUES[1]}"
		done`

		cmd.Printf("About to execute the following:\n%v\n", k3sStartScript)

		// I wonder if the below code is right - you could remove the if statement, but then it's slightly less
		// readable, assuming that a user has to deduce that returning err could be nil
		if err := utils.ExecuteInlineBashScript(cmd, k3sStartScript, "K3s failed to start."); err != nil {
			return "", err
		}
	}

	return li.DetectK3s("k3s")
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
	$KUBECTL_COMMAND config set-context --current --namespace=kubeflow
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
