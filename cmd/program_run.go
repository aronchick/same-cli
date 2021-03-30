package cmd

/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var runProgramCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a SAME program",
	Long:  `Runs a SAME program that was already created.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, err := cmd.PersistentFlags().GetString("file")
		if err != nil {
			return err
		}

		programName, err := cmd.PersistentFlags().GetString("program-name")
		if err != nil {
			programName = ""
		}

		programDescription, err := cmd.PersistentFlags().GetString("program-description")
		if err != nil {
			return err
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
			return err
		}

		experimentDescription, err := cmd.PersistentFlags().GetString("experiment-description")
		if err != nil {
			experimentDescription = ""
		}

		runOnly, err := cmd.PersistentFlags().GetBool("run-only")
		if err != nil {
			runOnly = false
		}

		kubectlCommand, err := cmd.PersistentFlags().GetString("kubectl-command")
		if err != nil {
			return err
		}

		if kubectlCommand == "" {
			kubectlCommand, err = infra.GetDependencyCheckers(cmd, args).IsKubectlOnPath(cmd)
			if err != nil {
				if utils.PrintErrorAndReturnExit(cmd, "could not get kubectl command: %v", err) {
					return err
				}
			}
		}

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(cmd); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Failed during dependency checks: %v", err) {
				return nil
			}
		}
		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v config set 'contexts.'`%v config current-context`'.namespace' kubeflow", kubectlCommand, kubectlCommand)).Run(); err != nil {
			message := fmt.Errorf("could not set kubeconfig default context to use kubeflow namespace: %v", err)
			log.Error(message.Error())
			return message
		}

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

		if sameConfigFile.Spec.ConfigFilePath == "" {
			sameConfigFile.Spec.ConfigFilePath = filePath
		}

		if sameConfigFile.Spec.Pipeline.Name != "" && programName == "" {
			programName = sameConfigFile.Spec.Pipeline.Name
		}

		pipelineID := ""
		pipelineVersionID := ""
		pipeline, err := FindPipelineByName(programName)
		if runOnly {
			if err == nil {
				pipelineID = pipeline.ID
			}
		} else {
			if err != nil {
				if sameConfigFile.Spec.Pipeline.Description != "" && programDescription == "" {
					programDescription = sameConfigFile.Spec.Pipeline.Description
				}
				uploadedPipeline, err := UploadPipeline(sameConfigFile, programName, programDescription)
				if err != nil {
					return err
				}
				pipelineID = uploadedPipeline.ID

				cmd.Printf("Pipeline Uploaded.\nName: %v\nID: %v", uploadedPipeline.Name, uploadedPipeline.ID)
			} else {
				pipelineID = pipeline.ID
				newID, _ := uuid.NewRandom()
				uploadedPipelineVersion, err := UpdatePipeline(sameConfigFile, pipelineID, newID.String())
				if err != nil {
					return err
				}
				pipelineVersionID = uploadedPipelineVersion.ID

				cmd.Printf("Pipeline Updated.\nName: %v\nVersionID: %v\nID: %v\n", uploadedPipelineVersion.Name, uploadedPipelineVersion.ID, pipeline.ID)
			}
		}

		// if ID is still blank we must exit
		if pipelineID == "" {
			log.Errorf("Could not find pipeline ID. Did you create the program?")
			return fmt.Errorf("could not determine program ID for run")
		}

		params, _ := cmd.PersistentFlags().GetStringSlice("run-param")

		runParams := make(map[string]string)

		if len(sameConfigFile.Spec.Run.Parameters) > 0 {
			runParams = sameConfigFile.Spec.Run.Parameters
		}

		// override the explicitly set run parameters
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				println(fmt.Sprintf("Invalid param format %q. Expect: key=value", param))
			}
			runParams[parts[0]] = parts[1]
		}

		experimentID := ""
		experiment, err := FindExperimentByName(experimentName)
		if experiment == nil || err != nil {
			experimentID = CreateExperiment(experimentName, experimentDescription).ID
		} else {
			experimentID = experiment.ID
		}

		if runName == "" && sameConfigFile.Spec.Run.Name != "" {
			runName = sameConfigFile.Spec.Run.Name
		}

		runDetails := CreateRun(runName, pipelineID, pipelineVersionID, experimentID, runDescription, runParams)

		fmt.Printf("Program run created with ID %s.", runDetails.Run.ID)

		return nil
	},
}

func init() {
	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.PersistentFlags().String("program-id", "", "The ID of a SAME Program")

	runProgramCmd.PersistentFlags().String("kubectl-command", "", "Kubectl binary command - include in single quotes.")

	runProgramCmd.PersistentFlags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")

	runProgramCmd.PersistentFlags().StringP("experiment-name", "e", "", "The name of a SAME Experiment to be created or reused.")
	err := runProgramCmd.MarkPersistentFlagRequired("experiment-name")
	if err != nil {
		message := "'experiment-name' is required for this to run.: %v"
		if utils.PrintErrorAndReturnExit(runProgramCmd, message, err) {
			return
		}
	}

	runProgramCmd.PersistentFlags().String("experiment-description", "", "The description of a SAME Experiment to be created.")
	runProgramCmd.PersistentFlags().StringP("run-name", "r", "", "The name of the SAME program run.")
	err = runProgramCmd.MarkPersistentFlagRequired("run-name")
	if err != nil {
		message := "'run-name' is required for this to run."
		if utils.PrintErrorAndReturnExit(RootCmd, message+"%v", err) {
			return
		}
		return
	}

	runProgramCmd.PersistentFlags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.PersistentFlags().StringSliceP("run-param", "p", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")
	runProgramCmd.PersistentFlags().String("program-description", "", "Brief description of the program")
	runProgramCmd.PersistentFlags().StringP("program-name", "n", "", "The program name")
	runProgramCmd.PersistentFlags().Bool("run-only", false, "Indicates whether to skip program upload")

}
