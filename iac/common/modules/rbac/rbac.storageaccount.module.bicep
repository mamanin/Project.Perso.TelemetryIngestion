// =======================================================================
// Storage Role Assignment Module
// -----------------------------------------------------------------------
// Module: rbac.storage.module.bicep
// Description: Creates role assignments for Storage Account resources
// =======================================================================

import { RbacRoleType } from '../../types/rbac.role.type.bicep'

@description('The name of the Storage Account')
param name string

@description('The principal ID to assign the role to')
param principalId string

@description('The roles to assign to the principal')
param roles RbacRoleType[]

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-01-01' existing = {
  name: name
}

resource roleAssignments 'Microsoft.Authorization/roleAssignments@2022-04-01' = [for role in roles: {
  scope: storageAccount
  name: guid(storageAccount.id, principalId, role.id)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', role.id)
    principalId: principalId
    description: role.description
  }
}]
