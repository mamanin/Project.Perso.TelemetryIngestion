// =======================================================================
// Redis Cache Role Assignment Module
// -----------------------------------------------------------------------
// Module: rbac.redis.module.bicep
// Description: Deploys Azure Redis Cache Role Assignments
// =======================================================================

@description('The name of the redis cache resource')
param name string

@description('The principal ID to assign the role to')
param principalId string

@description('The name of the access policy resource')
@minLength(1)
param policyNames string[]

resource redis 'Microsoft.Cache/redis@2024-11-01' existing = {
  name: name
}

resource roleAssignments 'Microsoft.Cache/redis/accessPolicyAssignments@2024-11-01' =  [for policyName in policyNames: {
  parent: redis
  name: guid(redis.id, principalId, policyName)
  properties: {
    accessPolicyName: policyName
    objectId: principalId
    objectIdAlias: principalId
  }
}]
