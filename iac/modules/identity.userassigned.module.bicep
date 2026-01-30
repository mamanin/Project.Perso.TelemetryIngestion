// =======================================================================
// User Assigned Identity Module
// -----------------------------------------------------------------------
// Module: identity.userassigned.module.bicep
// Description: Deploys an Azure User Assigned Managed Identity
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.managedidentity/userassignedidentities
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the User Assigned Identity')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

resource apiIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2025-01-31-preview' = {
  name: BuildResourceName(prefix, 'uai', number)
  location: location
  tags: tags
}

@description('The resource ID of the User Assigned Identity')
output id string = apiIdentity.id

@description('The principal ID of the User Assigned Identity')
output principalId string = apiIdentity.properties.principalId

@description('The client ID of the User Assigned Identity used for authentication to Azure services')
output clientId string = apiIdentity.properties.clientId
