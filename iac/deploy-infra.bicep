// =======================================================================
// Core Infrastructure Deployment
// -----------------------------------------------------------------------
// Module: deploy-infra.bicep
// Description: Deploys the core infrastructure resources required for the
//       different layers of the telemetry ingestion service. 
// =======================================================================

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
}

// -----------------------------------------------------------------------
// Monitoring & Logging
// -----------------------------------------------------------------------

module logAnalytics './modules/loganalytics.module.bicep' = {
  name: 'logAnalyticsDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    retentionInDays: 10
  }
}

module appInsights './modules/appinsights.module.bicep' = {
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

module containerRegistry './modules/containerregistry.module.bicep' = {
  name: 'containerRegistryDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: 'Basic'
  }
}

module containerAppEnvironment './modules/containerappenv.module.bicep' = {
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

// -----------------------------------------------------------------------
// Storage Resources
// -----------------------------------------------------------------------

module storageAccount './modules/storageaccount.module.bicep' = {
  name: 'storageAccountDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    skuName: 'Standard_LRS'
    blobPublicAccess: true
  }
}

module storageAccountBlobs './modules/storageaccount.containers.module.bicep' = {
  name: 'storageAccountBlobsDeploy'
  params: {
    storageAccountName: storageAccount.outputs.name
    containers: [
      {
        name: 'partition-checkpoints'
        publicAccess: 'None'
      }
    ]
  }
}

module redisCache './modules/redis.module.bicep' = {
  name: 'redisCacheDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: {
      name: 'Basic'
      capacity: 0
    }
    disableAccessKeyAuth: true
    enableAadAuth: true
  }
}

module kustoCluster './modules/kusto.cluster.module.bicep' = {
  name: 'kustoClusterDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: {
      name: 'Dev(No SLA)_Standard_E2a_v4'
      tier: 'Basic'
      capacity: 1
    }
    enableAutoStop: true
  }
}

module kustoDatabase './modules/kusto.cluster.database.module.bicep' = {
  name: 'kustoDatabaseDeploy'
  params: {
    kustoClusterName: kustoCluster.outputs.kustoClusterName
    databaseName: 'telemetries'
    scripts: [
      {
        name: 'create-metrics-tables'
        scriptVersion: 'v1.0.0'
        script: loadTextContent('scripts/kustodb.telemetries.tables.metrics.kql')
      }
      {
        name: 'create-10-minutes-views'
        scriptVersion: 'v1.0.0'
        script: loadTextContent('scripts/kustodb.telemetries.views.10m.kql')
      }
      {
        name: 'create-1-hour-views'
        scriptVersion: 'v1.0.0'
        script: loadTextContent('scripts/kustodb.telemetries.views.1h.kql')
      }
      {
        name: 'create-1-day-views'
        scriptVersion: 'v1.0.0'
        script: loadTextContent('scripts/kustodb.telemetries.views.1d.kql')
      }
    ]
  }
}

// -----------------------------------------------------------------------
// Real-time Communication
// -----------------------------------------------------------------------

module eventHubNamespace './modules/eventhub.namespace.module.bicep' = {
  name: 'eventHubNamespaceDeploy'
  params: {
    prefix: prefix
    location: location
    tags: tags
    sku: 'Basic'
    units: 4
  }
}

module rawEventHub './modules/eventhub.module.bicep' = {
  name: 'rawEventHubDeploy'
  params: {
    name: 'telemetry-raw'
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: 3
  }
}

module metricsEventHub './modules/eventhub.module.bicep' = {
  name: 'metricsEventHubDeploy'
  params: {
    name: 'telemetry-metrics'
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: 6
  }
}

module dataEventHub './modules/eventhub.module.bicep' = {
  name: 'dataEventHubDeploy'
  params: {
    name: 'telemetry-data'
    namespaceName: eventHubNamespace.outputs.name
    partitionCount: 6
  }
}

// -----------------------------------------------------------------------
// Access Management
// -----------------------------------------------------------------------

module redisAccessPolicy './modules/redis.accesspolicy.module.bicep' = {
  name: 'redisAccessPolicyDeploy'
  params: {
    redisName: redisCache.outputs.name
    name: 'telemetry-ingestion-silver-layer-policy'
    permissions: '~service.ingestion:silver:* +@read +@write'
  }
}
