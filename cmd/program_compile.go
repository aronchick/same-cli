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
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

var compileProgramCmd = &cobra.Command{
	Use:   "compile",
	Short: "TEMPORARY FUNCTION: Compiles a notebook to a SAME program. Just exploring until we see how this might work.",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		persistTempFiles, err := cmd.Flags().GetBool("persist-temp-files")
		if err != nil {
			return err
		}

		target, err := cmd.Flags().GetString("target")
		if err != nil {
			target = "kubeflow"
		}

		if target == "aml" {
			requiredFields := []string{"AML_SP_PASSWORD_VALUE",
				"AML_SP_TENANT_ID",
				"AML_SP_APP_ID",
				"WORKSPACE_SUBSCRIPTION_ID",
				"WORKSPACE_RESOURCE_GROUP",
				"WORKSPACE_NAME",
				"AML_COMPUTE_NAME"}

			missingFields := make([]string, 0)
			for _, field := range requiredFields {
				if os.Getenv(field) == "" {
					missingFields = append(missingFields, field)
				}
			}
			if len(missingFields) > 0 {
				return fmt.Errorf("missing environment variables for: %v", strings.Join(missingFields, ", "))
			}
		}

		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(); err != nil {
			return fmt.Errorf("Failed during dependency checks: %v", err)
		}

		err = infra.GetDependencyCheckers(cmd, args).CheckForMissingPackages(target)
		if err != nil {
			return err
		}
		// Load config file. Explicit parameters take precedent over config file.
		u := utils.GetUtils(cmd, args)
		sameConfigFilePath, err := u.GetConfigFilePath(filePath)
		if err != nil {
			log.Errorf("could not resolve SAME config file path: %v", err)
			return err
		}

		sameConfigFile, err := loaders.V1{}.LoadSAME(sameConfigFilePath)
		if err != nil {
			log.Errorf("could not load SAME config file: %v", err)
			return err
		}

		for env_name, env := range sameConfigFile.Spec.Environments {
			var missing_credentials []string
			if env.PrivateRegistry {
				// Since we may need to create a secret, we need to pass along a Kubeconfig
				b := bytes.Buffer{}
				e := gob.NewEncoder(&b)
				clientConfig, err := utils.GetKubeConfig()
				if err != nil {
					return fmt.Errorf("error fetching kubeconfig")
				}
				err = e.Encode(clientConfig)
				if err != nil {
					return fmt.Errorf("error encoding kubeconfig to string: %v", err)
				}
				sameConfigFile.Spec.KubeConfig = base64.StdEncoding.EncodeToString(b.Bytes())

				if (loaders.RepositoryCredentials{} != env.Credentials) {
					log.Warnf("The environment '%v' has the credentials hard coded in the same file. This is likely a specatularly bad decision from a security standpoint. Cowardly going ahead anyway.", env_name)
				}

				image_pull_secret_name, err := cmd.Flags().GetString("image-pull-secret-name")

				image_pull_secret_server, err := cmd.Flags().GetString("image-pull-secret-server")
				if err != nil || image_pull_secret_server == "" {
					missing_credentials = append(missing_credentials, "image-pull-secret-server")
				}

				image_pull_secret_username, err := cmd.Flags().GetString("image-pull-secret-username")
				if err != nil || image_pull_secret_username == "" {
					missing_credentials = append(missing_credentials, "image-pull-secret-username")
				}

				image_pull_secret_password, err := cmd.Flags().GetString("image-pull-secret-password")
				if err != nil || image_pull_secret_password == "" {
					missing_credentials = append(missing_credentials, "image-pull-secret-password")
				}
				image_pull_secret_email, err := cmd.Flags().GetString("image-pull-secret-email")
				if err != nil || image_pull_secret_email == "" {
					missing_credentials = append(missing_credentials, "image-pull-secret-email")
				}
				if len(missing_credentials) > 0 && image_pull_secret_name == "" {
					return fmt.Errorf("You set environment '%v' to be a private repository, but you missed the following flags during execution: %v", env_name, missing_credentials)
				} else {
					if image_pull_secret_name != "" {
						log.Tracef("Using %v secret - this will override if you set any other values.", image_pull_secret_name)
					}
					env.Credentials.SecretName = image_pull_secret_name
					env.Credentials.Server = image_pull_secret_server
					env.Credentials.Username = image_pull_secret_username
					env.Credentials.Password = image_pull_secret_password
					env.Credentials.Email = image_pull_secret_email
					sameConfigFile.Spec.Environments[env_name] = env
				}
			}
		}

		if sameConfigFile.Spec.ConfigFilePath == "" {
			sameConfigFile.Spec.ConfigFilePath = filePath
		}

		params, _ := cmd.Flags().GetStringSlice("run-param")

		runParams := make(map[string]interface{})

		if len(sameConfigFile.Spec.Run.Parameters) > 0 {
			runParams = sameConfigFile.Spec.Run.Parameters
		}

		// override the explicitly set run parameters
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				println(fmt.Sprintf("Invalid param format %q. Expect: key=value", param))
			}
			runParams[parts[0]] = parts[1]
		}

		compiledDir, _, err := CompileFile(target, *sameConfigFile, persistTempFiles)
		if err != nil {
			return err
		}
		if !persistTempFiles {
			defer os.Remove(compiledDir)
		}
		return nil
	},
}

func getTemporaryCompileDirectory() (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "SAME-compile-*")
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory to compile in: %v", err)
	}

	return dir, nil
}

func writeSameConfigFile(compiledDir string, sameConfigFile loaders.SameConfig) error {
	sameConfigFileYaml, err := yaml.Marshal(&sameConfigFile.Spec)
	if err != nil {
		return fmt.Errorf("error marshaling same config file: %v", err.Error())
	}
	err = os.WriteFile(path.Join(compiledDir, "same.yaml"), []byte(sameConfigFileYaml), 0400)
	if err != nil {
		return fmt.Errorf("error writing root.py file: %v", err.Error())
	}

	if err != nil {
		return fmt.Errorf("error writing same.yaml file to %v: %v", compiledDir, err)
	}

	return nil
}

func writeRootFile(compiledDir string, rootFileContents string) error {
	file_to_write := path.Join(compiledDir, "root.py")
	logrus.Tracef("File: %v\n", file_to_write)

	err = os.WriteFile(file_to_write, []byte(rootFileContents), 0400)
	if err != nil {
		return fmt.Errorf("Error writing root.py file: %v", err.Error())
	}

	return nil
}

func checkExecutableAndFile(sameConfigFile loaders.SameConfig) (string, string, error) {
	jupytextExecutable, err := exec.LookPath("jupytext")
	if err != nil {
		return "", "", fmt.Errorf("could not find 'jupytext'. Please run 'python3 -m pip install jupytext'. You may also need to add it to your path by executing: export PATH=$PATH:$HOME/.local/bin")
	}

	notebookRootDir := filepath.Dir(sameConfigFile.Spec.ConfigFilePath)
	notebookFilePath, err := utils.ResolveLocalFilePath(filepath.Join(notebookRootDir, sameConfigFile.Spec.Pipeline.Package))
	if err != nil {
		return "", "", fmt.Errorf("program_compile.go: could not find pipeline definition specified in SAME program: %v", notebookFilePath)
	}

	return jupytextExecutable, notebookFilePath, nil

}

func CompileFile(target string, sameConfigFile loaders.SameConfig, persistTempFiles bool) (compileDirectory string, updatedSameConfig loaders.SameConfig, err error) {
	var c = utils.GetCompileFunctions()
	jupytextExecutablePath, notebookFilePath, err := checkExecutableAndFile(sameConfigFile)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	if sameConfigFile.Spec.Metadata.Name == "" {
		return "", loaders.SameConfig{}, fmt.Errorf("no experiment name detected in Metadata.Name")
	}

	convertedText, err := c.ConvertNotebook(jupytextExecutablePath, notebookFilePath)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	foundSteps, err := c.FindAllSteps(convertedText)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	aggregatedSteps, err := c.CombineCodeSlicesToSteps(foundSteps)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	compiledDir, err := getTemporaryCompileDirectory()
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	packagesBySteps, err := c.WriteStepFiles(target, compiledDir, aggregatedSteps)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	for stepName, packageList := range packagesBySteps {
		thisCodeBlock := aggregatedSteps[stepName]
		if thisCodeBlock.PackagesToInstall == nil {
			thisCodeBlock.PackagesToInstall = make(map[string]string)
		}
		for packageString := range packageList {
			thisCodeBlock.PackagesToInstall[packageString] = ""
		}
		aggregatedSteps[stepName] = thisCodeBlock
	}

	rootFileContents, err := c.CreateRootFile(target, aggregatedSteps, sameConfigFile)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	err = writeRootFile(compiledDir, rootFileContents)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}

	sameConfigFile.Spec.Pipeline.Package = filepath.Join(compiledDir, "root.py")
	err = writeSameConfigFile(compiledDir, sameConfigFile)
	if err != nil {
		return "", loaders.SameConfig{}, err
	}
	updatedSameConfig = sameConfigFile

	fmt.Printf("Compilation complete! In order to upload, go to this directory (%v) and execute 'same program run'.\n", compiledDir)
	return compiledDir, updatedSameConfig, err

}

func init() {
	programCmd.AddCommand(compileProgramCmd)

	compileProgramCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml').")
	compileProgramCmd.Flags().Bool("persist-temp-files", false, "Persist the temporary compilation files.")
	compileProgramCmd.Flags().StringP("target", "t", "kubeflow", "Enter one of 'kubeflow', 'aml'. Defaults to: kubeflow")
	compileProgramCmd.Flags().String("image-pull-secret-server", "", "Image pull server for any private repos (only one server currently supported for all private repos)")
	compileProgramCmd.Flags().String("image-pull-secret-username", "", "Image pull username for any private repos (only one username currently supported for all private repos)")
	compileProgramCmd.Flags().String("image-pull-secret-password", "", "Image pull password for any private repos (only one password currently supported for all private repos)")
	compileProgramCmd.Flags().String("image-pull-secret-email", "", "Image pull email for any private repos (only one email currently supported for all private repos)")

}
