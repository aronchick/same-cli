package loaders

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Constants for the app
const (
	DefaultCacheDir = ".cache"
)

// SameConfig is metadata information about the file
type SameConfig struct {
	metav1.TypeMeta   `yaml:",inline"`
	metav1.ObjectMeta `yaml:"metadata,omitempty"`

	Spec SameSpec `yaml:"spec,omitempty"`
}

// SameSpec is the spec of a SAME project
type SameSpec struct {
	APIVersion            string                 `yaml:"apiVersion,omitempty"`
	Version               string                 `yaml:"version,omitempty"`
	Bases                 []string               `yaml:"bases,omitempty"`
	Metadata              Metadata               `yaml:"metadata,omitempty"`
	EnvFiles              []string               `yaml:"envfiles,omitempty"`
	Resources             Resource               `yaml:"resources,omitempty"`
	Workflow              Workflow               `yaml:"workflow,omitempty"`
	Pipeline              Pipeline               `yaml:"pipeline,omitempty"`
	Environments          map[string]Environment `yaml:"environments,omitempty"`
	DataSets              []DataSet              `yaml:"dataSets,omitempty"`
	Run                   Run                    `yaml:"run,omitempty"`
	DebuggingFeatureFlags map[string]bool        `yaml:"debugging_features_flags,omitempty"`
	ConfigFilePath        string                 `yaml:"configfilepath,omitempty"`
	KubeConfig            string                 `yaml:"kubeconfig,omitempty"`
}

// Metadata is summary data about the SAME program.
type Metadata struct {
	Name    string            `yaml:"name,omitempty"`
	SHA     string            `yaml:"sha,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty"`
	Version string            `yaml:"version,omitempty"`
}

// // EnvFile lists all files that have environment variables that should be mounted in the running pods.
// // TODO: Mount all environment variables in the pods.
// type EnvFile struct {
// 	File string `yaml:"envfiles,omitempty"`
// }

// Resource (may be poorly named) describes the agent pool to be provisioned in the Kubernetes cluster
type Resource struct {
	Provider          string `yaml:"provider,omitempty"`
	ClusterProfile    string `yaml:"cluster_profile,omitempty"`
	NodePoolName      string `yaml:"nodePoodName,omitempty"`
	CreateNewNodePool bool   `yaml:"createNewNodePool,omitempty"`
	Cores             Cores  `yaml:"cores,omitempty"`
	GPU               GPU    `yaml:"gpus,omitempty"`
	Disks             []Disk `yaml:"disks,omitempty"`
}

// Cores lists the requested, required cores for the entire cluster, and the minimum amount per machine.
// TODO: Do we need a machine structure? Too specific?
type Cores struct {
	Requested         int `yaml:"requested,omitempty"`
	Required          int `yaml:"required,omitempty"`
	MinimumPerMachine int `yaml:"minimum_per_machine,omitempty"`
}

// GPU names the specific GPU required for the machine (by string) and the number per machine
type GPU struct {
	Type       string `yaml:"type,omitempty"`
	PerMachine int    `yaml:"per_machine,omitempty"`
}

// Disk is the name, size and volume mount of disks to provision for the cluster. Assumed that all volume mounts will be made available for every pod.
type Disk struct {
	Name        string      `yaml:"name,omitempty"`
	Size        string      `yaml:"size,omitempty"`
	VolumeMount VolumeMount `yaml:"volumeMount,omitempty"`
}

// VolumeMount is the specific volume handle for the mounted disk, and a name.
type VolumeMount struct {
	MountPath string `yaml:"mountPath,omitempty"`
	Name      string `yaml:"name,omitempty"`
}

// Workflow is the workflow executor for SAME
// TODO: Obviously parameters can't just be a 'Kubeflow' but it's good enough for now
type Workflow struct {
	Type       string   `yaml:"mountPath,omitempty"`
	Parameters Kubeflow `yaml:"parameters,omitempty"`
}

// Kubeflow specifies the version of the Kubeflow cluster and all associated services to provision. It also names the namespace to deploy to.
type Kubeflow struct {
	KubernetesAPIServerURI string   `yaml:"kubernetesAPIServerURI,omitempty"`
	KubeflowVersion        string   `yaml:"kubeflowVersion,omitempty"`
	KubeflowNamespace      string   `yaml:"kubeflowNamespace,omitempty"`
	Services               []string `yaml:"services,omitempty"`
	CredentialFile         string   `yaml:"credentialFile,omitempty"`
}

// Pipeline names the specific (pre-compiled) pipeline package to upload (or verify already exists) on Kubeflow.
type Pipeline struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Package     string `yaml:"package,omitempty"`
}

// DataSet is the data to be downloaded/mounted into the cluster.
type DataSet struct {
	Type          string `yaml:"type,omitempty"`
	URL           string `yaml:"url,omitempty"`
	MakeLocalCopy bool   `yaml:"makeLocalCopy,omitempty"`
}

// Run is the name and specific parameters to run against one of the previously created pipelines.
// RunWrapper comes from here: https://medium.com/@nate510/dynamic-json-umarshalling-in-go-88095561d6a0
type Run struct {
	Name       string                 `yaml:"name,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

type Environment struct {
	ImageTag                 string                `yaml:"image_tag,omitempty"`
	AppendCurrentEnvironment bool                  `yaml:"append_current_environment,omitempty"`
	Packages                 []string              `yaml:"packages,omitempty,omitempty"`
	PrivateRegistry          bool                  `yaml:"private_registry,omitempty"`
	Credentials              RepositoryCredentials `yaml:"repository_credentials,omitempty"`
}

type RepositoryCredentials struct {
	SecretName string `yaml:"secretname,omitempty"`
	Server     string `yaml:"server,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	Email      string `yaml:"email,omitempty"`
}

// type Parameter struct {
// 	Parameter ParameterWrapper `yaml:"-,omitempty"`
// }
// type ParameterWrapper struct {
// 	Parameter
// 	Partial bool `json:”-”`
// }

// func (w *ParameterWrapper) UnmarshalJSON(data []byte) error {
// 	if id, err := strconv.Atoi(string(data)); err == nil {
// 		w.ID = id
// 		w.Partial = true
// 		return nil
// 	}
// 	return json.Unmarshal(data, &w.Sprocket)
// }
