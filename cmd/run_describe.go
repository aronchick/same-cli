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
	"os"
	"text/template"
	"time"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/azure-octo/same-cli/pkg/infra"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"

	"github.com/spf13/cobra"
)

var describeRunCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes a single SAME program run",
	Long:  `Describes a single SAME program run.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := infra.GetDependencyCheckers(cmd, args).CheckDependenciesInstalled(cmd); err != nil {
			if utils.PrintErrorAndReturnExit(cmd, "Failed during dependency checks: %v", err) {
				return err
			}
		}

		runId, err := cmd.Flags().GetString("run-id")
		if err != nil {
			return err
		}

		run, wf, err := GetRun(runId)
		if err != nil {
			return err
		}
		return prettyPrint(run, wf)
	},
}

func pipelineVersionID(run *run_model.APIRun) string {
	for _, ref := range run.ResourceReferences {
		if ref.Key.Type == run_model.APIResourceTypePIPELINEVERSION {
			return ref.Key.ID
		}
	}
	return ""
}

func pipelineVersionName(run *run_model.APIRun) string {
	versionID := pipelineVersionID(run)
	if err != nil {
		return err.Error()
	}
	version, err := GetPipelineVersion(versionID)
	if err != nil {
		return version.Name
	}
	return ""
}

func formatDate(t strfmt.DateTime) string {
	return time.Time(t).Format(time.RFC1123)
}

func prettyPrint(run *run_model.APIRunDetail, wf *v1alpha1.Workflow) error {
	funcs := map[string]interface{}{
		"PipelineVersionID":   pipelineVersionID,
		"PipelineVersionName": pipelineVersionName,
		"FormatDate":          formatDate,
	}
	runInfoTmpl := `Name:           {{ .Run.Name }}
ID:             {{ .Run.ID }}
Pipeline:
    Name:       {{ .Run.PipelineSpec.PipelineName }}
    Version:    {{ PipelineVersionName .Run }}
    VersionID:  {{ PipelineVersionID .Run }}
Parameters:
  {{- with .Run.PipelineSpec.Parameters }}{{- range . }}
    {{.Name}}:{{"\t"}}{{.Value}}
  {{- end }}{{- end }}
Created:        {{ FormatDate .Run.CreatedAt }}
Finished:       {{ FormatDate .Run.FinishedAt }}
Status:         {{ .Run.Status }}
Error:          {{ .Run.Error }}
Metrics:
  {{- with .Run.Metrics }}{{- range . }}
    {{.Name}}:{{"\t"}}{{.NumberValue}}
  {{- end }}{{- end }}
`
	t := template.Must(template.New("Run Detail").Funcs(funcs).Parse(runInfoTmpl))
	return t.Execute(os.Stdout, run)
}

func init() {
	describeRunCmd.Flags().StringP("run-id", "r", "", "The SAME run ID")
	_ = describeRunCmd.MarkFlagRequired("run-id")
	runCmd.AddCommand(describeRunCmd)
}
