// =======================================================================
// Kusto Cluster Database Role Assignment Module
// -----------------------------------------------------------------------
// Module: rbac.kusto.database.module.bicep
// Description: Creates role assignments for Kusto Cluster Database
//       resources.
// =======================================================================

@description('The kusto Cluster resource name')
param kustoClusterName string

@description('The name of the Kusto Database to create')
@minLength(4)
param databaseName string

@description('The principal ID to assign the role to')
param principalId string

@description('The principal type to assign the role to')
@allowed([
  'App'
  'Group'
  'User'
])
param principalType string = 'App'

@description('The roles to assign to the principal')
@allowed([
  'Admin'
  'Ingestor'
  'Monitor'
  'UnrestrictedViewer'
  'User'
  'Viewer'
])
param roles string[]

resource kustoCluster 'Microsoft.Kusto/clusters@2024-04-13' existing = {
  name: kustoClusterName
}

resource database 'Microsoft.Kusto/clusters/databases@2024-04-13' existing = {
  parent: kustoCluster
  name: databaseName
}

resource roleAssignments 'Microsoft.Kusto/clusters/databases/principalAssignments@2024-04-13' = [for role in roles: {
  parent: database
  name: guid(database.id, principalId, role)
  properties: {
    principalId: principalId
    principalType: principalType
    role: role
  }
}]
