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
	pipelineClientParams "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	apiclient "github.com/kubeflow/pipelines/backend/src/common/client/api_server"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a pipeline.",
	Long:  `Deletes a pipeline. Longer Description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flag("name").Value == nil && cmd.Flag("id").Value == nil {
			message := "'name' or 'id' must be set to delete a flag"
			return fmt.Errorf(message+"%v", err)
		}

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(); err != nil {
			return fmt.Errorf("Failed during dependency checks: %v", err)
		}

		kfpconfig, err := utils.NewKFPConfig()
		if err != nil {
			return err
		}

		pipelineID := ""
		if cmd.Flag("id").Value != nil {
			pipelineID = cmd.Flag("id").Value.String()
		} else {
			pipelineName := cmd.Flag("name").Value.String()
			pipeline, err := FindPipelineByName(pipelineName)
			if err != nil {
				message := fmt.Errorf("error while searching for pipeline: %v", err)
				log.Errorf("delete.go:" + message.Error())
				cmd.Print(message.Error())
				return message
			} else if pipeline == nil {
				message := fmt.Errorf("could not find a pipeline with the name: %v", pipelineName)
				log.Errorf("delete.go:" + message.Error())
				cmd.Print(message.Error())
				return message
			}
		}

		deleteClient, err := apiclient.NewPipelineClient(kfpconfig, false)
		if err != nil {
			message := fmt.Errorf("could not create API client for deleting a pipeline pipeline: %v", err)
			cmd.Print(message.Error())
			log.Errorf("delete.go:" + message.Error())
			return message
		}

		deleteParams := pipelineClientParams.NewDeletePipelineParams()
		deleteParams.ID = pipelineID

		err = deleteClient.Delete(deleteParams)
		if err != nil {
			message := fmt.Sprintf("could not delete the pipeline with ID (%v): %v", pipelineID, err)
			return fmt.Errorf(message+"%v", err)
		}

		cmd.Printf("Successfully deleted pipeline ID: %v", pipelineID)

		return nil

	},
}

func init() {
	deleteCmd.PersistentFlags().StringP("id", "i", "", "ID of the pipeline to delete.")
	deleteCmd.PersistentFlags().StringP("name", "n", "", "Name of the pipeline to delete. No check is made for duplicate pipelines.")

	programCmd.AddCommand(deleteCmd)
}
