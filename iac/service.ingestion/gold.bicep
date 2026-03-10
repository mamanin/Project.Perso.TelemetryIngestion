// =======================================================================
// Gold Injestion Service Deployment
// -----------------------------------------------------------------------
// Module: gold.bicep
// Description: Deploys the infrastructure resources required for the
//       gold layer of the telemetry ingestion service.
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
  application: 'gold'
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

resource kustoCluster 'Microsoft.Kusto/clusters@2024-04-13' existing = {
  name: BuildResourceName(prefix, 'adx', '001')
}

// -----------------------------------------------------------------------
// Service Core
// -----------------------------------------------------------------------

module identity '../common/modules/identity.userassigned.module.bicep' = {
  name: 'identityDeploy'
  params: {
    prefix: prefix
    number: '003'
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

module dataEventHubRoleAssignment '../common/modules/rbac/rbac.eventhub.module.bicep' = {
  name: 'dataEventHubRoleAssignmentDeploy'
  params: {
    namespaceName: namespace.name
    name: ingestionConstants.eventhub.dataName
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.eventhub['Azure Event Hubs Data Receiver']
    ]
  }
}

module kustoClusterDatabaseRoleAssignment '../common/modules/rbac/rbac.kusto.database.module.bicep' = {
  name: 'kustoClusterDatabaseRoleAssignmentDeploy'
  params: {
    kustoClusterName: kustoCluster.name
    databaseName: 'telemetries'
    principalId: identity.outputs.principalId
    roles: [
      'Ingestor'
    ]
  }
}

module containerApp '../common/modules/containerapp.module.bicep' = {
  name: 'containerAppDeploy'
  dependsOn: [
    containerRegistryRoleAssignment
    storageAccountRoleAssignment
    dataEventHubRoleAssignment
    kustoClusterDatabaseRoleAssignment
  ]
  params: {
    prefix: prefix
    number: '003'
    location: location
    tags: tags
    managedIdentityId: identity.outputs.id
    containerAppEnvironmentId: containerAppEnvironment.id
    containerServer: containerRegistry.properties.loginServer
    containerImage: '/service.ingestion/gold:${version}'
    applicationPort: 8080
    minReplicas: 0
    maxReplicas: ingestionConstants.gold.partitionCount
    scaleRules: [{
      name: 'eventhub-scaler'
      custom: {
        type: 'azure-eventhub'
        identity: identity.outputs.id
        metadata: {
            eventHubNamespace: namespace.name
            eventHubName: ingestionConstants.eventhub.dataName
            storageAccountName: storageAccount.name
            blobContainer: 'partition-checkpoints'
            checkpointStrategy: 'blobMetadata'
            unprocessedEventThreshold: string(ingestionConstants.gold.scalingEventThreshold)
            activationUnprocessedEventThreshold: string(ingestionConstants.gold.scalingActivationEventThreshold)
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

      // TODO: add app settings

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
        value: 'layer.gold'
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
