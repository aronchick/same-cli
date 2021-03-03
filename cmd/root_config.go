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
	"fmt"

	"github.com/azure-octo/same-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Initialize your .same/config.yaml file.",
	Long:  `Creates and initializes environment wide settings in .same/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defaultLocation := "~/.same/config.yaml"
		err = utils.LoadConfig(defaultLocation)
		if err != nil {
			message := fmt.Errorf("Error loading config file from '%v': %v", defaultLocation, err)
			return message
		}
		allSettings := viper.AllSettings()

		// len in go checks for both nil and 0
		if len(allSettings) == 0 {
			log.Trace("No settings found, assuming there is no config file (probably wrong)\n")
			message := fmt.Sprintf("No SAME config file detected, initializing with default values in '%v'", defaultLocation)
			cmd.Println(message)
			viper.SetConfigFile(defaultLocation)
			viper.Set("METADATA_STORE", nil)
			err := viper.SafeWriteConfig()
			if err != nil {
				err_message := fmt.Errorf("Error writing to SAME config file '%v': %v", defaultLocation, err)
				log.Errorf(err_message.Error())
				return err_message
			}
			log.Tracef("Wrote config to %v\n", defaultLocation)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(configCmd)

	log.Tracef("In config init")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("init", "i", false, "Initialize the SAME configuration file with the default values. Stored in '~/.same/config.yaml'")
}
