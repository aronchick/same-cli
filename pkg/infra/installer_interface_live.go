package infra

import (
	"fmt"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

type LiveInstallers struct {
	_cmd            *cobra.Command
	_cmdArgs        []string
	_kubectlCommand string
}

func (i *LiveInstallers) InstallKFP() (err error) {
	cmd := i.GetCmd()

	log.Tracef("Inside InstallKFP()")
	kubectlCommand, err := i.GetKubectlCmd()
	if err != nil {
		return err
	}
	log.Tracef("kubectlCommand: %v\n", kubectlCommand)
	kfpInstall := fmt.Sprintf(`
	#!/bin/bash
	set -e
	export PIPELINE_VERSION=1.5.0
	export KUBECTL_COMMAND=%v
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/cluster-scoped-resources?ref=$PIPELINE_VERSION"
	$KUBECTL_COMMAND wait --for condition=established --timeout=60s crd/applications.app.k8s.io
	$KUBECTL_COMMAND apply -k "github.com/kubeflow/pipelines/manifests/kustomize/env/platform-agnostic-pns?ref=$PIPELINE_VERSION"
	$KUBECTL_COMMAND wait pods -l application-crd-id=kubeflow-pipelines -n kubeflow --for condition=Ready --timeout=1800s
	`, kubectlCommand)

	log.Tracef("About to execute: %v\n", kfpInstall)
	if _, err := utils.ExecuteInlineBashScript(cmd, kfpInstall, "KFP failed to install.", true); err != nil {
		log.Tracef("Error executing: %v\n", err.Error())
		return err
	}

	return nil
}

func (i *LiveInstallers) SetKubectlCmd(s string) {
	i._kubectlCommand = s
}

func (i *LiveInstallers) GetKubectlCmd() (string, error) {
	cmd := i.GetCmd()
	dc := GetDependencyCheckers(cmd, i.GetCmdArgs())
	if i._kubectlCommand == "" {
		kubectlPath, err := dc.IsKubectlOnPath()
		if err != nil || kubectlPath == "" {
			if err == nil {
				err = fmt.Errorf("")
			}
			return "", fmt.Errorf("Unable to detect the kubectl binary on your path, please check with 'which kubectl'. %v", err)
		} else {
			i.SetKubectlCmd(kubectlPath)
		}
	}
	return i._kubectlCommand, nil
}

func (i *LiveInstallers) GetCmd() *cobra.Command {
	return i._cmd
}

func (i *LiveInstallers) SetCmd(cmd *cobra.Command) {
	i._cmd = cmd
}

func (i *LiveInstallers) GetCmdArgs() []string {
	return i._cmdArgs
}

func (i *LiveInstallers) SetCmdArgs(args []string) {
	i._cmdArgs = args
}
