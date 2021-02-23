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

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

func init() {
	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.PersistentFlags().String("program-id", "", "The ID of a SAME Program")
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

	RootCmd.AddCommand(programCmd)

}
