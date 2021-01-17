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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SameSpec `json:"spec,omitempty"`
}

// SameSpec is the spec of a SAME project
type SameSpec struct {
	APIVersion string    `json:"apiVersion,omitempty"`
	Version    string    `json:"version,omitempty"`
	Bases      []string  `json:"bases,omitempty"`
	Metadata   Metadata  `json:"metadata,omitempty"`
	EnvFiles   EnvFile   `json:"envs,omitempty"`
	Resources  Resource  `json:"resources,omitempty"`
	Kubeflow   Kubeflow  `json:"kubeflow,omitempty"`
	Pipeline   Pipeline  `json:"pipeline,omitempty"`
	DataSets   []DataSet `json:"data_sets,omitempty"`
	Run        Run       `json:"run,omitempty"`
}

// Metadata is summary data about the SAME program.
type Metadata struct {
	Name    string            `json:"name,omitempty"`
	SHA     string            `json:"sha,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Version string            `json:"version,omitempty"`
}

// EnvFile lists all files that have environment variables that should be mounted in the running pods.
// TODO: Mount all environment variables in the pods.
type EnvFile struct {
	EnvFiles []string `json:"envs,omitempty"`
}

// Resource (may be poorly named) describes the agent pool to be provisioned in the Kubernetes cluster
type Resource struct {
	NodePoolName      string `json:"node_pool_name,omitempty"`
	CreateNewNodePool bool   `json:"create_new_node_pool,omitempty"`
	Cores             Cores  `json:"cores,omitempty"`
	GPU               GPUs   `json:"gpu,omitempty"`
	Disks             []Disk `json:"disks,omitempty"`
}

// Cores lists the requested, required cores for the entire cluster, and the minimum amount per machine.
// TODO: Do we need a machine structure? Too specific?
type Cores struct {
	Requested         int `json:"requested,omitempty"`
	Required          int `json:"required,omitempty"`
	MinimumPerMachine int `json:"minimum_per_machine,omitempty"`
}

// GPUs names the specific GPU required for the machine (by string) and the number per machine
type GPUs struct {
	Type       string `json:"type,omitempty"`
	PerMachine int    `json:"per_machine,omitempty"`
}

// Disk is the name, size and volume mount of disks to provision for the cluster. Assumed that all volume mounts will be made available for every pod.
type Disk struct {
	Name        string      `json:"name,omitempty"`
	Size        string      `json:"size,omitempty"`
	VolumeMount VolumeMount `json:"volume_mount,omitempty"`
}

// VolumeMount is the specific volume handle for the mounted disk, and a name.
type VolumeMount struct {
	MountPath string `json:"mount_path,omitempty"`
	Name      string `json:"name,omitempty"`
}

// Kubeflow specifies the version of the Kubeflow cluster and all associated services to provision. It also names the namespace to deploy to.
type Kubeflow struct {
	KubernetesAPIServer string   `json:"kubeflow_api_server,omitempty"`
	KubeflowVersion     string   `json:"kubeflow_version,omitempty"`
	KubeflowNamespace   string   `json:"kubeflow_namespace,omitempty"`
	Services            []string `json:"services,omitempty"`
	CredentialFile      string   `json:"credential_file,omitempty"`
}

// Pipeline names the specific (pre-compiled) pipeline package to upload (or verify already exists) on Kubeflow.
type Pipeline struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Package     string `json:"package,omitempty"`
}

// DataSet is the data to be downloaded/mounted into the cluster.
type DataSet struct {
	Type          string `json:"type,omitempty"`
	URL           string `json:"url,omitempty"`
	MakeLocalCopy bool   `json:"make_local_copy,omitempty"`
}

// Run is the name and specific parameters to run against one of the previously created pipelines.
type Run struct {
	Name       string            `json:"name,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}
