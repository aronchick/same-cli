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
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// programCmd represents the program command
var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Create and Update Programs",
}

var createProgramCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a SAME program",
	Long: `Creates a SAME program from a SAME program file.
	
	A SAME program can be a ML pipeline.
	
	This command configures the program but does not execute it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, err := cmd.PersistentFlags().GetString("file")
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

		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", "kubectl config set 'contexts.'`kubectl config current-context`'.namespace' kubeflow").Run(); err != nil {
			println("Could not set kubeconfig default context to use kubeflow namespace.")
			return err
		}

		// for demo
		fmt.Println(fileName)

		UploadPipeline(fileName, programName, programDescription)

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
				log.Fatal(fmt.Sprintf("error loading configuration file: %v", err))
				os.Exit(1)
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
			experimentName = "A SAME Experiment"
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

		// HACK: Currently Kubeconfig must define default namespace
		if err := exec.Command("/bin/bash", "-c", "kubectl config set 'contexts.'`kubectl config current-context`'.namespace' kubeflow").Run(); err != nil {
			println("Could not set kubeconfig default context to use kubeflow namespace.")
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
	programCmd.AddCommand(createProgramCmd)

	createProgramCmd.PersistentFlags().String("file", "", "a SAME program file")
	err := createProgramCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		log.Fatal(fmt.Sprintf("could not set 'file' flag as required: %v", err))
		os.Exit(1)
	}

	createProgramCmd.PersistentFlags().String("name", "SAME Program", "The program name")
	createProgramCmd.PersistentFlags().String("description", "", "Brief description of the program")

	programCmd.AddCommand(runProgramCmd)

	runProgramCmd.PersistentFlags().String("program-id", "", "The ID of a SAME Program")
	runProgramCmd.PersistentFlags().String("experiment-name", "", "The name of a SAME Experiment to be created or reused.")
	err = runProgramCmd.MarkPersistentFlagRequired("experiment-name")
	if err != nil {
		log.Fatal(fmt.Sprintf("could not set 'experiment-name' flag as required: %v", err))
		os.Exit(1)
	}

	runProgramCmd.PersistentFlags().String("experiment-description", "", "The description of a SAME Experiment to be created.")
	runProgramCmd.PersistentFlags().String("run-name", "", "The name of the SAME program run.")
	err = runProgramCmd.MarkPersistentFlagRequired("run-name")
	if err != nil {
		log.Fatal(fmt.Sprintf("could not set 'run-name' flag as required: %v", err))
		os.Exit(1)
	}

	runProgramCmd.PersistentFlags().String("run-description", "", "A description of the SAME program run.")
	runProgramCmd.PersistentFlags().StringSlice("run-param", nil, "A paramater to pass to the program in key=value form. Repeat for multiple params.")

	rootCmd.AddCommand(programCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// programCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// programCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
