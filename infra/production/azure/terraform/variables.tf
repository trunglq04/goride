variable "location" {
  description = "Azure region to deploy all resources"
  type        = string
  default     = "malaysiawest"
}

variable "resource_group_name" {
  description = "Name of the Azure Resource Group"
  type        = string
  default     = "goride-rg"
}

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
  default     = "goride-aks"
}

variable "aks_node_count" {
  description = "Number of nodes in the default AKS node pool"
  type        = number
  default     = 2
}

variable "aks_node_size" {
  description = "VM SKU for AKS nodes (e.g. Standard_D2s_v3, Standard_B2s)"
  type        = string
  default     = "Standard_D2s_v3"
}

variable "acr_name" {
  description = "Azure Container Registry name (must be globally unique, alphanumeric only)"
  type        = string
  default     = "goridecr"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the AKS cluster (null = use latest stable)"
  type        = string
  default     = null
}

variable "vnet_address_space" {
  description = "CIDR block for the Virtual Network"
  type        = string
  default     = "10.0.0.0/16"
}

variable "subnet_address_prefix" {
  description = "CIDR block for the AKS subnet"
  type        = string
  default     = "10.0.1.0/24"
}
