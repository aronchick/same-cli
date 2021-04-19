package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	netUrl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	gogetter "github.com/hashicorp/go-getter"
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

func (u *UtilsLive) IsK3sRunning() (running bool, err error) {
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

func (u *UtilsLive) Detect(absFilePath string, cwd string, detectors []gogetter.Detector) (string, error) {
	return gogetter.Detect(absFilePath, cwd, detectors)
}

func (u *UtilsLive) GetFile(tempSameFilename string, corrected_url string) error {
	return gogetter.GetFile(tempSameFilename, corrected_url)
}

// getFilePath returns a file path to the local drive of the SAME config file, or error if invalid.
// If the file is remote, it pulls from a GitHub repo.
// Expects a full file path (including the file name)
func (u *UtilsLive) GetConfigFilePath(putativeFilePath string) (filePath string, err error) {
	// TODO: aronchick: This is all probably unnecessary. We could just swap everything out
	// for gogetter.GetFile() and punt the whole problem at it.
	// HOWEVER, that doesn't solve for when a github url has an https schema, which causes
	// gogetter to weirdly reformats he URL (dropping the repo).
	// E.g., gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml")
	// Fails with a bad response code: 404
	// and 	gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo/same.yaml")
	// Fails with fatal: repository 'https://github.com/SAME-Project/' not found

	isRemoteFile, err := u.IsRemoteFilePath(putativeFilePath)

	if err != nil {
		log.Errorf("could not tell if the file was remote or not: %v", err)
		return "", err
	}

	if isRemoteFile {
		// Use the default system temp directory and a randomly generated name
		tempSameDir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Errorf("error creating a temporary directory to copy the file to (we're using the standard temporary directory from your system, so this could be an issue of the permissions this CLI is running under): %v", err)
			return "", err
		}

		// Get path to store the file to
		tempSameFile, err := ioutil.TempFile(tempSameDir, "")
		if err != nil {
			return "", fmt.Errorf("could not create temporary file in %v", tempSameDir)
		}

		configFileUri, err := netUrl.Parse(putativeFilePath)
		if err != nil {
			return "", fmt.Errorf("could not parse sameFile url: %v", err)
		}

		// TODO: Hard coding 'same.yaml' in now - should be optional
		finalUrl, err := u.UrlToRetrive(configFileUri.String(), "same.yaml")
		if err != nil {
			message := fmt.Errorf("unable to process the url to retrieve from the provided configFileUri(%v): %v", configFileUri.String(), err)
			log.Error(message)
			return "", message
		}

		corrected_url := finalUrl.String()
		if (finalUrl.Scheme == "https") || (finalUrl.Scheme == "http") {
			log.Info("currently only support http and https on github.com because we need to prefix with git")
			corrected_url = "git::" + finalUrl.String()
		}

		log.Infof("Downloading from %v to %v", corrected_url, tempSameFile.Name())
		errGet := u.GetFile(tempSameFile.Name(), corrected_url)
		if errGet != nil {
			return "", fmt.Errorf("could not download SAME file from URL '%v': %v", finalUrl.String(), errGet)
		}

		filePath = tempSameFile.Name()
	} else {
		cwd, _ := os.Getwd()
		log.Tracef("CWD: %v", cwd)
		absFilePath, _ := filepath.Abs(putativeFilePath)
		filePath, _ = u.Detect(absFilePath, cwd, []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.FileDetector)})

		if !u.FileExists(filePath) {
			return "", fmt.Errorf("could not find sameFile at: %v\nerror: %v", putativeFilePath, err)
		}
	}
	return filePath, nil
}

func (u *UtilsLive) FileExists(path string) (fileDoesExist bool) {
	resolvedPath, err := netUrl.Parse(path)
	if err != nil {
		log.Errorf("could not parse path '%v': %v", path, err)
		return false
	}
	_, err = os.Stat(resolvedPath.Path)
	return err == nil
}

// IsRemoteFile checks if the path configFile is remote (e.g. http://github...)
func (u *UtilsLive) IsRemoteFile(configFile string) (bool, error) {
	if configFile == "" {
		message := fmt.Errorf("config file must be a URI or a path")
		log.Errorf(message.Error())
		return false, message
	}
	url, err := netUrl.Parse(configFile)
	if err != nil {
		message := fmt.Errorf("error parsing file path: %v", err)
		log.Errorf(message.Error())
		return false, message
	}
	if url.Scheme != "" {
		return true, nil
	}
	return false, nil
}

func (u *UtilsLive) IsEndpointReachable(url string) (bool, error) {
	timeout := 1 * time.Second
	parsed_url, err := netUrl.Parse(url)
	if err != nil {
		return false, fmt.Errorf("could not parse url: %v", err)
	}

	// We need to execute the below because if a URL comes in with no schema, then we should just pass it through as a raw string
	// This will break if someone has a URL with a slash in it (is that even possible?) but there's only so much we can do here.
	url_to_query := parsed_url.Host
	if parsed_url.Host == "" {
		url_to_query = parsed_url.String()
		url_to_query = strings.Split(url_to_query, "/")[0]
	}

	_, err = net.DialTimeout("tcp", url_to_query, timeout)
	if err != nil {
		return false, fmt.Errorf("could not reach endpoint (%v): %v", url, err)
	}

	return true, nil
}
