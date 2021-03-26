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
package cmd

import (
	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installK3sCmd represents the installK3s command
var installK3sCmd = &cobra.Command{
	Use:   "installK3s",
	Short: "Install K3s on the local machine",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log.Trace("Starting installK3s command")
		var i = GetClusterInstallMethods()

		_, err = i.InstallK3s(cmd)
		if err != nil {
			log.Fatalf("error installing k3s: %v", err)
		}
		cmd.Println("K3s installed.")
		k3sRunning, _ := utils.GetUtils().IsK3sRunning(cmd)
		if err == nil && k3sRunning {
			cmd.Println("K3s started.")
		} else {
			log.Fatalf("K3s installed with no error, but is not running. Error: %v", err)
		}

		// Need explicit path to detect
		_, err = utils.GetUtils().DetectK3s()
		if err != nil {
			log.Fatalf("error detecting k3s: %v", err)
		}
		cmd.Println("K3s detected.")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(installK3sCmd)

}
