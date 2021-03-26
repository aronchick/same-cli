package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
)

type UtilsLive struct {
	// TODO: I hate this ... probably should fix (should move cmdArgs to somewhere else?)
	_cmdArgs []string
	_cmd     *cobra.Command
}

func (u *UtilsLive) DetectK3s() (string, error) {
	log.Info("Executing detect on: k3s")
	return exec.LookPath("k3s")
}

func (u *UtilsLive) IsK3sRunning(cmd *cobra.Command) (running bool, err error) {
	_, err = exec.LookPath("/usr/local/bin/k3s")
	var scriptCmd *exec.Cmd

	if err != nil {
		if runtime.GOOS == "darwin" {
			_, err := exec.LookPath("k3d")
			if err == nil {
				scriptCmd = exec.Command("/bin/bash", "-c", "kubectl get deployments --namespace=kube-system -o json")
			} else {
				log.Tracef("Neither K3s nor K3d found in path.")
				return false, fmt.Errorf("Neither K3s nor K3d appear in your path: %v", err)
			}
		} else {
			log.Tracef("K3s not found in path.")
			return false, fmt.Errorf("K3s does not appear in your path: %v", err)
		}
	} else {
		if !IsSudoer() {
			log.Tracef("K3s not found in path.")
			return false, fmt.Errorf("You must be part of sudoers in order to test for k3s: %v", err)
		}
		scriptCmd = exec.Command("/bin/bash", "-c", "sudo k3s kubectl get deployments --namespace=kube-system -o json")
	}

	scriptOutput, err := scriptCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("Failed to test if k3s is running. That's all we know: %v", err)
	}

	// Declared an empty interface
	//var result map[string]interface{}
	var result v1.DeploymentList

	//  waiting_pod_array=("k8s-app=kube-dns;kube-system"
	// "k8s-app=metrics-server;kube-system"

	// Unmarshal or Decode the JSON to the interface.
	//err = json.Unmarshal([]byte(scriptOutput), &result)
	err = json.Unmarshal(scriptOutput, &result)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshall result of k3s test: %v", err)
	}

	kubeDnsRunning := false
	metricsServerRunning := false

	for _, deployment := range result.Items {
		if k8sLabel, ok := deployment.Spec.Selector.MatchLabels["k8s-app"]; ok {
			switch k8sLabel {
			case "metrics-server":
				metricsServerRunning = (deployment.Status.ReadyReplicas > 0)
			case "kube-dns":
				kubeDnsRunning = (deployment.Status.ReadyReplicas > 0)
			}
		}
	}

	return kubeDnsRunning && metricsServerRunning, nil
}

func (u *UtilsLive) GetCmdArgs() []string {
	return u._cmdArgs
}

func (u *UtilsLive) SetCmdArgs(args []string) {
	u._cmdArgs = args
}

func (u *UtilsLive) SetCmd(cmd *cobra.Command) {
	u._cmd = cmd
}

func (u *UtilsLive) GetCmd() *cobra.Command {
	return u._cmd
}
