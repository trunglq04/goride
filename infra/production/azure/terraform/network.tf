resource "azurerm_virtual_network" "main" {
  name                = "goride-vnet"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  address_space       = [var.vnet_address_space]

  tags = {
    project     = "goride"
    environment = "production"
  }
}

resource "azurerm_subnet" "aks" {
  name                 = "goride-aks-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.subnet_address_prefix]
}
