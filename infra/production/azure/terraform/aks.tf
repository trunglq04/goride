resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = var.cluster_name
  kubernetes_version  = var.kubernetes_version

  # System-assigned managed identity — Azure manages credentials automatically
  identity {
    type = "SystemAssigned"
  }

  default_node_pool {
    name           = "system"
    node_count     = var.aks_node_count
    vm_size        = var.aks_node_size
    vnet_subnet_id = azurerm_subnet.aks.id

    # Enable auto-scaling (optional, adjust min/max as needed)
    # enable_auto_scaling = true
    # min_count           = 2
    # max_count           = 5
  }

  network_profile {
    network_plugin    = "azure"   # Azure CNI — pods get real VNet IPs
    load_balancer_sku = "standard"
  }

  tags = {
    project     = "goride"
    environment = "production"
  }
}
