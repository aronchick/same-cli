package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	experimentparams "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	experimentmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_model"
	pipelineuploadparams "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	runparams "github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	runmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	"github.com/kubeflow/pipelines/backend/src/common/client/api_server"
	apiclient "github.com/kubeflow/pipelines/backend/src/common/client/api_server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/mocks"
	"github.com/azure-octo/same-cli/pkg/utils"
	"github.com/spf13/viper"
)

// COMPILEDPIPELINE : Temporary placeholder
var COMPILEDPIPELINE = "pipeline.tar.gz"
var configWriter ConfigFileIO

type ConfigFileIO interface {
	ConfigWriter(viper.Viper) error
}

type LiveConfigFileIO struct {
}

func (lcfio *LiveConfigFileIO) ConfigWriter(viper viper.Viper) error {
	err = viper.WriteConfig()
	if err != nil {
		log.Fatalf("error while writing file flag using viper as required: %v", err)
	}
	return err
}

// NewKFPConfig : Create Kubernetes API config compatible with Pipelines from KubeConfig
func NewKFPConfig() *clientcmd.ClientConfig {
	// Load kubeconfig
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		panic("Could not find kube config!")
	}

	kubebytes, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		panic(err)
	}
	// uses kubeconfig current context
	config, err := clientcmd.NewClientConfigFromBytes(kubebytes)
	if err != nil {
		panic(err)
	}

	return &config
}

// CreateRunFromCompiledPipeline : Create and run a pipeline
func CreateRunFromCompiledPipeline(sameConfigFile *loaders.SameConfig, pipelineName string, pipelineDescription string, experimentName string, experimentDescription string, runName string, runDescription string, runParams map[string]string) string {

	if pipelineName == "" {
		pipelineName = "New pipeline"
	}

	if pipelineDescription == "" {
		pipelineDescription = "Description of a new pipeline."
	}

	if experimentName == "" {
		experimentName = "Default"
	}

	if experimentDescription == "" {
		experimentDescription = "Description of a new experiment."
	}

	if runName == "" {
		runName = "New run"
	}

	if runDescription == "" {
		runDescription = "Description of a new run."
	}

	uploadedPipeline, _ := UploadPipeline(sameConfigFile, pipelineName, pipelineDescription)
	createdExperiment := CreateExperiment(experimentName, experimentDescription)
	runDetails := CreateRun(runName, uploadedPipeline.ID, createdExperiment.ID, runDescription, runParams)

	fmt.Println("Pipeline ID: " + uploadedPipeline.ID)
	fmt.Println("Run: " + runDetails.Run.ID + ":" + runDetails.Run.Status)

	return runDetails.Run.ID
}

func UploadPipeline(sameConfigFile *loaders.SameConfig, pipelineName string, pipelineDescription string) (uploadedPipeline *pipeline_model.APIPipeline, err error) {
	kfpconfig := *NewKFPConfig()

	configWriter = &LiveConfigFileIO{}
	if os.Getenv("TEST_PASS") == "1" {
		configWriter = &mocks.MockConfigFileIO{}
	}

	uploadclient, err := apiclient.NewPipelineUploadClient(kfpconfig, false)
	if err != nil {
		log.Errorf("could not create API client for pipeline: %v", err)
		return nil, err
	}

	uploadparams := pipelineuploadparams.NewUploadPipelineParams()
	uploadparams.Name = &pipelineName
	uploadparams.Description = &pipelineDescription

	// TODO: We only support local compressed pipelines (for now)
	pipelineFilePath, err := utils.ResolveLocalFilePath(sameConfigFile.Spec.Pipeline.Package)
	if err != nil {
		return nil, err
	}

	// uploadedPipeline will always be nil until we fix the swagger implementation
	_, err = uploadclient.UploadFile(pipelineFilePath, uploadparams)

	if err != nil {
		if strings.Contains(err.Error(), "can be resolved by supporting TextUnmarshaler interface") {
			log.Warn("Skipping error on return due to lack of support in go-swagger for TextUnmarshaler interface (it doesn't support blank response.body)")
		} else {
			log.Errorf("deploy_or_update_a_pipeline.go: failed to upload pipeline: %v", err)
			return nil, err
		}
	}

	// TODO: The below is a GROSS HACK. go-swagger produces the following error for everything with an empty body:
	// s:"&{0 [] } (*pipeline_upload_model.APIStatus) is not supported by the TextConsumer, can be resolved by supporting TextUnmarshaler interface"
	// This does not indicate an error (we think), so I'm bailing out.
	// We SHOULD fix go-swagger so it doesn't produce this.

	// Commenting out completely and swallowing the previous error. We'll just crawl the KFP endpoint and look for the name.
	// if strings.Contains(err.Error(), "supporting TextUnmarshaler interface") {
	uploadedPipeline, err = findPipeline(kfpconfig, *uploadparams.Name)
	if err != nil {
		// It's not an error we know about, and we couldn't find the pipeline we uploaded, so assuming it didn't get uploaded
		log.Errorf("deploy_or_update_a_pipeline.go: could not search for a pipeline: %v", err)
		return nil, err
	} else if uploadedPipeline.ID == "" {
		log.Errorf("deploy_or_update_a_pipeline.go: returned with no error, but we couldn't resolve it to an ID: %v", err)
		return nil, err
	}

	viper.Set("activepipeline", uploadedPipeline.ID)
	err = configWriter.ConfigWriter(viper.Viper{})
	if err != nil {
		return nil, err
	}

	return uploadedPipeline, nil
}

func findPipeline(kfpconfig clientcmd.ClientConfig, pipelineName string) (uploadedPipeline *pipeline_model.APIPipeline, err error) {
	pClient, _ := api_server.NewPipelineClient(kfpconfig, false)

	pipelineClientParams := pipeline_service.NewListPipelinesParams()

	listOfPipelines, _ := pClient.ListAll(pipelineClientParams, 100)
	for _, thisPipeline := range listOfPipelines {
		if pipelineName == thisPipeline.Name {
			return thisPipeline, nil
		}
	}
	return nil, fmt.Errorf("could not find a pipeline with the name: %v", pipelineName)
}

func CreateExperiment(experimentName string, experimentDescription string) *experimentmodel.APIExperiment {
	kfpconfig := *NewKFPConfig()
	experimentclient, err := apiclient.NewExperimentClient(kfpconfig, false)
	if err != nil {
		panic(err)
	}
	createExperimentParams := experimentparams.NewCreateExperimentParams()
	expBody := experimentmodel.APIExperiment{
		Name:        experimentName,
		Description: experimentDescription,
	}
	createExperimentParams.Body = &expBody
	createdExperiment, err := experimentclient.Create(createExperimentParams)

	if err != nil {
		panic(err)
	}

	return createdExperiment
}

func CreateRun(runName string, pipelineID string, experimentID string, runDescription string, runParameters map[string]string) *runmodel.APIRunDetail {
	kfpconfig := *NewKFPConfig()

	runParams := make([]*runmodel.APIParameter, 0)

	for name, value := range runParameters {
		runParams = append(runParams, &runmodel.APIParameter{Name: name, Value: value})
	}

	runclient, err := apiclient.NewRunClient(kfpconfig, false)
	if err != nil {
		panic(err)
	}

	createRunParams := runparams.NewCreateRunParams()
	runBody := runmodel.APIRun{
		Name:        runName,
		Description: runDescription,
		PipelineSpec: &runmodel.APIPipelineSpec{
			Parameters: runParams,
			PipelineID: pipelineID,
		},
	}
	createRunParams.Body = &runBody

	resourceKey := runmodel.APIResourceKey{ID: experimentID, Type: runmodel.APIResourceTypeEXPERIMENT}
	resourceRef := runmodel.APIResourceReference{
		Key:          &resourceKey,
		Relationship: runmodel.APIRelationship(runmodel.APIRelationshipOWNER),
	}

	createRunParams.Body.ResourceReferences = append(createRunParams.Body.ResourceReferences, &resourceRef)

	runDetail, _, err := runclient.Create(createRunParams)

	if err != nil {
		panic(err)
	}

	return runDetail
}
