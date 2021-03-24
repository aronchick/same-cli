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
	"os"
	"os/exec"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var CreateProgramCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a SAME program",
	Long: `Creates a SAME program from a SAME program file.
	
	A SAME program can be a ML pipeline.
	
	This command configures the program but does not execute it.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log.Tracef("In Create.RunE")

		// There's probably a better way to do this, but need to figure out how to pass back a value from initConfig (when tests fail but panics are mocked)
		if os.Getenv("TEST_EXIT") == "1" {
			log.Traceln("Detected that we're in a test and TEST_EXIT is set, so returning.")
			return
		}

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

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(cmd); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Error while checking dependencies: %v", err) {
				return nil
			}
		}

		// HACK: Currently Kubeconfig must define default namespace
		commandToRun := "kubectl config set 'contexts.'`kubectl config current-context`'.namespace' kubeflow"
		log.Tracef("About to run: %v", commandToRun)
		if err := exec.Command("/bin/bash", "-c", commandToRun).Run(); err != nil {
			message := fmt.Errorf("Could not set kubeconfig default context to use kubeflow namespace: %v", err)
			log.Error(message.Error())
			return message
		}

		// for demo
		log.Tracef("File Location: %v\n", filePath)
		log.Tracef("File Name: %v\n", fileName)

		// Load config file. Explicit parameters take precedent over config file.
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

		if sameConfigFile.Spec.Pipeline.Name != "" && programName == "" {
			programName = sameConfigFile.Spec.Pipeline.Name
		}

		if sameConfigFile.Spec.Pipeline.Description != "" && programDescription == "" {
			programDescription = sameConfigFile.Spec.Pipeline.Description
		}

		pipeline, err := FindPipelineByName(programName)
		if err != nil {
			uploadedPipeline, err := UploadPipeline(sameConfigFile, programName, programDescription)
			if err != nil {
				return err
			}

			cmd.Printf("Pipeline Uploaded.\nName: %v\nID: %v", uploadedPipeline.Name, uploadedPipeline.ID)
		} else {
			newID, _ := uuid.NewRandom()
			uploadedPipelineVersion, err := UpdatePipeline(sameConfigFile, pipeline.ID, newID.String())
			if err != nil {
				return err
			}

			cmd.Printf("Pipeline Updated.\nName: %v\nVersionID: %v\nID: %v", uploadedPipelineVersion.Name, uploadedPipelineVersion.ID, pipeline.ID)
		}

		return nil
	},
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
	CreateProgramCmd.PersistentFlags().StringP("name", "n", "", "The program name")
	CreateProgramCmd.PersistentFlags().String("description", "", "Brief description of the program")

}
