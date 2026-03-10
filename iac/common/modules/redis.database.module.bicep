// =======================================================================
// Redis Database Module
// -----------------------------------------------------------------------
// Module: redis.database.module.bicep
// Description: Deploys Azure Redis Database
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.cache/redisenterprise/databases
// =======================================================================

@description('The name of the redis cache resource')
param redisName string

@description('Enable access key authentication for the Redis Database')
param enableAccessKeyAuth bool = false

resource redis 'Microsoft.Cache/redisEnterprise@2025-08-01-preview' existing = {
  name: redisName
}

resource database 'Microsoft.Cache/redisEnterprise/databases@2025-08-01-preview' = {
  parent: redis
  name: 'default'
  properties: {
    accessKeysAuthentication: enableAccessKeyAuth ? 'Enabled' : 'Disabled'
    clientProtocol: 'Encrypted'
  }
}

@description('The Redis Database resource ID')
output id string = database.id

@description('The Redis Database resource name')
output name string = database.name

@description('The Redis Database resource port')
output port int = database.properties.port
