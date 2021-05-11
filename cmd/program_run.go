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
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		programName, err := cmd.Flags().GetString("program-name")
		if err != nil {
			programName = ""
		}

		programDescription, err := cmd.Flags().GetString("program-description")
		if err != nil {
			return err
		}

		runName, err := cmd.Flags().GetString("run-name")
		if err != nil {
			return err
		}
		runDescription, err := cmd.Flags().GetString("run-description")
		if err != nil {
			runDescription = ""
		}

		experimentName, err := cmd.Flags().GetString("experiment-name")
		if err != nil {
			return err
		}

		experimentDescription, err := cmd.Flags().GetString("experiment-description")
		if err != nil {
			experimentDescription = ""
		}

		runOnly, err := cmd.Flags().GetBool("run-only")
		if err != nil {
			runOnly = false
		}

		persistTemporaryFiles, err := cmd.Flags().GetBool("persist-temporary-files")
		if err != nil {
			persistTemporaryFiles = false
		}

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(); err != nil {
			return fmt.Errorf("Failed during dependency checks: %v", err)
		}

		// Load config file. Explicit parameters take precedent over config file.
		u := utils.GetUtils(cmd, args)
		sameConfigFilePath, err := u.GetConfigFilePath(filePath)
		if err != nil {
			return fmt.Errorf("could not resolve SAME config file path: %v", err)
		}

		sameConfigFile, err := loaders.V1{}.LoadSAME(sameConfigFilePath)
		if err != nil {
			return fmt.Errorf("could not load SAME config file: %v", err)
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
				uploadedPipeline, err := UploadPipeline(sameConfigFile, programName, programDescription, persistTemporaryFiles)
				if err != nil {
					return err
				}
				pipelineID = uploadedPipeline.ID

				cmd.Printf(`
Pipeline Uploaded.
Name: %v
ID: %v\n
`, uploadedPipeline.Name, uploadedPipeline.ID)
			} else {
				pipelineID = pipeline.ID
				newID, _ := uuid.NewRandom()
				uploadedPipelineVersion, err := UpdatePipeline(sameConfigFile, pipelineID, newID.String(), persistTemporaryFiles)
				if err != nil {
					return err
				}
				pipelineVersionID = uploadedPipelineVersion.ID

				cmd.Printf(`
Pipeline Updated.
Name: %v
ID: %v
VersionID: %v

`, uploadedPipelineVersion.Name, pipeline.ID, uploadedPipelineVersion.ID)
			}
		}

		// if ID is still blank we must exit
		if pipelineID == "" {
			log.Errorf("Could not find pipeline ID. Did you create the program?")
			return fmt.Errorf("could not determine program ID for run")
		}

		params, _ := cmd.Flags().GetStringSlice("run-param")

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
			experimentEntity, err := CreateExperiment(experimentName, experimentDescription)
			if err != nil {
				return err
			}
			experimentID = experimentEntity.ID
		} else {
			experimentID = experiment.ID
		}

		if runName == "" && sameConfigFile.Spec.Run.Name != "" {
			runName = sameConfigFile.Spec.Run.Name
		}

		runDetails, err := CreateRun(runName, pipelineID, pipelineVersionID, experimentID, runDescription, runParams)
		if err != nil {
			return err
		}

		fmt.Printf("Program run created with ID %s.\n", runDetails.Run.ID)

		return nil
	},
}

func init() {
	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.Flags().String("program-id", "", "The ID of a SAME Program")

	runProgramCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")

	runProgramCmd.Flags().StringP("experiment-name", "e", "", "The name of a SAME Experiment to be created or reused.")
	err := runProgramCmd.MarkFlagRequired("experiment-name")
	if err != nil {
		message := "'experiment-name' is required for this to run.: %v\n"
		fmt.Printf(message, err)
		return
	}

	runProgramCmd.Flags().String("experiment-description", "", "The description of a SAME Experiment to be created.")
	runProgramCmd.Flags().StringP("run-name", "r", "", "The name of the SAME program run.")
	err = runProgramCmd.MarkFlagRequired("run-name")
	if err != nil {
		message := "'run-name' is required for this to run."
		fmt.Printf(message+"%v", err)
		return
	}

	runProgramCmd.Flags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.Flags().StringSliceP("run-param", "p", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")
	runProgramCmd.Flags().String("program-description", "", "Brief description of the program")
	runProgramCmd.Flags().StringP("program-name", "n", "", "The program name")
	runProgramCmd.Flags().Bool("run-only", false, "Indicates whether to skip program upload")
	runProgramCmd.Flags().BoolP("persist-temporary-files", "t", false, "Persist temporary files in /tmp.")

}
