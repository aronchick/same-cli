provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "same_rg" {
  name     = "${var.prefix}-k8s-resources"
  location = var.location
}

resource "azurerm_kubernetes_cluster" "same_cluster" {
  name                = "${var.prefix}-k8s"
  location            = azurerm_resource_group.same_rg.location
  resource_group_name = azurerm_resource_group.same_rg.name
  dns_prefix          = "${var.prefix}-k8s"

  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_DS2_v2"
  }

  identity {
    type = "SystemAssigned"
  }

  addon_profile {
    aci_connector_linux {
      enabled = false
    }

    azure_policy {
      enabled = false
    }

    http_application_routing {
      enabled = false
    }

    kube_dashboard {
      enabled = true
    }

    oms_agent {
      enabled = false
    }
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "same_node_pool" {
  name                  = "same_node_pool"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.same_cluster.id
  vm_size               = "Standard_DS2_v2"
  node_count            = 5

  tags = {
    Environment = "Production"
  }
}