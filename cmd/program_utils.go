package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	apiclient "github.com/kubeflow/pipelines/backend/src/common/client/api_server"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/utils"
)

func UploadPipeline(sameConfigFile *loaders.SameConfig, pipelineName string, pipelineDescription string, persistTemporaryFiles bool) (uploadedPipeline *pipeline_upload_model.APIPipeline, err error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}

	uploadclient, err := apiclient.NewPipelineUploadClient(kfpconfig, false)
	if err != nil {
		log.Errorf("could not create API client for pipeline: %v", err)
		return nil, err
	}

	uploadparams := pipeline_upload_service.NewUploadPipelineParams()
	uploadparams.Name = &pipelineName
	uploadparams.Description = &pipelineDescription

	if strings.HasSuffix(strings.TrimSpace(sameConfigFile.Spec.Pipeline.Package), ".ipynb") {
		tempCompileDir, updatedSameConfig, err := CompileFile(*sameConfigFile, true)
		if err != nil {
			return nil, err
		}
		if !persistTemporaryFiles {
			defer os.Remove(tempCompileDir)
		}
		sameConfigFile.Spec.ConfigFilePath = updatedSameConfig.Spec.ConfigFilePath
		sameConfigFile.Spec.Pipeline.Package = updatedSameConfig.Spec.Pipeline.Package
	}

	pipelinePath, _ := filepath.Abs(sameConfigFile.Spec.Pipeline.Package)
	pipelineFilePath, err := utils.CompileForKFP(pipelinePath)
	if err != nil {
		return nil, err
	}

	uploadedPipeline, err = uploadclient.UploadFile(pipelineFilePath, uploadparams)
	if !persistTemporaryFiles {
		defer os.Remove(pipelineFilePath)
	}

	if err != nil {
		// It's not an error we know about, and we couldn't find the pipeline we uploaded, so assuming it didn't get uploaded
		log.Errorf("deploy_or_update_a_pipeline.go: could not upload pipeline: %v", err)
		return nil, err
	} else {
		if uploadedPipeline == nil {
			log.Fatalf("both uploadedPipeline and err are nil, unclear how you got here.")
		}
	}
	return uploadedPipeline, nil
}

func UpdatePipeline(sameConfigFile *loaders.SameConfig, pipelineID string, pipelineVersion string, persistTemporaryFiles bool) (uploadedPipelineVersion *pipeline_upload_model.APIPipelineVersion, err error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}

	uploadclient, err := apiclient.NewPipelineUploadClient(kfpconfig, false)
	if err != nil {
		log.Errorf("could not create API client for pipeline: %v", err)
		return nil, err
	}

	uploadparams := pipeline_upload_service.NewUploadPipelineVersionParams()
	uploadparams.Pipelineid = &pipelineID
	uploadparams.Name = &pipelineVersion

	if strings.HasSuffix(strings.TrimSpace(sameConfigFile.Spec.Pipeline.Package), ".ipynb") {
		tempCompileDir, updatedSameConfig, err := CompileFile(*sameConfigFile, true)
		if err != nil {
			return nil, err
		}
		if !persistTemporaryFiles {
			defer os.Remove(tempCompileDir)
		}
		sameConfigFile.Spec.ConfigFilePath = updatedSameConfig.Spec.ConfigFilePath
		sameConfigFile.Spec.Pipeline.Package = updatedSameConfig.Spec.Pipeline.Package
	}
	pipelinePath, _ := filepath.Abs(sameConfigFile.Spec.Pipeline.Package)
	pipelineFilePath, err := utils.CompileForKFP(pipelinePath)
	if err != nil {
		return nil, err
	}

	uploadedPipelineVersion, err = uploadclient.UploadPipelineVersion(pipelineFilePath, uploadparams)
	if !persistTemporaryFiles {
		defer os.Remove(pipelineFilePath)
	}

	if err != nil {
		// It's not an error we know about, so assuming the version didn't get uploaded
		log.Errorf("deploy_or_update_a_pipeline.go: could not upload pipeline version: %v", err)
		return nil, err
	} else {
		if uploadedPipelineVersion == nil {
			log.Fatalf("both uploadedPipelineVersion and err are nil, unclear how you got here.")
		}
	}

	return uploadedPipelineVersion, nil
}

func FindPipelineByName(pipelineName string) (uploadedPipeline *pipeline_model.APIPipeline, err error) {
	listOfPipelines, err := ListPipelines()
	if err != nil {
		return nil, err
	}
	for _, thisPipeline := range listOfPipelines {
		if pipelineName == thisPipeline.Name {
			return thisPipeline, nil
		}
	}
	return nil, fmt.Errorf("could not find a pipeline with the name: %v", pipelineName)
}

func ListPipelines() ([]*pipeline_model.APIPipeline, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	pClient, _ := apiclient.NewPipelineClient(kfpconfig, false)

	resourceType := pipeline_model.APIResourceTypeNAMESPACE
	pipelineClientParams := pipeline_service.NewListPipelinesParams().WithResourceReferenceKeyType((*string)(&resourceType))
	listOfPipelines, _ := pClient.ListAll(pipelineClientParams, 10000)
	return listOfPipelines, nil
}

func ListRunsForExperiment(experimentID string) ([]*run_model.APIRun, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	client, _ := apiclient.NewRunClient(kfpconfig, false)
	resourceType := run_model.APIResourceTypeEXPERIMENT
	params := run_service.NewListRunsParams().WithResourceReferenceKeyType((*string)(&resourceType)).WithResourceReferenceKeyID(&experimentID)
	return client.ListAll(params, 10000)
}

func ListRunsForPipelineVersion(pipelineVersionId string) ([]*run_model.APIRun, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	client, _ := apiclient.NewRunClient(kfpconfig, false)
	resourceType := run_model.APIResourceTypePIPELINEVERSION
	params := run_service.NewListRunsParams().WithResourceReferenceKeyID(&pipelineVersionId).WithResourceReferenceKeyType((*string)(&resourceType))
	return client.ListAll(params, 10000)
}

func GetRun(runId string) (*run_model.APIRunDetail, *v1alpha1.Workflow, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, nil, err
	}
	client, _ := apiclient.NewRunClient(kfpconfig, false)
	params := run_service.NewGetRunParams().WithRunID(runId)
	return client.Get(params)
}

func GetPipelineVersion(versionID string) (*pipeline_model.APIPipelineVersion, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	client, _ := apiclient.NewPipelineClient(kfpconfig, false)
	params := pipeline_service.NewGetPipelineVersionParams().WithVersionID(versionID)
	return client.GetPipelineVersion(params)
}

func ListPipelineVersions(pipelineID string) ([]*pipeline_model.APIPipelineVersion, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	pClient, _ := apiclient.NewPipelineClient(kfpconfig, false)
	pipelineType := pipeline_model.APIResourceTypePIPELINE
	listPipelineVersionParams := pipeline_service.NewListPipelineVersionsParams().WithResourceKeyType((*string)(&pipelineType)).WithResourceKeyID(&pipelineID)
	sortBy := "created_at desc"
	listPipelineVersionParams.SetSortBy((*string)(&sortBy))
	listOfPipelineVersions, _, _, vErr := pClient.ListPipelineVersions(listPipelineVersionParams)
	return listOfPipelineVersions, vErr
}

func FindExperimentByName(experimentName string) (experiment *experiment_model.APIExperiment, err error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	eClient, _ := apiclient.NewExperimentClient(kfpconfig, false)
	apiExperimentType := experiment_model.APIResourceTypeEXPERIMENT
	experimentClientParams := experiment_service.NewListExperimentParams().WithResourceReferenceKeyType((*string)(&apiExperimentType))
	listOfExperiments, _, _, err := eClient.List(experimentClientParams)
	if err != nil {
		return nil, err
	}
	for _, thisExperiment := range listOfExperiments {
		if experimentName == thisExperiment.Name {
			return thisExperiment, nil
		}
	}
	return nil, fmt.Errorf("could not find an experiment with the name: %v", experimentName)
}

func CreateExperiment(experimentName string, experimentDescription string) (*experiment_model.APIExperiment, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}
	experimentclient, err := apiclient.NewExperimentClient(kfpconfig, false)
	if err != nil {
		panic(err)
	}
	createExperimentParams := experiment_service.NewCreateExperimentParams()
	expBody := experiment_model.APIExperiment{
		Name:        experimentName,
		Description: experimentDescription,
	}
	createExperimentParams.Body = &expBody
	createdExperiment, err := experimentclient.Create(createExperimentParams)

	if err != nil {
		panic(err)
	}

	return createdExperiment, nil
}

func CreateRun(runName string, pipelineID string, pipelineVersionID string, experimentID string, runDescription string, runParameters map[string]string) (*run_model.APIRunDetail, error) {
	kfpconfig, err := utils.NewKFPConfig()
	if err != nil {
		return nil, err
	}

	runParams := make([]*run_model.APIParameter, 0)

	for name, value := range runParameters {
		runParams = append(runParams, &run_model.APIParameter{Name: name, Value: value})
	}

	runclient, err := apiclient.NewRunClient(kfpconfig, false)
	if err != nil {
		panic(err)
	}

	createRunParams := run_service.NewCreateRunParams()
	runBody := run_model.APIRun{
		Name:        runName,
		Description: runDescription,
		PipelineSpec: &run_model.APIPipelineSpec{
			Parameters: runParams,
			PipelineID: pipelineID,
		},
	}
	createRunParams.Body = &runBody

	// associate run with experiment
	resourceKey := run_model.APIResourceKey{ID: experimentID, Type: run_model.APIResourceTypeEXPERIMENT}
	resourceRef := run_model.APIResourceReference{
		Key:          &resourceKey,
		Relationship: run_model.APIRelationship(run_model.APIRelationshipOWNER),
	}
	createRunParams.Body.ResourceReferences = append(createRunParams.Body.ResourceReferences, &resourceRef)

	if pipelineVersionID != "" {
		// We want to run a specific pipeline version, so let us specify it
		versionResourceKey := run_model.APIResourceKey{ID: pipelineVersionID, Type: run_model.APIResourceTypePIPELINEVERSION}
		versionResourceRef := run_model.APIResourceReference{
			Key:          &versionResourceKey,
			Relationship: run_model.APIRelationship(run_model.APIRelationshipCREATOR),
		}
		createRunParams.Body.ResourceReferences = append(createRunParams.Body.ResourceReferences, &versionResourceRef)
	}

	runDetail, _, err := runclient.Create(createRunParams)

	if err != nil {
		return nil, err
	}

	return runDetail, nil
}
