// =======================================================================
// Redis Cache Access Policy Module
// -----------------------------------------------------------------------
// Module: redis.accesspolicy.module.bicep
// Description: Deploys Azure Redis Cache Access Policies
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.cache/redis/accesspolicies
// =======================================================================

@description('The name of the redis cache resource')
param redisName string

@description('The name of the access policy')
@minLength(4)
param name string

@description('The permissions assigned to the access policy')
param permissions string

resource redis 'Microsoft.Cache/redis@2024-11-01' existing = {
  name: redisName
}

resource redisAccessPolicy 'Microsoft.Cache/redis/accessPolicies@2024-11-01' = {
  parent: redis
  name: name
  properties: {
    permissions: permissions
  }
}
