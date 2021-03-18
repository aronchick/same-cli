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
	"path"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
func Execute(version string) {
	RootCmd.Version = version

	cobra.OnInitialize(initConfig)

	setVersion()

	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}

func setVersion() {
	template := fmt.Sprintf("SAME version: %s\n", RootCmd.Version)
	RootCmd.SetVersionTemplate(template)
}

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to debug
	if !ok {
		lvl = "debug"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.InfoLevel
	}
	// set global log level
	logrus.SetLevel(ll)

	log.Tracef("Log Level: %v\n", logrus.GetLevel())
	log.Traceln("Starting Root Init")
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
	viper.SetEnvPrefix("same")
	viper.AutomaticEnv()

	log.Traceln("- in Root.initConfig")

	if cfgFile == "" {

		if (os.Geteuid() == 0) && (os.Getenv("SUDO_UID") != "") {
			RootCmd.Println("Running as sudo, skipping over checking for configuration file in your root directory. If this is a mistake, please use the --config flag.")
			return
		}
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			if !utils.PrintErrorAndReturnExit(RootCmd, "could not find home directory: ", err) {
				// Use the below to pass along that we've failed and we should exit at the first check
				os.Setenv("TEST_EXIT", "1")
				return
			}
		}

		cfgFile = path.Join(home, ".same", "config.yaml")
	}
	err = utils.LoadConfig(cfgFile)
	if err != nil {
		if !utils.PrintErrorAndReturnExit(RootCmd, "Error reading config file: %v", err) {

			// Use the below to pass along that we've failed and we should exit at the first check
			os.Setenv("TEST_EXIT", "1")
			return
		}
	}
}
