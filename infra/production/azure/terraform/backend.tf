# Remote state in Azure Blob Storage
# ─────────────────────────────────────────────────────────────────────────────
# PREREQUISITE: Create this storage account ONCE manually before running
# terraform init (Terraform can't bootstrap its own backend):
#
#   az group create --name goride-tfstate-rg --location eastasia
#   az storage account create \
#     --name goridetfstate \
#     --resource-group goride-tfstate-rg \
#     --sku Standard_LRS \
#     --encryption-services blob
#   az storage container create \
#     --name tfstate \
#     --account-name goridetfstate
#
# Then run: terraform init
# ─────────────────────────────────────────────────────────────────────────────
terraform {
  backend "azurerm" {
    resource_group_name  = "goride-tfstate-rg"
    storage_account_name = "goridetfstate"
    container_name       = "tfstate"
    key                  = "production.terraform.tfstate"
  }
}
