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

	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"

	"github.com/spf13/cobra"
)

var listProgramCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all deployed SAME programs",
	Long:  `Lists all deployed SAME programs.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Failed during dependency checks: %v", err) {
				return err
			}
		}

		listOfPipelines, err := ListPipelines()
		if err != nil {
			return err
		}
		for _, thisPipeline := range listOfPipelines {
			//TODO: Making the formatting nicer
			fmt.Println(thisPipeline.ID, thisPipeline.Name, thisPipeline.Description)
		}

		return nil
	},
}

func init() {
	programCmd.AddCommand(listProgramCmd)
}
