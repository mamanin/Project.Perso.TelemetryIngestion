// =======================================================================
// Kusto Cluster Module
// -----------------------------------------------------------------------
// Module: kusto.cluster.module.bicep
// Description: Deploys an Azure Kusto Cluster
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.kusto/clusters
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'
import { KustoClusterSkuDescription } from '../types/kusto.cluster.sku.type.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The Azure region for the Kusto Cluster')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The SKU of the Kusto Cluster')
param sku KustoClusterSkuDescription

@description('Enable auto-stop for the Kusto Cluster to save costs when idle')
param enableAutoStop bool = false

@description('Enable streaming ingest for the Kusto Cluster')
param enableStreamingIngest bool = false

resource cluster 'Microsoft.Kusto/clusters@2024-04-13' = {
  name: BuildResourceName(prefix, 'adx', number)
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  tags: tags
  sku: {
    name: sku.name
    tier: sku.tier
    capacity: sku.capacity
  }
  properties: {
    enableAutoStop: enableAutoStop
    enableStreamingIngest: enableStreamingIngest
    enableDiskEncryption: true
  }
}

@description('The Kusto Cluster resource name')
output kustoClusterName string = cluster.name

@description('The Kusto Cluster resource ID')
output kustoClusterId string = cluster.id

@description('The Kusto Cluster URI')
output kustoClusterUri string = cluster.properties.uri
