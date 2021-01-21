provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

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

resource "azurerm_kubernetes_cluster_node_pool" "samenodepool" {
  name                  = "samenodepool"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.same_cluster.id
  vm_size               = "Standard_DS2_v2"
  node_count            = 5

  tags = {
    Environment = "Production"
  }
}

resource "azurerm_storage_account" "sameac" {
  name                     = "${var.prefix}ac"
  resource_group_name      = azurerm_resource_group.same_rg.name
  location                 = azurerm_resource_group.same_rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "samecontainer" {
  name                  = "${var.prefix}-container"
  storage_account_name  = azurerm_storage_account.sameac.name
  container_access_type = "private"
}

resource "azurerm_key_vault" "same_kv" {
  name                        = "${var.prefix}-kv"
  location                    = azurerm_resource_group.same_rg.location
  resource_group_name         = azurerm_resource_group.same_rg.name
  enabled_for_disk_encryption = true
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  sku_name = "standard"
}

resource "azurerm_key_vault_access_policy" "a_policy" {
  key_vault_id = azurerm_key_vault.same_kv.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  secret_permissions = [
    "get", "set", "list", "delete"
  ]
}

resource "azurerm_key_vault_access_policy" "c_policy" {
  key_vault_id = azurerm_key_vault.same_kv.id
  tenant_id    = azurerm_kubernetes_cluster.same_cluster.identity[0].tenant_id
  object_id    = azurerm_kubernetes_cluster.same_cluster.identity[0].principal_id

  secret_permissions = [
    "get"
  ]
}

resource "azurerm_key_vault_secret" "container_secret" {
  name         = "same-container-access"
  value        = azurerm_storage_account.sameac.primary_access_key
  key_vault_id = azurerm_key_vault.same_kv.id
}
