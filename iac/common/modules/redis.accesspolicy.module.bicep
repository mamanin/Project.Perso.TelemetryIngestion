// =======================================================================
// Redis Cache Access Policy Module
// -----------------------------------------------------------------------
// Module: redis.accesspolicy.module.bicep
// Description: Deploys Azure Redis Cache Access Policies
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.cache/redisenterprise/databases/accesspolicyassignments
// =======================================================================

@description('The name of the redis cache resource')
param redisName string

@description('The principal ID to assign the role to')
param principalId string

resource redis 'Microsoft.Cache/redisEnterprise@2025-08-01-preview' existing = {
  name: redisName
}

resource database 'Microsoft.Cache/redisEnterprise/databases@2025-08-01-preview' existing = {
  parent: redis
  name: 'default'
}

resource roleAssignment 'Microsoft.Cache/redisEnterprise/databases/accessPolicyAssignments@2025-08-01-preview' = {
  parent: database
  name: guid(database.id, principalId, 'default')
  properties: {
    accessPolicyName: 'default' // Currently the only supported access policy is 'default' allowing all permissions.
    user: {
      objectId: principalId
    }
  }
}

