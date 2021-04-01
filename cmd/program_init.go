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
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
)

var sameRepoName string
var randSuffix bool
var err error

// initCmd represents the init command
var programInitCmd = &cobra.Command{
	Use:   "init -f ${CONFIG}",
	Short: "Initialize an empty directory with a SAME compliant repo.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		randSuffix, _ = cmd.Flags().GetBool("random_suffix")
		ExecuteInit(string(sameRepoName))

		return err
	},
}

func init() {
	programCmd.AddCommand(programInitCmd)

	// Config file option
	programInitCmd.PersistentFlags().StringVarP(&sameRepoName, "name", "n", "", `Name of the SAME repo to use.`)
	programInitCmd.PersistentFlags().BoolP("random_suffix", "r", false, "Add a random suffix to the repo (for testing purposes, usually).")

}

// ExecuteInit executes a full initialization of the directory.
func ExecuteInit(sameRepoName string) {
	// Using a fixed seed will produce the same output on every run.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if randSuffix {
		sameRepoName = fmt.Sprintf("%v-%v", sameRepoName, r.Int31())
	}

	log.Info(fmt.Sprintf("SameRepoName is %v", sameRepoName))

	dir, err := ioutil.TempDir(os.TempDir(), "foobaz")
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info(fmt.Sprintf("Working directory: cd %v", dir))

	nextCommand("git", []string{"init"}, dir)

	full_repo_url := fmt.Sprintf("https://github.com/Contoso-University/%v", sameRepoName)
	nextCommand("gh", []string{"repo", "create", "-p", "the-same-project/same-repo", "-y", "--public", fmt.Sprintf("Contoso-University/%v", sameRepoName)}, dir)

	if err != nil {
		log.Errorf("unable to clone the SAME project repository: %v", err.Error())
	}

	final_local_dir := fmt.Sprintf("%v/%v", os.TempDir(), sameRepoName)

	log.Printf("The location of the repo is: %v", final_local_dir)
	time.Sleep(5 * time.Second)

	log.Info(fmt.Sprintf("git clone %s %s", full_repo_url, os.TempDir()))
	repo, err := git.PlainClone(final_local_dir, false, &git.CloneOptions{
		URL: full_repo_url,
	})

	if err != nil {
		log.Errorf("unable to clone the with go-git repository: %v", err.Error())
	}
	_ = repo

}

func nextCommand(command string, args []string, dir string) {
	// cmd := exec.Command("myCommand", "arg1", "arg2")
	// cmd.Dir = "/path/to/work/dir"
	// cmd.Run()

	commandLookPath, _ := exec.LookPath(command)
	cmd := exec.Command(commandLookPath, args...)
	cmd.Dir = dir
	finalCommand := cmd.String()
	log.Info(fmt.Sprintf("Command: %v", finalCommand))
	out, err := cmd.CombinedOutput()

	log.Info(fmt.Sprintf("Output: %v", string(out[:])))

	if err != nil {
		log.Error("================")
		log.Error("=  Error       =")
		log.Error("================")
		log.Error("")

		log.Errorln(fmt.Sprintf("Final Command: %v", string(finalCommand)))
		log.Errorln(fmt.Sprintf("Output: %v", string(out[:])))
		log.Errorln(fmt.Sprintf("Error found: %v\n\n", err))

		return
	}
}
