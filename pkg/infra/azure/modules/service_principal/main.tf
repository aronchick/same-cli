resource "azurerm_resource_group" "daron-kf-cluster-rg" {
  name     = "daron-kf-cluster-rg"
  location = "westeurope"
}

resource "azuread_service_principal" "kf_sp" {
  app_role_assignment_required = false
  name                         = "kubeconfig-read-mM1u0"
  role                         = "Azure Kubernetes Service Cluster User Role"
  scopes                       = ["/subscriptions/2865c7d1-29fa-485a-8862-717377bdbf1b/resourcegroups/daron-kf-cluster-rg"]
  tags                         = ["same", "service", "principal"]
  depends_on                   = [resource.same_cluster]
}