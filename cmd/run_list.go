/*
Copyright © 2021 The SAME author.

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
	"os"
	"sort"
	"text/tabwriter"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var listRunCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all SAME runs for a given program",
	Long:  `Lists all SAME runs for a given program.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(cmd); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Failed during dependency checks: %v", err) {
				return err
			}
		}

		// Load config file. Explicit parameters take precedent over config file.
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		sameConfigFilePath, err := getConfigFilePath(filePath)
		if err != nil {
			log.Errorf("could not resolve SAME config file path: %v", err)
			return err
		}
		sameConfigFile, err := loaders.LoadSAME(sameConfigFilePath)
		if err != nil {
			log.Errorf("could not load SAME config file: %v", err)
			return err
		}
		programName := sameConfigFile.Spec.Pipeline.Name
		if programNameFlagValue, _ := cmd.Flags().GetString("program-name"); programNameFlagValue != "" {
			programName = programNameFlagValue
		}

		pipeline, err := FindPipelineByName(programName)
		if err != nil {
			return err
		}
		versions, err := ListPipelineVersions(pipeline.ID)
		if err != nil {
			return err
		}
		allRuns := []*run_model.APIRun{}

		pipelineVersionLookupMap := make(map[string]string)
		for _, version := range versions {
			pipelineVersionLookupMap[version.ID] = version.Name
			runs, err := ListRunsForPipelineVersion(version.ID)
			if err != nil {
				return err
			}
			allRuns = append(allRuns, runs...)
		}
		prettyPrintRunList(allRuns, pipelineVersionLookupMap)
		return nil
	},
}

// NewTabWriter returns a *tabwriter.Writer with some visually
// pleasing settings.
func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout,
		0,   // minWidth
		0,   // tabWidth
		3,   // padding
		' ', // padChar
		0,   // default formatting options */
	)
}

func prettyPrintRunList(runs []*run_model.APIRun, pipelineVersionLookupMap map[string]string) {
	metricNames := getMetricsNames(runs)
	w := NewTabWriter()
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s", "ID", "NAME", "PIPELINEVERSION", "CREATED", "STATUS")
	for _, metricName := range metricNames {
		fmt.Fprintf(w, "\t%s", metricName)
	}
	fmt.Fprintln(w)
	for _, run := range runs {
		pipelineVersionID := ""
		pipelineID := ""
		versionName := ""
		for _, ref := range run.ResourceReferences {
			if ref.Key.Type == run_model.APIResourceTypePIPELINEVERSION {
				pipelineVersionID = ref.Key.ID
			} else if ref.Key.Type == run_model.APIResourceTypePIPELINE {
				pipelineID = ref.Key.ID
			}
		}
		if pipelineVersionID != "" {
			versionName = pipelineVersionLookupMap[pipelineVersionID]
		} else {
			versionName = pipelineID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s", run.ID, run.Name, versionName, run.CreatedAt, run.Status)
		for _, metricName := range metricNames {
			if metricValue, exist := getMetric(run, metricName); exist {
				fmt.Fprintf(w, "\t%.4f", metricValue)
			} else {
				fmt.Fprint(w, "\t-")
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

func getMetricsNames(runs []*run_model.APIRun) []string {
	metricNames := make(map[string]bool)
	sorted := []string{}
	for _, run := range runs {
		for _, metric := range run.Metrics {
			if _, exist := metricNames[metric.Name]; !exist {
				metricNames[metric.Name] = true
				sorted = append(sorted, metric.Name)
			}
		}
	}
	sort.Strings(sorted)
	return sorted
}

func getMetric(run *run_model.APIRun, metricName string) (value float64, exist bool) {
	for _, metric := range run.Metrics {
		if metricName == metric.Name {
			return metric.NumberValue, true
		}
	}
	return 0, false
}

func init() {
	listRunCmd.Flags().StringP("experiment-name", "e", "", "The SAME Experiment name")
	listRunCmd.Flags().StringP("program-name", "n", "", "The SAME Program name")
	listRunCmd.Flags().StringP("file", "f", "same.yaml", "a SAME program file (defaults to 'same.yaml')")
	runCmd.AddCommand(listRunCmd)
}
