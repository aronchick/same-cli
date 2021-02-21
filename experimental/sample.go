package main

import (
	"fmt"

	"github.com/azure-octo/same-cli/cmd"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/src/common/client/api_server"
)

func main() {
	// os.Setenv("PATH", "/sbin")
	// path, err := exec.LookPath("kubectl")
	// if err != nil {
	// 	log.Fatal("installing kubectl is in your future")
	// }
	// fmt.Printf("fortune is available at %s\n", path)

	// tempFile, _ := ioutil.TempFile("", "")
	// fmt.Printf("file: %v\n", tempFile.Name())
	// d, err := gogetter.Detect("https://github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml", "", []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.GitLabDetector), new(gogetter.BitBucketDetector), new(gogetter.GCSDetector)})
	// d, err := gogetter.Detect("github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml", ".", []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.GitLabDetector), new(gogetter.BitBucketDetector), new(gogetter.GCSDetector)})
	// d, _ := gogetter.Detect("github/SAME-Project/Sample-SAME-Data-Science/same.yaml", "/", []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.GitLabDetector), new(gogetter.BitBucketDetector), new(gogetter.GCSDetector), new(gogetter.FileDetector)})
	// cwd, _ := os.Getwd()
	//d, _ := gogetter.Detect("same.yaml", cwd, []gogetter.Detector{new(gogetter.GitHubDetector), new(gogetter.GitLabDetector), new(gogetter.BitBucketDetector), new(gogetter.GCSDetector), new(gogetter.FileDetector)})
	// err := gogetter.GetFile(tempFile.Name(), "https://github.com/SAME-Project/Sample-SAME-Data-Science/same.yaml")
	// err := gogetter.GetFile(tempFile.Name(), d)
	// fmt.Printf("d: %v\n", d)
	// fmt.Printf("err: %v", err)

	// d, _ := os.Getwd()
	// // s, _ := getter.Detect("file:///home/daaronch/same-cli/same.yaml", d, []getter.Detector{new(getter.FileDetector)})
	// s, _ := getter.Detect("https://github.com/dapr/dapr/same.yaml", d, getter.Detectors)
	// u, _ := url.ParseRequestURI(s)
	// sameConfig, err := loaders.LoadSAMEConfig(u.Path)
	// fmt.Printf("same u: %v\n", u.String())
	// fmt.Printf("same err: %v\n", err)
	// _ = sameConfig

	// a, b := os.Stat("/home/daaronch/same-cli/test/testdata/badpipeline.yaml")
	// _ = a
	// _ = b

	kfpconfig := *cmd.NewKFPConfig()
	pClient, _ := api_server.NewPipelineClient(kfpconfig, false)

	pipelineClientParams := pipeline_service.NewListPipelinesParams()

	arr, _ := pClient.ListAll(pipelineClientParams, 100)
	for _, s := range arr {
		fmt.Println(s.Name)
	}
}
