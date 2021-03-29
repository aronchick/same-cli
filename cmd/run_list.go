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

		experimentName, _ := cmd.Flags().GetString("experiment-name")
		experiment, err := FindExperimentByName(experimentName)
		if err != nil {
			return err
		}

		runs, err := ListRunsForExperiment(experiment.ID)
		if err == nil {
			prettyPrint(runs)
		}
		return err
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

func prettyPrint(runs []*run_model.APIRun) {
	metricNames := getMetricsNames(runs)
	w := NewTabWriter()
	fmt.Fprintf(w, "%s\t%s\t%s\t%s", "ID", "NAME", "CREATED", "STATUS")
	for _, metricName := range metricNames {
		fmt.Fprintf(w, "\t%s", metricName)
	}
	fmt.Fprintln(w)
	for _, run := range runs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s", run.ID, run.Name, run.CreatedAt, run.Status)
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
	if err := listRunCmd.MarkFlagRequired("experiment-name"); err != nil {
		log.Fatal("Cannot mark a flag as required", err)
	}
	runCmd.AddCommand(listRunCmd)
}
