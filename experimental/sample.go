// THIS FILE IS NOT FOR PRODUCTION USE OR INCLUSION IN ANY PACKAGE
// It is a convient place to add libraries from the rest of the

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// // Settings default user setting
// type Settings struct {
// 	// Repo is --plugin-repo
// 	Repo string `yaml:"repo"`
// 	// UseKubectl use kubectl instead of k3s
// 	UseKubectl bool `yaml:"use-kubectl"`
// }

// type Config struct {
// 	Kind                string          `yaml:"kind"`
// 	TargetCustomization []TargetCustoms `yaml:"targetCustomizations,flow"`
// }

// //PluginGroup represent the structure for the inline plugins
// type PluginGroup struct {
// 	Repo string `yaml:"repo,omitempty"`
// 	Name string `yaml:"name,omitempty"`
// }

// //TargetCustoms represent the single customization group
// type TargetCustoms struct {
// 	Name              string        `yaml:"name"`
// 	Enabled           bool          `yaml:"enabled"`
// 	Type              string        `yaml:"type"`
// 	Config            string        `yaml:"config"`
// 	ClusterName       string        `yaml:"clusterName"`
// 	ClusterDeployment string        `yaml:"clusterDeployment"`
// 	ClusterStart      string        `yaml:"clusterStart,omitempty"`
// 	Spec              Spec          `yaml:"spec,omitempty"`
// 	Plugins           []PluginGroup `yaml:"plugins,flow"`
// }

// type Spec struct {
// 	Wsl             string `yaml:"wsl,omitempty"`
// 	Mac             string `yaml:"mac,omitempty"`
// 	Linux           string `yaml:"linux,omitempty"`
// 	Windows         string `yaml:"windows,omitempty"`
// 	cloudType       string `yaml:"cloudType,omitempty"`
// 	cloudNodes      string `yaml:"cloudNodes,omitempty"`
// 	cloudSecretPath string `yaml:"cloudSecretPath,omitempty"`
// }

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

	// dockerGroupId, err := user.LookupGroup("docker")

	// if _, ok := err.(user.UnknownGroupError); ok {
	// 	message := fmt.Errorf("could not find the group 'docker' on your system. This is required to run.")
	// 	log.Fatal(message)
	// } else if err != nil {
	// 	message := fmt.Errorf("unknown error while trying to retrieve list of groups on your system. Sorry that's all we know: %v", err)
	// 	log.Fatal(message)
	// }

	// a, _ := user.Current()
	// allGroups, err := a.GroupIds()
	// if err != nil {
	// 	message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
	// 	log.Fatal(message)
	// }

	// if !utils.ContainsString(allGroups, dockerGroupId.Gid) {
	// 	message := fmt.Errorf("could not retrieve a list of groups for the current user: %v", err)
	// 	log.Fatal(message)
	// }
	// fmt.Printf("Runtime: %v - %v", runtime.GOOS, runtime.GOARCH)

	// u, _ := user.Current()
	// kDir := path.Join(u.HomeDir, ".kube")
	// if _, err := os.Stat(kDir); os.IsNotExist(err) {
	// 	logrus.Tracef("%v does not exist, creating it now.", kDir)
	// 	os.Mkdir(kDir, 0755)
	// 	uid, _ := strconv.Atoi(u.Uid)
	// 	gid, _ := strconv.Atoi(u.Gid)
	// 	os.Chown(kDir, uid, gid)
	// }

	// cmd := cmd.RootCmd
	// a, _ := exec.LookPath("/usr/local/bin/k3s")
	// fmt.Printf("Cmd: %v", a)

	// b, _ := utils.K3sRunning(cmd)
	// fmt.Printf("Cmd B: %v", b)

	// k8s, err := utils.GetKubernetesClient(2)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// v, err := k8s.GetVersion()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(v)
	// args := []string{mocks.DEPENDENCY_CHECKER_KUBECTL_ON_PATH_PROBE}
	// c := cmd.RootCmd
	// _ = c
	// for _, a := range args {
	// 	fmt.Println(a)
	// }
	// good, err := utils.GetUtils(&cobra.Command{}, []string{}).IsEndpointReachable("https://aksmlproductioncluster-dns-c848e407.hcp.eastus2.azmk8s.io:443/foobaz")
	// fmt.Printf("Endpoint reached: %v\n", good)
	// fmt.Printf("Endpoint error: %v\n", err)

	// bad, err := utils.GetUtils(&cobra.Command{}, []string{}).IsEndpointReachable("aksmlproductioncluster-dns-c848e407.hcp.eastus2.azmk8s.io:443")
	// fmt.Printf("Endpoint reached: %v\n", bad)
	// fmt.Printf("Endpoint error: %v\n", err)

	// bad, err = utils.GetUtils(&cobra.Command{}, []string{}).IsEndpointReachable("aksmlproductioncluster-dns-c848e407.hcp.eastus2.azmk8s.io:443/foobaz")
	// fmt.Printf("Endpoint reached: %v\n", bad)
	// fmt.Printf("Endpoint error: %v\n", err)

	// bad, err = utils.GetUtils(&cobra.Command{}, []string{}).IsEndpointReachable("kubernetes.docker.internal:6443")
	// fmt.Printf("Endpoint reached: %v\n", bad)
	// fmt.Printf("Endpoint error: %v\n", err)
	// a := `
	// import tensorflow

	// a = 3

	// # +`
	// import_regex := regexp.MustCompile(`(?mi)^\s*(?:from|import)\s+(\w+(?:\s*,\s*\w+)*)`)
	// all_imports := import_regex.FindAllStringSubmatch(a, -2)
	// fmt.Printf("Match: %v", all_imports)

	var (
		ZERO_NAMED_STEPS = `
	# ---

	foo = "bar"

	# +
	import tensorflow
	`
		ZERO_NAMED_STEPS_WITH_PARAMS = `
	# ---
	
	# + tags=["parameters"]
	foo = "bar"
	
	# +
	import tensorflow
	`

		ONE_STEP = `
	# ---
	
	# + tags=["parameters"]
	foo = "bar"
	
	# +
	# + tags=["same-step-1"]
	import tensorflow
	`

		ONE_STEP_WITH_CACHE = `
	# ---
	
	# + tags=["parameters"]
	foo = "bar"
	
	# +
	# + tags=["same-step-1", "cache=20d"]
	import tensorflow
	`
	)

	process_steps(ZERO_NAMED_STEPS, "ZERO_NAMED_STEPS")
	process_steps(ZERO_NAMED_STEPS_WITH_PARAMS, "ZERO_NAMED_STEPS_WITH_PARAMS")
	process_steps(ONE_STEP, "ONE_STEP")
	process_steps(ONE_STEP_WITH_CACHE, "ONE_STEP_WITH_CACHE")
}

func process_steps(s string, name string) {
	re := regexp.MustCompile(`(?m)^\s*# (?:\+|\-) ?(.*?)$`)
	stepsFound := re.FindAllStringSubmatch(s, -1)
	fmt.Printf("Steps for %v: %v\n", name, len(stepsFound))

	for i, j := range stepsFound {
		re_tags_text := `tags=\[([^\]]*)\]`
		re_tags := regexp.MustCompile(re_tags_text)
		tags_found := re_tags.FindAllStringSubmatch(j[1], -1)
		fmt.Printf(" - Tags for %v[%v]: %v\n", name, i, len(tags_found))
		if len(tags_found) > 0 {
			all_tags := strings.Split(tags_found[0][1], ",")
			for _, this_tag := range all_tags {
				this_tag = strings.TrimSpace(this_tag)
				if this_tag[0] == '"' {
					this_tag = this_tag[1:]
				}
				if end := len(this_tag) - 1; this_tag[end] == '"' {
					this_tag = this_tag[:end]
				}
				if strings.HasPrefix(this_tag, "cache=") {
					fmt.Printf("   - Cache: %v\n", strings.Split(this_tag, "=")[1])
				} else if strings.HasPrefix(this_tag, "same-step-") {
					fmt.Printf("   - Step: %v\n", strings.Split(this_tag, "-")[2])
				} else {
					fmt.Printf("   - Generic tag: %v\n", this_tag)
				}
			}
		}
	}
}
