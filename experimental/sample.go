package main

import (
	"fmt"
	"log"
	user "os/user"

	"github.com/azure-octo/same-cli/pkg/utils"
)

// Settings default user setting
type Settings struct {
	// Repo is --plugin-repo
	Repo string `yaml:"repo"`
	// UseKubectl use kubectl instead of k3s
	UseKubectl bool `yaml:"use-kubectl"`
}

type Config struct {
	Kind                string          `yaml:"kind"`
	TargetCustomization []TargetCustoms `yaml:"targetCustomizations,flow"`
}

//PluginGroup represent the structure for the inline plugins
type PluginGroup struct {
	Repo string `yaml:"repo,omitempty"`
	Name string `yaml:"name,omitempty"`
}

//TargetCustoms represent the single customization group
type TargetCustoms struct {
	Name              string        `yaml:"name"`
	Enabled           bool          `yaml:"enabled"`
	Type              string        `yaml:"type"`
	Config            string        `yaml:"config"`
	ClusterName       string        `yaml:"clusterName"`
	ClusterDeployment string        `yaml:"clusterDeployment"`
	ClusterStart      string        `yaml:"clusterStart,omitempty"`
	Spec              Spec          `yaml:"spec,omitempty"`
	Plugins           []PluginGroup `yaml:"plugins,flow"`
}

type Spec struct {
	Wsl             string `yaml:"wsl,omitempty"`
	Mac             string `yaml:"mac,omitempty"`
	Linux           string `yaml:"linux,omitempty"`
	Windows         string `yaml:"windows,omitempty"`
	cloudType       string `yaml:"cloudType,omitempty"`
	cloudNodes      string `yaml:"cloudNodes,omitempty"`
	cloudSecretPath string `yaml:"cloudSecretPath,omitempty"`
}

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

	// kfpconfig := *cmd.NewKFPConfig()
	// pClient, _ := api_server.NewPipelineClient(kfpconfig, false)

	// pipelineClientParams := pipeline_service.NewListPipelinesParams()

	// arr, _ := pClient.ListAll(pipelineClientParams, 100)
	// for _, s := range arr {
	// 	fmt.Println(s.Name)
	// }

	// var c Config

	// yamlFile, err := ioutil.ReadFile("/home/daaronch/same-cli/test/testdata/k3ai/default.yaml")
	// if err != nil {
	// 	log.Printf("yamlFile.Get err   #%v ", err)
	// }
	// err = yaml.Unmarshal(yamlFile, &c)
	// if err != nil {
	// 	log.Fatalf("Unmarshal: %v", err)
	// }

	dockerGroupId, err := user.LookupGroup("docker")

	if _, ok := err.(user.UnknownGroupError); ok {
		message := fmt.Errorf("could not find the group 'docker' on your system. This is required to run.")
		log.Fatal(message)
	} else if err != nil {
		message := fmt.Errorf("unknown error while trying to retrieve list of groups on your system. Sorry that's all we know: %v", err)
		log.Fatal(message)
	}

	a, _ := user.Current()
	allGroups, err := a.GroupIds()
	if err != nil {
		message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
		log.Fatal(message)
	}

	if !utils.ContainsString(allGroups, dockerGroupId.Gid) {
		message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
		log.Fatal(message)
	}
}
