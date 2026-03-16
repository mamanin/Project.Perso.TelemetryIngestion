// =======================================================================
// Ingestion Core Infrastructure Deployment
// -----------------------------------------------------------------------
// Module: main.bicep
// Description: Deploys the ingestion core infrastructure resources 
//       required for the different layers of the telemetry ingestion
//       service.
// =======================================================================

import { ingestionConstants } from './constants/ingestion.constants.bicep'

// -----------------------------------------------------------------------
// Parameters and Variables
// -----------------------------------------------------------------------

@description('The name of the resource group to deploy to')
var location = resourceGroup().location

@description('The name prefix of the resource to deploy')
var prefix = 'tispoc'

@description('The tags to apply to all resources')
var tags = {
  project: prefix
  'managed-by': 'bicep'
  service: 'ingestion'
}

// -----------------------------------------------------------------------
// Monitoring & Logging
// -----------------------------------------------------------------------

module logAnalytics '../common/modules/loganalytics.module.bicep' = {
  name: 'logAnalyticsDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
  }
}

module appInsights '../common/modules/appinsights.module.bicep' = {
  name: 'appInsightsDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    workspaceId: logAnalytics.outputs.id
  }
}

// -----------------------------------------------------------------------
// Application Support Infrastructure
// -----------------------------------------------------------------------

module containerRegistry '../common/modules/containerregistry.module.bicep' = {
  name: 'containerRegistryDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: 'Basic'
  }
}

module containerAppEnvironment '../common/modules/containerappenv.module.bicep' = {
  name: 'containerAppEnvironmentDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    logAnalyticsWorkspaceName: logAnalytics.outputs.name
    enableOTelEndpoints: true
    appInsightsConnectionString: appInsights.outputs.connectionString
  }
}

module bronzeIdentity '../common/modules/identity.userassigned.module.bicep' = {
  name: 'bronzeIdentityDeploy'
  params: {
    prefix: prefix
    number: '001'
    location: location
    tags: union(tags, {
      application: 'bronze'
    })
  }
}

module silverIdentity '../common/modules/identity.userassigned.module.bicep' = {
  name: 'silverIdentityDeploy'
  params: {
    prefix: prefix
    number: '002'
    location: location
    tags: union(tags, {
      application: 'silver'
    })
  }
}

module goldIdentity '../common/modules/identity.userassigned.module.bicep' = {
  name: 'goldIdentityDeploy'
  params: {
    prefix: prefix
    number: '003'
    location: location
    tags: union(tags, {
      application: 'gold'
    })
  }
}

// -----------------------------------------------------------------------
// Storage Resources
// -----------------------------------------------------------------------

module storageAccount '../common/modules/storageaccount.module.bicep' = {
  name: 'storageAccountDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    skuName: 'Standard_LRS'
    blobPublicAccess: true
  }
}

module storageAccountBlobs '../common/modules/storageaccount.containers.module.bicep' = {
  name: 'storageAccountBlobsDeploy'
  params: {
    storageAccountName: storageAccount.outputs.name
    containers: [
      {
        name: ingestionConstants.tableStorage.checkpointsTableName
        publicAccess: 'None'
      }
    ]
  }
}

module redisCache '../common/modules/redis.module.bicep' = {
  name: 'redisCacheDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: 'ComputeOptimized_X20'
    highAvailability: false
  }
}

module redisDatabase '../common/modules/redis.database.module.bicep' = {
  name: 'redisDatabaseDeploy'
  params: {
    redisName: redisCache.outputs.name
    enableAccessKeyAuth: false
  }
}

module kustoCluster '../common/modules/kusto.cluster.module.bicep' = {
  name: 'kustoClusterDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: {
      name: 'Standard_E8ads_v5'
      tier: 'Standard'
      capacity: 2
    }
    enableAutoStop: true
  }
}

module kustoDatabase '../common/modules/kusto.cluster.database.module.bicep' = {
  name: 'kustoDatabaseDeploy'
  params: {
    kustoClusterName: kustoCluster.outputs.kustoClusterName
    databaseName: ingestionConstants.dataExplorer.telemetryDatabaseName
    scripts: [
      {
        name: 'create-metrics-tables'
        scriptVersion: 'v1.0.1'
        script: loadTextContent('scripts/telemetries.tables.metrics.kql')
      }
      {
        name: 'create-devices-status-view'
        scriptVersion: 'v1.0.1'
        script: loadTextContent('scripts/telemetries.views.devices.state.kql')
      }
      {
        name: 'create-10-minutes-view'
        scriptVersion: 'v1.0.1'
        script: loadTextContent('scripts/telemetries.views.metrics.10m.kql')
      }
      {
        name: 'create-1-hour-view'
        scriptVersion: 'v1.0.1'
        script: loadTextContent('scripts/telemetries.views.metrics.1h.kql')
      }
      {
        name: 'create-1-day-view'
        scriptVersion: 'v1.0.1'
        script: loadTextContent('scripts/telemetries.views.metrics.1d.kql')
      }
    ]
  }
}

// -----------------------------------------------------------------------
// Real-time Communication
// -----------------------------------------------------------------------

module eventHubNamespace '../common/modules/eventhub.namespace.module.bicep' = {
  name: 'eventHubNamespaceDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: 'Premium'
    capacity: 2
  }
}

module rawEventHub '../common/modules/eventhub.module.bicep' = {
  name: 'rawEventHubDeploy'
  params: {
    name: ingestionConstants.eventhub.rawName
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: ingestionConstants.bronze.partitionCount
    consumerGroups: [
      ingestionConstants.consumerGroups.bronze
    ]
  }
}

module metricsEventHub '../common/modules/eventhub.module.bicep' = {
  name: 'metricsEventHubDeploy'
  params: {
    name: ingestionConstants.eventhub.metricsName
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: ingestionConstants.silver.partitionCount
    consumerGroups: [
      ingestionConstants.consumerGroups.silver
    ]
  }
}

module dataEventHub '../common/modules/eventhub.module.bicep' = {
  name: 'dataEventHubDeploy'
  params: {
    name: ingestionConstants.eventhub.dataName
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: ingestionConstants.gold.partitionCount
    consumerGroups: [
      ingestionConstants.consumerGroups.gold
    ]
  }
}
