// =======================================================================
// Container Registry Module
// -----------------------------------------------------------------------
// Module: containerregistry.module.bicep
// Description: Deploys an Azure Container Registry (ACR)
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.containerregistry/registries
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The Azure region for the Container Registry')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('SKU for Container Registry')
@allowed([
  'Basic'
  'Standard'
  'Premium'
])
param sku string = 'Basic'

resource containerRegistry 'Microsoft.ContainerRegistry/registries@2025-11-01' = {
  name: BuildResourceName(prefix, 'acr', number)
  location: location
  sku: {
    name: sku
  }
  tags: tags
}

@description('The resource ID of the Container Registry')
output id string = containerRegistry.id

@description('The name of the Container Registry')
output name string = containerRegistry.name

@description('The login server for the Container Registry')
output loginServer string = containerRegistry.properties.loginServer
