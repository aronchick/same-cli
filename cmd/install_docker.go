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
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/onsi/gomega/gbytes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installDockerCmd represents the installDocker command
var installDockerCmd = &cobra.Command{
	Use:   "install_docker",
	Short: "Installs docker on your system. Requires running as sudo.",
	Long:  `Installs docker on your system. Requires running as sudo. Longer description.`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Geteuid() == 0 {
			if os.Getenv("SUDO_UID") == "" || os.Getenv("SUDO_UID") == "0" {
				log.Fatalf("You are executing this command as root, which is almost certainly not what you want to do. Please log in as the user that will use 'same' who has sudoer rights (but is not root).")
			}
		} else {
			log.Fatalf("You have to execute this command under a sudo environment. Please use the command `sudo same install_docker`")
		}
		executingUser, err := user.LookupId(os.Getenv("SUDO_UID"))
		if err != nil {
			log.Fatalf("Could not find the user who executed this command.")
		}

		if utils.DetectRootless() {
			logrus.Info("It _appears_ you have docker rootless installed (based on the presence of both *rootless* apt packages and your DOCKER_HOST is pointing at '/run/user'.). This has been an issue in the past - please uninstall if you run into problems.")
		}

		logrus.Info("Updating apt")
		updateAptScript := `
		#!/bin/bash
		set -e
		apt-get update
		apt-get install -y \
			apt-transport-https \
			ca-certificates \
			curl \
			gnupg-agent \
			software-properties-common
		`
		buf := gbytes.NewBuffer()
		cmd.SetOut(buf)

		if err := utils.ExecuteInlineBashScript(cmd, updateAptScript, "Failure updating apt."); err != nil {
			log.Fatalf("Failed to install Docker. That's all we know: %v", err)
		}

		logrus.Info("Updating docker packages")
		updateDockerPackages := `
		#!/bin/bash
		set -e
		curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
		add-apt-repository \
			"deb [arch=amd64] https://download.docker.com/linux/ubuntu \
			$(lsb_release -cs) \
			stable"
		apt-get update
		`

		buf = gbytes.NewBuffer()
		cmd.SetOut(buf)

		if err := utils.ExecuteInlineBashScript(cmd, updateDockerPackages, "Failure during adding Docker packages."); err != nil {
			log.Fatalf("Failed to install Docker. That's all we know: %v", err)
		}

		logrus.Info("Installing docker packages.")
		installDockerScript := `
		#!/bin/bash
		set -e
		apt-get install -y docker-ce docker-ce-cli containerd.io
		`

		buf = gbytes.NewBuffer()
		cmd.SetOut(buf)

		if err := utils.ExecuteInlineBashScript(cmd, installDockerScript, "Failure during executing installing Docker."); err != nil {
			log.Fatalf("Failed to install Docker. That's all we know: %v", err)
		}

		logrus.Infof("Adding user '%v' to docker system group.", executingUser.Username)
		addingToUserGroupScript := `
		#!/bin/bash
		set -e
		usermod -aG docker ` + executingUser.Username + `
		sudo su ` + executingUser.Username + `
		export DOCKER_HOST=unix:///run/docker.sock
		docker run hello-world
		`

		buf = gbytes.NewBuffer()
		cmd.SetOut(buf)

		if err := utils.ExecuteInlineBashScript(cmd, addingToUserGroupScript, "Could not add user to usergroup."); err != nil {
			log.Fatalf("Failed to install Docker. That's all we know: %v", err)
		}

		if strings.Contains(string(buf.Contents()), "Hello from Docker!") {
			message := "Docker was successfully installed!"
			cmd.Println(message)
			fmt.Println(message)
		}

	},
}

func init() {
	RootCmd.AddCommand(installDockerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installDockerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installDockerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
