/*
Copyright © 2021 Bernd Verst <beverst@microsoft.ocm>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	netUrl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/utils"
	gogetter "github.com/hashicorp/go-getter"
	"github.com/spf13/cobra"
)

var CreateProgramCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a SAME program",
	Long: `Creates a SAME program from a SAME program file.
	
	A SAME program can be a ML pipeline.
	
	This command configures the program but does not execute it.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log.Debug("in create")
		for _, arg := range args {
			log.Debugf("arg: %v", arg)
		}

		filePath, err := cmd.PersistentFlags().GetString("file")
		if err != nil {
			return err
		}

		fileName, err := cmd.PersistentFlags().GetString("filename")
		if err != nil {
			return err
		}

		programName, err := cmd.PersistentFlags().GetString("name")
		if err != nil {
			return err
		}
		programDescription, err := cmd.PersistentFlags().GetString("description")
		if err != nil {
			return err
		}

		kubectlCommand, err := cmd.PersistentFlags().GetString("kubectl-command")
		if err != nil {
			return err
		}

		if kubectlCommand == "" {
			if _, err := kubectlExists(); err != nil {
				log.Error(err.Error())
				return err
			}
			kubectlCommand = "kubectl"
		} else {
			// Remove beginning and ending quotes if present
			kubectlCommand = regexp.MustCompile(`['"]*([^'"]*)['"]*`).ReplaceAllString(kubectlCommand, `$1`)
		}
		// HACK: Currently Kubeconfig must define default namespace
		commandToRun := fmt.Sprintf("%v config set 'contexts.'`%v config current-context`'.namespace' kubeflow", kubectlCommand, kubectlCommand)
		log.Infof("About to run: %v", commandToRun)
		if err := exec.Command("/bin/bash", "-c", commandToRun).Run(); err != nil {
			message := fmt.Errorf("Could not set kubeconfig default context to use kubeflow namespace: %v", err)
			log.Error(message.Error())
			return message
		}

		// for demo
		log.Infof("File Location: %v\n", filePath)
		log.Infof("File Name: %v\n", fileName)

		sameConfigFilePath, err := getConfigFilePath(filePath)
		if err != nil {
			log.Errorf("could not resolve SAME config file path: %v", err)
			return err
		}

		sameConfigFile, err := loaders.LoadSAME(sameConfigFilePath)
		if err != nil {
			log.Errorf("could not load SAME config file: %v", err)
			return err
		}

		if sameConfigFile.Spec.Pipeline.Name != "" {
			programName = sameConfigFile.Spec.Pipeline.Name
		}

		if sameConfigFile.Spec.Pipeline.Description != "" {
			programDescription = sameConfigFile.Spec.Pipeline.Description
		}

		uploadedPipeline, err := UploadPipeline(sameConfigFile, programName, programDescription)
		if err != nil {
			return err
		}

		cmd.Printf("Pipeline Uploaded.\nName: %v\nID: %v", uploadedPipeline.Name, uploadedPipeline.ID)

		return nil
	},
}

// getFilePath returns a file path to the local drive of the SAME config file, or error if invalid.
// If the file is remote, it pulls from a GitHub repo.
// Expects a full file path (including the file name)
func getConfigFilePath(putativeFilePath string) (filePath string, err error) {
	// TODO: aronchick: This is all probably unnecessary. We could just swap everything out
	// for gogetter.GetFile() and punt the whole problem at it.
	// HOWEVER, that doesn't solve for when a github url has an https schema, which causes
	// gogetter to weirdly reformats he URL (dropping the repo).
	// E.g., gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml")
	// Fails with a bad response code: 404
	// and 	gogetter.GetFile(tempFile.Name(), "github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml")
	// Fails with fatal: repository 'https://github.com/SAME-Project/' not found

	isRemoteFile, err := utils.IsRemoteFilePath(putativeFilePath)

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
		finalUrl, err := utils.UrlToRetrive(configFileUri.String(), "same.yaml")
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

		log.Infof("Downloading from %v to %v", corrected_url, tempSameFile)
		errGet := gogetter.GetFile(tempSameFile.Name(), corrected_url)
		if errGet != nil {
			return "", fmt.Errorf("could not download SAME file from URL '%v': %v", finalUrl.String(), errGet)
		}

		filePath = tempSameFile.Name()
	} else {
		cwd, _ := os.Getwd()
		absFilePath, _ := filepath.Abs(putativeFilePath)
		filePath, _ = gogetter.Detect(absFilePath, cwd, []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.FileDetector)})

		if !fileExists(filePath) {
			return "", fmt.Errorf("could not find sameFile at: %v\nerror: %v", putativeFilePath, err)
		}
	}
	return filePath, nil
}

func kubectlExists() (kubectlDoesExist bool, err error) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		err := fmt.Errorf("the 'kubectl' binary is not on your PATH: %v", os.Getenv("PATH"))
		return false, err
	}
	log.Infof("'kubectl' found at %v", path)
	return true, nil
}

func fileExists(path string) (fileDoesExist bool) {
	resolvedPath, err := netUrl.Parse(path)
	if err != nil {
		log.Errorf("could not parse path '%v': %v", path, err)
		return false
	}
	_, err = os.Stat(resolvedPath.Path)
	return err == nil
}

func init() {
	programCmd.AddCommand(CreateProgramCmd)

	CreateProgramCmd.PersistentFlags().StringP("file", "f", "", "a SAME program file")
	err := CreateProgramCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		log.Errorf("could not set 'file' flag as required: %v", err)
		return
	}

	CreateProgramCmd.PersistentFlags().StringP("filename", "c", "same.yaml", "The filename for the same file (defaults to 'same.yaml')")

	CreateProgramCmd.PersistentFlags().StringP("name", "n", "SAME Program", "The program name")
	CreateProgramCmd.PersistentFlags().String("description", "", "Brief description of the program")
	CreateProgramCmd.PersistentFlags().String("kubectl-command", "", "Kubectl binary command - include in single quotes.")

}
