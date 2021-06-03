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
	"os"
	"path/filepath"
	"regexp"
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

		runDescription, err := cmd.Flags().GetString("run-description")
		if err != nil {
			runDescription = ""
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

		target, err := cmd.Flags().GetString("target")
		if err != nil {
			target = "kubeflow"
		}

		_ = target

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

		params, _ := cmd.Flags().GetStringSlice("run-param")

		runParams := make(map[string]interface{})

		if len(sameConfigFile.Spec.Run.Parameters) > 0 {
			runParams = sameConfigFile.Spec.Run.Parameters
		}

		// override the explicitly set run parameters
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("Invalid param format %q. Expect: key=value", param)
			}

			runParams[parts[0]] = parts[1]
		}

		log.Tracef("Target: %v", target)
		if target == "kubeflow" {

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
					uploadedPipeline, err := UploadPipeline(target, sameConfigFile, programName, programDescription, persistTemporaryFiles)
					if err != nil {
						return err
					}
					pipelineID = uploadedPipeline.ID

					cmd.Printf(`
Pipeline Uploaded.
Name: %v
ID: %v
`, uploadedPipeline.Name, uploadedPipeline.ID)
				} else {
					pipelineID = pipeline.ID
					newID, _ := uuid.NewRandom()
					uploadedPipelineVersion, err := UpdatePipeline(target, sameConfigFile, pipelineID, newID.String(), persistTemporaryFiles)
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

			experimentID := ""
			experiment, err := FindExperimentByName(sameConfigFile.Spec.Metadata.Name)
			if experiment == nil || err != nil {
				experimentEntity, err := CreateExperiment(sameConfigFile.Spec.Metadata.Name, experimentDescription)
				if err != nil {
					return err
				}
				experimentID = experimentEntity.ID
			} else {
				experimentID = experiment.ID
			}

			runDetails, err := CreateRun(sameConfigFile.Spec.Run.Name, pipelineID, pipelineVersionID, experimentID, runDescription, runParams)
			if err != nil {
				return err
			}

			fmt.Printf("Program run created with ID %s.\n", runDetails.Run.ID)
		} else if target == "aml" {
			log.Tracef("Executing AML target")
			log.Tracef("Freezing python packages")
			pipCommand := `
	#!/bin/bash
	set -e
	python3 -m pip freeze
	`

			cmdReturn, err := utils.ExecuteInlineBashScript(cmd, pipCommand, "Pip output failed", false)

			if err != nil {
				log.Tracef("Error executing: %v\n", err.Error())
			}
			requiredLibraries := []string{"dill", "azureml", "azureml.core", "azureml.pipeline"}

			missingLibraries := make([]string, 0)
			for _, lib := range requiredLibraries {
				r, _ := regexp.Compile(lib)
				if r.FindString(cmdReturn) == "" {
					missingLibraries = append(missingLibraries, lib)
				}
			}

			log.Tracef("Testing for missing libraries")
			if len(missingLibraries) > 0 {
				err = fmt.Errorf(`could not find all necessary libraries to execute. Please run:
pip3 install %v`, strings.Join(missingLibraries, " "))
				fmt.Println(err.Error())
				return err
			}

			requiredFields := []string{"AML_SP_PASSWORD_VALUE",
				"AML_SP_TENANT_ID",
				"AML_SP_APP_ID",
				"WORKSPACE_SUBSCRIPTION_ID",
				"WORKSPACE_RESOURCE_GROUP",
				"WORKSPACE_NAME",
				"AML_COMPUTE_NAME"}

			missingFields := make([]string, 0)
			for _, field := range requiredFields {
				if os.Getenv(field) == "" {
					missingFields = append(missingFields, field)
				}
			}

			if len(missingFields) > 0 {
				return fmt.Errorf("missing environment variables for: %v", strings.Join(missingFields, ", "))
			}

			compileDir, _, err := CompileFile("aml", *sameConfigFile, persistTemporaryFiles)
			if err != nil {
				return err
			}

			executeAMLPipeline := fmt.Sprintf(`
#!/bin/bash
set -e
cd %v
python3 %v
`, compileDir, filepath.Join(compileDir, "root.py"))

			log.Tracef("About to execute: %v\n", executeAMLPipeline)
			if cmdOut, err := utils.ExecuteInlineBashScript(cmd, executeAMLPipeline, "Running against AML pipelines failed:", true); err != nil {
				log.Tracef("Error executing: %v\n", err.Error())
				log.Tracef("Command output: %v\n", cmdOut)
				return err
			}

		}

		return nil
	},
}

func init() {
	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.Flags().String("program-id", "", "The ID of a SAME Program")

	runProgramCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")

	runProgramCmd.Flags().String("experiment-description", "", "The description of a SAME Experiment to be created.")

	runProgramCmd.Flags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.Flags().StringSliceP("run-param", "p", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")
	runProgramCmd.Flags().String("program-description", "", "Brief description of the program")
	runProgramCmd.Flags().StringP("program-name", "n", "", "The program name")
	runProgramCmd.Flags().Bool("run-only", false, "Indicates whether to skip program upload")
	runProgramCmd.Flags().Bool("persist-temporary-files", false, "Persist temporary files in /tmp.")
	runProgramCmd.Flags().StringP("target", "t", "kubeflow", "Enter one of 'kubeflow', 'aml'. Defaults to: kubeflow")

}
