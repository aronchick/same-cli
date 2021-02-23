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
	"path"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "same",
	Short: "Interact with self-assembling machine learning environment configurations",
	Long:  `A longer SAME Root Description.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log.Info("in base")
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}

func init() {
	log.Info("in root init")
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.same.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log.Info("in initConfig")
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			message := fmt.Sprintf("could not find home directory: %v", err)
			RootCmd.Println(message)
			log.Fatalf(message)
			return
		}

		cfgFile = path.Join(home, ".same", "config.yaml")
	}

	err := utils.LoadConfig(cfgFile)
	if err != nil {
		message := fmt.Sprintf("Error reading config file: %v", err)
		RootCmd.Println(message)
		log.Fatalf(message)
	}
}
