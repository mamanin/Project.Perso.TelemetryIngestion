// =======================================================================
// Container Registry Role Assignment Module
// -----------------------------------------------------------------------
// Module: rbac.containerregistry.module.bicep
// Description: Creates role assignments for Container Registry resources
// =======================================================================

import { RbacRoleType } from '../../types/rbac.role.type.bicep'

@description('The name of the Container Registry instance')
param name string

@description('The principal ID to assign the role to')
param principalId string

@description('The roles to assign to the principal')
param roles RbacRoleType[]

resource containerRegistry 'Microsoft.ContainerRegistry/registries@2025-05-01-preview' existing = {
  name: name
}

resource roleAssignments 'Microsoft.Authorization/roleAssignments@2022-04-01' = [for role in roles: {
  scope: containerRegistry
  name: guid(containerRegistry.id, principalId, role.id)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', role.id)
    principalId: principalId
    description: role.description
  }
}]
