// =======================================================================
// Redis Cache Module
// -----------------------------------------------------------------------
// Module: redis.module.bicep
// Description: Deploys Azure Redis Cache
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.cache/redisenterprise
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Redis Cache')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The SKU of the Redis Cache')
param sku 'Balanced_B0' | 'Balanced_B1' | 'Balanced_B10' | 'ComputeOptimized_X10' | 'ComputeOptimized_X20' = 'Balanced_B0'

@description('Enable high availability for the Redis Cache')
param highAvailability bool = true

resource redis 'Microsoft.Cache/redisEnterprise@2025-08-01-preview' = {
  name: BuildResourceName(prefix, 'red', number)
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  tags: tags
  sku: {
    name: sku
  }
  properties: {
    highAvailability: highAvailability ? 'Enabled' : 'Disabled'
    publicNetworkAccess: 'Enabled'
    minimumTlsVersion: '1.2'
  }
}

@description('The Redis Cache resource ID')
output id string = redis.id

@description('The Redis Cache resource name')
output name string = redis.name

@description('The Redis Cache host name')
output hostName string = redis.properties.hostName
