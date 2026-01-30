// =======================================================================
// Kusto Cluster Database Module
// -----------------------------------------------------------------------
// Module: kusto.cluster.database.module.bicep
// Description: Deploys an Azure Kusto Cluster Database
// See: 
//    - https://learn.microsoft.com/en-us/azure/templates/microsoft.kusto/clusters/databases
//    - https://learn.microsoft.com/en-us/azure/templates/microsoft.kusto/clusters/databases/scripts
// =======================================================================

import { KustoClusterDatabasesScriptDescription } from '../types/kusto.cluster.databases.script.type.bicep'

@description('The Kusto Cluster resource name')
param kustoClusterName string

@description('The name of the Kusto Database to create')
@minLength(4)
param databaseName string

@description('The Azure region for the Kusto Cluster database')
param location string = resourceGroup().location

@description('The scripts to run on Kusto Cluster Database creation.')
param scripts KustoClusterDatabasesScriptDescription[] = []

resource kustoCluster 'Microsoft.Kusto/clusters@2024-04-13' existing = {
  name: kustoClusterName
}

resource database 'Microsoft.Kusto/clusters/databases@2024-04-13' = {
  parent: kustoCluster
  name: databaseName
  kind: 'ReadWrite'
  location: location
  // Retention and caching should be defined via scripts
}

@batchSize(1)
resource script 'Microsoft.Kusto/clusters/databases/scripts@2024-04-13' = [for scriptItem in scripts: {
  parent: database
  name: scriptItem.name
  properties: {
    forceUpdateTag: scriptItem.scriptVersion
    scriptContent: scriptItem.script
    continueOnErrors: false
  }
}]

@description('The Kusto Cluster Database resource name')
output kustoClusterDatabaseName string = database.name

@description('The Kusto Cluster Database resource ID')
output kustoClusterDatabaseId string = database.id
