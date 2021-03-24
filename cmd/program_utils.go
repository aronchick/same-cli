package cmd

import (
	"fmt"
	"io/ioutil"
	netUrl "net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
	gogetter "github.com/hashicorp/go-getter"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	experimentparams "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	experimentmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_model"
	pipelineuploadparams "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	pipelineuploadmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_model"
	runparams "github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	runmodel "github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	apiclient "github.com/kubeflow/pipelines/backend/src/common/client/api_server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"

	"github.com/azure-octo/same-cli/pkg/utils"
)

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

func UploadPipeline(sameConfigFile *loaders.SameConfig, pipelineName string, pipelineDescription string) (uploadedPipeline *pipelineuploadmodel.APIPipeline, err error) {
	log.Traceln("- In program_utils.UploadPipeline")
	kfpconfig := *NewKFPConfig()

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

	uploadedPipeline, err = uploadclient.UploadFile(pipelineFilePath, uploadparams)

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

func UpdatePipeline(sameConfigFile *loaders.SameConfig, pipelineID string, pipelineVersion string) (uploadedPipelineVersion *pipelineuploadmodel.APIPipelineVersion, err error) {
	kfpconfig := *NewKFPConfig()

	uploadclient, err := apiclient.NewPipelineUploadClient(kfpconfig, false)
	if err != nil {
		log.Errorf("could not create API client for pipeline: %v", err)
		return nil, err
	}

	uploadparams := pipelineuploadparams.NewUploadPipelineVersionParams()
	uploadparams.Pipelineid = &pipelineID
	uploadparams.Name = &pipelineVersion

	// TODO: We only support local compressed pipelines (for now)
	pipelineFilePath, err := utils.ResolveLocalFilePath(sameConfigFile.Spec.Pipeline.Package)
	if err != nil {
		return nil, err
	}

	uploadedPipelineVersion, err = uploadclient.UploadPipelineVersion(pipelineFilePath, uploadparams)

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
	listOfPipelines := ListPipelines()
	for _, thisPipeline := range listOfPipelines {
		if pipelineName == thisPipeline.Name {
			return thisPipeline, nil
		}
	}
	return nil, fmt.Errorf("could not find a pipeline with the name: %v", pipelineName)
}

func ListPipelines() []*pipeline_model.APIPipeline {
	kfpconfig := *NewKFPConfig()
	pClient, _ := apiclient.NewPipelineClient(kfpconfig, false)
	pipelineClientParams := pipeline_service.NewListPipelinesParams()
	listOfPipelines, _ := pClient.ListAll(pipelineClientParams, 10000)
	return listOfPipelines
}

func ListPipelineVersions(pipelineID string) ([]*pipeline_model.APIPipelineVersion, error) {
	kfpconfig := *NewKFPConfig()
	pClient, _ := apiclient.NewPipelineClient(kfpconfig, false)
	listPipelineVersionParams := pipeline_service.NewListPipelineVersionsParams()
	pipelineType := pipeline_model.APIResourceTypePIPELINE
	listPipelineVersionParams.SetResourceKeyType((*string)(&pipelineType))
	listPipelineVersionParams.SetResourceKeyID(&pipelineID)
	sortBy := "created_at desc"
	listPipelineVersionParams.SetSortBy((*string)(&sortBy))
	listOfPipelineVersions, _, _, vErr := pClient.ListPipelineVersions(listPipelineVersionParams)
	return listOfPipelineVersions, vErr
}

func FindExperimentByName(experimentName string) (experiment *experimentmodel.APIExperiment, err error) {
	kfpconfig := *NewKFPConfig()
	eClient, _ := apiclient.NewExperimentClient(kfpconfig, false)
	experimentClientParams := experiment_service.NewListExperimentParams()
	apiExperimentType := experiment_model.APIResourceTypeEXPERIMENT
	experimentClientParams.SetResourceReferenceKeyType((*string)(&apiExperimentType))
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

	// associate run with experiment
	resourceKey := runmodel.APIResourceKey{ID: experimentID, Type: runmodel.APIResourceTypeEXPERIMENT}
	resourceRef := runmodel.APIResourceReference{
		Key:          &resourceKey,
		Relationship: runmodel.APIRelationship(runmodel.APIRelationshipOWNER),
	}
	createRunParams.Body.ResourceReferences = append(createRunParams.Body.ResourceReferences, &resourceRef)

	// fetch and specify latest pipeline version
	listOfPipelineVersions, vErr := ListPipelineVersions(pipelineID)

	if vErr != nil {
		// We found a pipeline version, so let us specify it
		latestPipelineVersionID := listOfPipelineVersions[0].ID
		resourceKey = runmodel.APIResourceKey{ID: latestPipelineVersionID, Type: runmodel.APIResourceTypePIPELINEVERSION}
		resourceRef = runmodel.APIResourceReference{
			Key:          &resourceKey,
			Relationship: runmodel.APIRelationship(runmodel.APIRelationshipOWNER),
		}
		createRunParams.Body.ResourceReferences = append(createRunParams.Body.ResourceReferences, &resourceRef)
	}

	runDetail, _, err := runclient.Create(createRunParams)

	if err != nil {
		panic(err)
	}

	return runDetail
}

// getFilePath returns a file path to the local drive of the SAME config file, or error if invalid.
// If the file is remote, it pulls from a GitHub repo.
// Expects a full file path (including the file name)
func getConfigFilePath(putativeFilePath string) (filePath string, err error) {
	// TODO: aronchick: This is all probably unnecessary. We could just swap everything out
	// for gogetter.GetFile() and punt the whole problem at it.
	// HOWEVER, that doesn't solve for when a github url has an https schema, which causes
	// gogetter to weirdly reformats he URL (dropping the repo).
	// E.g., gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml")
	// Fails with a bad response code: 404
	// and 	gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/EXAMPLE-SAME-Enabled-Data-Science-Repo/same.yaml")
	// Fails with fatal: repository 'https://github.com/SAME-Project/' not found

	isRemoteFile, err := utils.IsRemoteFilePath(putativeFilePath)

	if err != nil {
		log.Errorf("could not tell if the file was remote or not: %v", err)
		return "", err
	}

	if isRemoteFile {
		// Use the default system temp directory and a randomly generated name
		tempSameDir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Errorf("error creating a temporary directory to copy the file to (we're using the standard temporary directory from your system, so this could be an issue of the permissions this CLI is running under): %v", err)
			return "", err
		}

		// Get path to store the file to
		tempSameFile, err := ioutil.TempFile(tempSameDir, "")
		if err != nil {
			return "", fmt.Errorf("could not create temporary file in %v", tempSameDir)
		}

		configFileUri, err := netUrl.Parse(putativeFilePath)
		if err != nil {
			return "", fmt.Errorf("could not parse sameFile url: %v", err)
		}

		// TODO: Hard coding 'same.yaml' in now - should be optional
		finalUrl, err := utils.UrlToRetrive(configFileUri.String(), "same.yaml")
		if err != nil {
			message := fmt.Errorf("unable to process the url to retrieve from the provided configFileUri(%v): %v", configFileUri.String(), err)
			log.Error(message)
			return "", message
		}

		corrected_url := finalUrl.String()
		if (finalUrl.Scheme == "https") || (finalUrl.Scheme == "http") {
			log.Info("currently only support http and https on github.com because we need to prefix with git")
			corrected_url = "git::" + finalUrl.String()
		}

		log.Infof("Downloading from %v to %v", corrected_url, tempSameFile.Name())
		errGet := gogetter.GetFile(tempSameFile.Name(), corrected_url)
		if errGet != nil {
			return "", fmt.Errorf("could not download SAME file from URL '%v': %v", finalUrl.String(), errGet)
		}

		filePath = tempSameFile.Name()
	} else {
		cwd, _ := os.Getwd()
		log.Tracef("CWD: %v", cwd)
		absFilePath, _ := filepath.Abs(putativeFilePath)
		filePath, _ = gogetter.Detect(absFilePath, cwd, []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.FileDetector)})

		if !fileExists(filePath) {
			return "", fmt.Errorf("could not find sameFile at: %v\nerror: %v", putativeFilePath, err)
		}
	}
	return filePath, nil
}

func kubectlExists() (kubectlDoesExist bool, err error) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		err := fmt.Errorf("the 'kubectl' binary is not on your PATH: %v", os.Getenv("PATH"))
		return false, err
	}
	log.Tracef("'kubectl' found at %v", path)
	return true, nil
}

func fileExists(path string) (fileDoesExist bool) {
	resolvedPath, err := netUrl.Parse(path)
	if err != nil {
		log.Errorf("could not parse path '%v': %v", path, err)
		return false
	}
	_, err = os.Stat(resolvedPath.Path)
	return err == nil
}
