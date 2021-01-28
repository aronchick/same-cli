package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-12-01/containerservice"
	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"

	experimentparams "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	experimentmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	pipelineuploadparams "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	pipelineuploadmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_model"
	runparams "github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	runmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	apiclient "github.com/kubeflow/pipelines/backend/src/common/client/api_server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spf13/viper"
)

// DeployOrUpdateAPipeline takes a sameConfig and communicates with Kubeflow to deploy a piplene
func DeployOrUpdateAPipeline(ctx context.Context, resourceGroupName string, aksCluster containerservice.ManagedCluster, sameConfig loaders.SameConfig) (err error) {
	//NYI

	// This needs to first download the KubeConfig using the GO SDK

	// Then from the sameConfig figure out:
	// * path string of the compile pipeline
	// * params to use for a pipeline run

	return fmt.Errorf("method 'DeployOrUpdateAPipeline' has not yet been implemented")
}

// COMPILEDPIPELINE : Temporary placeholder
var COMPILEDPIPELINE = "pipeline.tar.gz"

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
func CreateRunFromCompiledPipeline(filePath string, pipelineName string, pipelineDescription string, experimentName string, experimentDescription string, runName string, runDescription string, runParams map[string]string) string {

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

	uploadedPipeline := UploadPipeline(filePath, pipelineName, pipelineDescription)
	createdExperiment := CreateExperiment(experimentName, experimentDescription)
	runDetails := CreateRun(runName, uploadedPipeline.ID, createdExperiment.ID, runDescription, runParams)

	fmt.Println("Pipeline ID: " + uploadedPipeline.ID)
	fmt.Println("Run: " + runDetails.Run.ID + ":" + runDetails.Run.Status)

	return runDetails.Run.ID
}

func UploadPipeline(filePath string, pipelineName string, pipelineDescription string) *pipelineuploadmodel.APIPipeline {
	kfpconfig := *NewKFPConfig()

	uploadclient, err := apiclient.NewPipelineUploadClient(kfpconfig, true)
	if err != nil {
		panic(err)
	}
	uploadparams := pipelineuploadparams.NewUploadPipelineParams()
	uploadparams.Name = &pipelineName
	uploadparams.Description = &pipelineDescription
	uploadedPipeline, err := uploadclient.UploadFile(filePath, uploadparams)

	if err != nil {
		panic(err)
	}

	viper.Set("activepipeline", uploadedPipeline.ID)
	viper.WriteConfig()

	return uploadedPipeline
}

func CreateExperiment(experimentName string, experimentDescription string) *experimentmodel.APIExperiment {
	kfpconfig := *NewKFPConfig()
	kfpconfig.Namespace()
	experimentclient, err := apiclient.NewExperimentClient(kfpconfig, true)
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

	runclient, err := apiclient.NewRunClient(kfpconfig, true)
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

// func main() {
// 	// Consider adding param

// 	CreateRunFromCompiledPipeline(COMPILEDPIPELINE, "", "", "", "", "", "", make(map[string]string))
// }
