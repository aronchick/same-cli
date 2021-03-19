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
	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var runProgramCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a SAME program",
	Long:  `Runs a SAME program that was already created.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pipelineID, err := cmd.PersistentFlags().GetString("program-id")
		if err != nil {
			pipelineID = ""
		}

		filePath, err := cmd.PersistentFlags().GetString("file")
		if err != nil {
			return err
		}

		programName, err := cmd.PersistentFlags().GetString("name")
		if err != nil {
			programName = ""
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
		}

		if err := GetDependencyCheckers().CheckDependenciesInstalled(cmd); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Failed during dependency checks: %v", err) {
				return nil
			}
		}
		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v config set 'contexts.'`%v config current-context`'.namespace' kubeflow", kubectlCommand, kubectlCommand)).Run(); err != nil {
			message := fmt.Errorf("Could not set kubeconfig default context to use kubeflow namespace: %v", err)
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

		if pipelineID == "" {
			if sameConfigFile.Spec.Pipeline.Name != "" && programName == "" {
				programName = sameConfigFile.Spec.Pipeline.Name
			}
			pipeline, err := FindPipelineByName(programName)
			if err == nil {
				pipelineID = pipeline.ID
			}

			// if ID is still blank we must exit
			if pipelineID == "" {
				log.Errorf("Could not find pipeline ID. Did you create the program?")
				return fmt.Errorf("Could not determine program ID for run.")
			}
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

		runDetails := CreateRun(runName, pipelineID, experimentID, runDescription, runParams)

		fmt.Printf("Program run created with ID %s.", runDetails.Run.ID)

		return nil
	},
}

func init() {
	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.PersistentFlags().String("program-id", "", "The ID of a SAME Program")

	runProgramCmd.PersistentFlags().String("kubectl-command", "", "Kubectl binary command - include in single quotes.")

	runProgramCmd.PersistentFlags().StringP("file", "f", "", "a SAME program file")
	err = runProgramCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		log.Errorf("could not set 'file' flag as required: %v", err)
		return
	}
	runProgramCmd.PersistentFlags().StringP("filename", "c", "same.yaml", "The filename for the same file (defaults to 'same.yaml')")

	runProgramCmd.PersistentFlags().String("experiment-name", "", "The name of a SAME Experiment to be created or reused.")
	err := runProgramCmd.MarkPersistentFlagRequired("experiment-name")
	if err != nil {
		message := "'experiment-name' is required for this to run."
		RootCmd.Println(message)
		log.Fatalf(message)
		return
	}

	runProgramCmd.PersistentFlags().String("experiment-description", "", "The description of a SAME Experiment to be created.")
	runProgramCmd.PersistentFlags().String("run-name", "", "The name of the SAME program run.")
	err = runProgramCmd.MarkPersistentFlagRequired("run-name")
	if err != nil {
		message := "'run-name' is required for this to run."
		RootCmd.Println(message)
		log.Fatalf(message)
		return
	}

	runProgramCmd.PersistentFlags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.PersistentFlags().StringSlice("run-param", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")

}
