// =======================================================================
// Redis Cache Module
// -----------------------------------------------------------------------
// Module: redis.module.bicep
// Description: Deploys Azure Redis Cache
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.cache/redis
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'
import { RedisSkuDescription } from '../types/redis.sku.type.bicep'

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
param sku RedisSkuDescription

@description('Disable access key authentication for the Redis Cache')
param disableAccessKeyAuth bool = false

@description('Enable Azure AD authentication for the Redis Cache')
param enableAadAuth bool = true

resource redis 'Microsoft.Cache/redis@2024-11-01' = {
  name: BuildResourceName(prefix, 'red', number)
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  tags: tags
  properties: {
    sku: {
      name: sku.name
      family: sku.name == 'Premium' ? 'P' : 'C'
      capacity: sku.capacity
    }
    disableAccessKeyAuthentication: disableAccessKeyAuth
    redisConfiguration: {
        'aad-enabled' : enableAadAuth ? 'true' : 'false'
    }
    minimumTlsVersion: '1.2'
  }
}

@description('The Redis Cache resource ID')
output id string = redis.id

@description('The Redis Cache resource name')
output name string = redis.name

@description('The Redis Cache host name')
output hostName string = redis.properties.hostName

@description('The Redis Cache port')
output port int = redis.properties.port
