// =======================================================================
// Bronze Injestion Service Deployment
// -----------------------------------------------------------------------
// Module: bronze.bicep
// Description: Deploys the infrastructure resources required for the
//       bronze layer of the telemetry ingestion service.
// =======================================================================

import { rbacRoles } from '../common/constants/rbac.role.constants.bicep'
import { BuildResourceName } from '../common/functions/core.functions.bicep'
import { ingestionConstants } from './constants/ingestion.constants.bicep'

// -----------------------------------------------------------------------
// Parameters and Variables
// -----------------------------------------------------------------------

@description('The version of the application to deploy')
param version string

@description('The name of the resource group to deploy to')
var location = resourceGroup().location

@description('The name prefix of the resource to deploy')
var prefix = 'tispoc'

@description('The tags to apply to all resources')
var tags = {
  project: prefix
  'managed-by': 'bicep'
  service: 'ingestion'
  application: 'bronze'
}

// -----------------------------------------------------------------------
// Existing Resources
// -----------------------------------------------------------------------

resource containerRegistry 'Microsoft.ContainerRegistry/registries@2025-05-01-preview' existing = {
  name: BuildResourceName(prefix, 'acr', '001')
}

resource containerAppEnvironment 'Microsoft.App/managedEnvironments@2024-03-01' existing = {
  name: BuildResourceName(prefix, 'ace', '001')
}

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-01-01' existing = {
  name: BuildResourceName(prefix, 'sto', '001')
}

resource namespace 'Microsoft.EventHub/namespaces@2025-05-01-preview' existing = {
  name: BuildResourceName(prefix, 'ehn', '001')
}

// -----------------------------------------------------------------------
// Service Core
// -----------------------------------------------------------------------

module identity '../common/modules/identity.userassigned.module.bicep' = {
  name: 'identityDeploy'
  params: {
    prefix: prefix
    number: '001'
    location: location
    tags: tags
  }
}

module containerRegistryRoleAssignment '../common/modules/rbac/rbac.containerregistry.module.bicep' = {
  name: 'containerRegistryRoleAssignmentDeploy'
  params: {
    name: containerRegistry.name
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.containerregistry['Acr Pull']
    ]
  }
}

module storageAccountRoleAssignment '../common/modules/rbac/rbac.storageaccount.module.bicep' = {
  name: 'storageAccountRoleAssignmentDeploy'
  params: {
    name: storageAccount.name
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.storageaccount['Storage Blob Data Contributor']
    ]
  }
}

module rawEventHubRoleAssignment '../common/modules/rbac/rbac.eventhub.module.bicep' = {
  name: 'rawEventHubRoleAssignmentDeploy'
  params: {
    namespaceName: namespace.name
    name: ingestionConstants.eventhub.rawName
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.eventhub['Azure Event Hubs Data Receiver']
    ]
  }
}

module metricsEventHubRoleAssignment '../common/modules/rbac/rbac.eventhub.module.bicep' = {
  name: 'metricsEventHubRoleAssignmentDeploy'
  params: {
    namespaceName: namespace.name
    name: ingestionConstants.eventhub.metricsName
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.eventhub['Azure Event Hubs Data Sender']
    ]
  }
}

module containerApp '../common/modules/containerapp.module.bicep' = {
  name: 'containerAppDeploy'
  dependsOn: [
    containerRegistryRoleAssignment
    storageAccountRoleAssignment
    rawEventHubRoleAssignment
    metricsEventHubRoleAssignment
  ]
  params: {
    prefix: prefix
    number: '001'
    location: location
    tags: tags
    managedIdentityId: identity.outputs.id
    containerAppEnvironmentId: containerAppEnvironment.id
    containerServer: containerRegistry.properties.loginServer
    containerImage: '/service.ingestion/bronze:${version}'
    applicationPort: 8080
    activeRevisionsMode: 'Single' // Required as partitions can't be shared between revisions
    minReplicas: 0
    maxReplicas: ingestionConstants.bronze.partitionCount
    scaleRules: [{
      name: 'eventhub-scaler'
      custom: {
        type: 'azure-eventhub'
        identity: identity.outputs.id
        metadata: {
            eventHubNamespace: namespace.name
            eventHubName: ingestionConstants.eventhub.rawName
            storageAccountName: storageAccount.name
            blobContainer: ingestionConstants.tableStorage.checkpointsTableName
            checkpointStrategy: 'blobMetadata'
            unprocessedEventThreshold: string(ingestionConstants.bronze.scalingEventThreshold)
            activationUnprocessedEventThreshold: string(ingestionConstants.bronze.scalingActivationEventThreshold)
        }
      }
    }]
    env: [
      {
        name: 'AZURE_TENANT_ID'
        value: subscription().tenantId
      }
      {
        name: 'AZURE_CLIENT_ID'
        value: identity.outputs.clientId
      }

      // App settings
      {
        name: 'Processor__Workers'
        value: string(ingestionConstants.bronze.processingWorkerCount)
      }
      {
        name: 'Container__Url'
        value: '${storageAccount.properties.primaryEndpoints.blob}${ingestionConstants.tableStorage.checkpointsTableName}'
      }
      {
        name: 'EventHub__FullyQualifiedNamespace'
        value: '${namespace.name}.servicebus.windows.net'
      }
      {
        name: 'EventHub__Subscriber__Name'
        value: ingestionConstants.eventhub.rawName
      }
      {
        name: 'EventHub__Subscriber__BatchSize'
        value: string(ingestionConstants.bronze.processingBatchSize)
      }
      {
        name: 'EventHub__Subscriber__PrefetchSize'
        value: string(ingestionConstants.bronze.processingBatchSize * (ingestionConstants.bronze.processingWorkerCount + ingestionConstants.bronze.processingWorkerCount / 2))
      }
      {
        name: 'EventHub__Publisher__Name'
        value: ingestionConstants.eventhub.metricsName
      }

      // OpenTelemetry settings
      // OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_EXPORTER_OTLP_PROTOCOL are added automatically by the container environment
      // https://learn.microsoft.com/en-us/azure/container-apps/opentelemetry-agents?tabs=bicep%2Carm-example#environment-variables
      // But we still need to override OTEL_EXPORTER_OTLP_ENDPOINT to remove the `http://` prefix for the go.opentelemetry.io library to work correctly 
      {
        name: 'OTEL_ENABLED'
        value: 'true'
      }
      {
        name: 'OTEL_EXPORTER_OTLP_ENDPOINT'
        value: 'k8se-otel.k8se-apps.svc.cluster.local:4317'
      }
      {
        name: 'OTEL_SERVICE_NAME'
        value: 'service.ingestion'
      }
      {
        name: 'OTEL_SERVICE_LAYER'
        value: 'layer.bronze'
      }
      {
        name: 'OTEL_SERVICE_VERSION'
        value: version
      }
    ]
    probes: [
      {
        type: 'Startup'
        scheme: 'TCP'
        port: 8080
        failureThreshold: 60
        periodSeconds: 1
      }
      {
        type: 'Liveness'
        scheme: 'TCP'
        port: 8081
        failureThreshold: 5
        periodSeconds: 5
        initialDelaySeconds: 5
      }
      {
        type: 'Readiness'
        scheme: 'TCP'
        port: 8082
        failureThreshold: 30
        periodSeconds: 3
        initialDelaySeconds: 5
      }
    ]
  }
}
