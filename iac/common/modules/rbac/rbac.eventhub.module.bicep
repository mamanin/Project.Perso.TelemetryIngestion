// =======================================================================
// Event Hub Role Assignment Module
// -----------------------------------------------------------------------
// Module: rbac.eventhub.module.bicep
// Description: Creates role assignments for Event Hub resources
// =======================================================================

import { RbacRoleType } from '../../types/rbac.role.type.bicep'

@description('The name of the Event Hub Namespace')
param namespaceName string

@description('The name of the Event Hub')
param name string

@description('The principal ID to assign the role to')
param principalId string

@description('The roles to assign to the principal')
param roles RbacRoleType[]

resource namespace 'Microsoft.EventHub/namespaces@2025-05-01-preview' existing = {
  name: namespaceName
}

resource eventHub 'Microsoft.EventHub/namespaces/eventhubs@2025-05-01-preview' existing = {
  parent: namespace
  name: name
}

resource roleAssignments 'Microsoft.Authorization/roleAssignments@2022-04-01' = [for role in roles: {
  scope: eventHub
  name: guid(eventHub.id, principalId, role.id)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', role.id)
    principalId: principalId
    description: role.description
  }
}]
