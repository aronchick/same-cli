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
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/utils"
	gogetter "github.com/hashicorp/go-getter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// programCmd represents the program command
var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Create and Update Programs",
}

var CreateProgramCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a SAME program",
	Long: `Creates a SAME program from a SAME program file.
	
	A SAME program can be a ML pipeline.
	
	This command configures the program but does not execute it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if _, err := kubectlExists(); err != nil {
			log.Error(err.Error())
			return err
		}

		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", "kubectl config set 'contexts.'`kubectl config current-context`'.namespace' kubeflow").Run(); err != nil {
			message := fmt.Errorf("Could not set kubeconfig default context to use kubeflow namespace: %v", err)
			log.Error(message.Error())
			return message
		}

		// for demo
		log.Infof("File Location: %v\n", filePath)
		log.Infof("File Name: %v\n", fileName)

		onDiskLocation, err := getFilePath(filePath)
		if err != nil {
			log.Errorf("could not load same config file: %v", err)
			return err
		}

		UploadPipeline(onDiskLocation, programName, programDescription)

		return nil
	},
}

var runProgramCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a SAME program",
	Long:  `Runs a SAME program that was already created.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pipelineId, err := cmd.PersistentFlags().GetString("program-id")
		if err != nil {
			pipelineId = ""
		}
		if pipelineId == "" {
			err = viper.ReadInConfig()
			if err != nil {
				log.Errorf(fmt.Sprintf("error loading configuration file: %v", err))
				return err
			}
			pipelineId = viper.GetString("activepipeline")
			if pipelineId == "" {
				println("Must specify --program-id, or create new SAME program.")
			}
		}

		runName, err := cmd.PersistentFlags().GetString("run-name")
		if err != nil {
			return err
		}
		runDescription, err := cmd.PersistentFlags().GetString("run-description")
		if err != nil {
			runDescription = ""
		}

		experimentName, err := cmd.PersistentFlags().GetString("experiment-name")
		if err != nil {
			experimentName = "SAME Experiment"
		}

		experimentDescription, err := cmd.PersistentFlags().GetString("experiment-description")
		if err != nil {
			experimentDescription = "A SAME Experiment Description"
		}

		params, _ := cmd.PersistentFlags().GetStringSlice("run-param")

		runParams := make(map[string]string)

		for _, param := range params {
			parts := strings.Split(param, "=")
			if len(parts) != 2 {
				println(fmt.Sprintf("Invalid param format %s. Expect: key=value", param))
			}
			runParams[parts[0]] = parts[1]
		}

		if _, err := kubectlExists(); err != nil {
			log.Errorf(err.Error())
			return err
		}

		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", "kubectl config set 'contexts.'`kubectl config current-context`'.namespace' kubeflow").Run(); err != nil {
			log.Errorf("Could not set kubeconfig default context to use kubeflow namespace.")
			return err
		}

		// TODO: Use an existing experiment if name exists
		experimentId := CreateExperiment(experimentName, experimentDescription).ID
		runDetails := CreateRun(runName, pipelineId, experimentId, runDescription, runParams)

		fmt.Printf("Program run created with ID %s.", runDetails.Run.ID)

		return nil
	},
}

// getFilePath returns a file path to the local drive of the SAME config file, or error if invalid.
// If the file is remote, it pulls from a GitHub repo.
// Expects a full file path (including the file name)
func getFilePath(putativeFilePath string) (filePath string, err error) {
	isRemoteFile, err := utils.IsRemoteFilePath(putativeFilePath)

	if err != nil {
		log.Errorf("could not tell if the file was remote or not: %v", err)
		return "", err
	}

	if isRemoteFile {
		tempSameDir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Errorf("error creating a temporary directory to copy the file to: %v", err)
			return "", err
		}

		// Get path to store the file to
		tempSamePath := path.Join(tempSameDir, "tmp_same.yaml")

		configFileUri, err := netUrl.Parse(putativeFilePath)
		if err != nil {
			return "", fmt.Errorf("could not parse sameFile url: %v", err)
		}
		finalUrl := utils.UrlToRetrive(configFileUri.String(), putativeFilePath)
		log.Infof("Downloading from %v to %v", finalUrl, tempSamePath)
		errGet := gogetter.GetFile(tempSamePath, finalUrl.String())
		if errGet != nil {
			return "", fmt.Errorf("could not parse sameFile url: %v", errGet)
		} else {
			g := new(gogetter.FileGetter)
			g.Copy = true
			errGet := g.GetFile(tempSamePath, configFileUri)
			if errGet != nil {
				return "", fmt.Errorf("could not get sameFile from url: %v\nerror: %v", configFileUri, err)
			}
		}

		filePath = tempSamePath
	} else {
		if !fileExists(putativeFilePath) {
			return "", fmt.Errorf("could not find sameFile at: %v\nerror: %v", putativeFilePath, err)
		}
	}
	return putativeFilePath, nil
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
	_, err := os.Stat(path)
	return os.IsExist(err)
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

	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.PersistentFlags().StringP("program-id", "i", "", "The ID of a SAME Program")
	runProgramCmd.PersistentFlags().StringP("experiment-name", "e", "", "The name of a SAME Experiment to be created or reused.")

	runProgramCmd.PersistentFlags().String("experiment-description", "", "The description of a SAME Experiment to be created.")
	runProgramCmd.PersistentFlags().String("run-name", "", "The name of the SAME program run.")

	runProgramCmd.PersistentFlags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.PersistentFlags().StringSlice("run-param", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")

	RootCmd.AddCommand(programCmd)

}
