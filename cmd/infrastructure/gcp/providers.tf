provider "google" {
  project = var.project
}

provider "random" {
}

provider "null" {
  version = "~> 2.1"
}

provider "kubernetes" {
  # don't load config from ~/.kube/config
  load_config_file = false

  # instead use the cluster managed by this module
  host                   = "https://${local.cluster_endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(local.cluster_ca_certificate)
}

locals {
  cluster_endpoint       = google_container_cluster.kubeflow_cluster.endpoint
  cluster_ca_certificate = google_container_cluster.kubeflow_cluster.master_auth.0.cluster_ca_certificate
}
