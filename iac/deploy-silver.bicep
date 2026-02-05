// =======================================================================
// Silver Injestion Service Deployment
// -----------------------------------------------------------------------
// Module: deploy-silver.bicep
// Description: Deploys the infrastructure resources required for the
//       silver layer of the telemetry ingestion service.
// =======================================================================

import { rbacRoles } from './constants/rbac.role.constants.bicep'
import { BuildResourceName } from 'functions/core.functions.bicep'

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
  application: 'silver'
}

@description('The version of the application to deploy')
param version string

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
  name: BuildResourceName(prefix, 'evh', '001')
}

resource redis 'Microsoft.Cache/redis@2024-11-01' existing = {
  name: BuildResourceName(prefix, 'red', '001')
}

// -----------------------------------------------------------------------
// Service Core
// -----------------------------------------------------------------------

module identity './modules/identity.userassigned.module.bicep' = {
  name: 'identityDeploy'
  params: {
    prefix: prefix
    number: '002'
    location: location
    tags: tags
  }
}

module containerRegistryRoleAssignment './modules/rbac/rbac.containerregistry.module.bicep' = {
  name: 'containerRegistryRoleAssignmentDeploy'
  params: {
    name: containerRegistry.name
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.containerregistry['Acr Pull']
    ]
  }
}

module storageAccountRoleAssignment './modules/rbac/rbac.storageaccount.module.bicep' = {
  name: 'storageAccountRoleAssignmentDeploy'
  params: {
    name: storageAccount.name
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.storageaccount['Storage Blob Data Contributor']
    ]
  }
}

module metricsEventHubRoleAssignment './modules/rbac/rbac.eventhub.module.bicep' = {
  name: 'metricsEventHubRoleAssignmentDeploy'
  params: {
    namespaceName: namespace.name
    name: 'telemetry-metrics'
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.eventhub['Azure Event Hubs Data Receiver']
    ]
  }
}

module dataEventHubRoleAssignment './modules/rbac/rbac.eventhub.module.bicep' = {
  name: 'dataEventHubRoleAssignmentDeploy'
  params: {
    namespaceName: namespace.name
    name: 'telemetry-data'
    principalId: identity.outputs.principalId
    roles: [
      rbacRoles.eventhub['Azure Event Hubs Data Sender']
    ]
  }
}

module redisRoleAssignment './modules/rbac/rbac.redis.module.bicep' = {
  name: 'redisRoleAssignmentDeploy'
  params: {
    name: redis.name
    principalId: identity.outputs.principalId
    policyNames: [
      'telemetry-ingestion-silver-layer-policy'
    ]
  }
}

// TODO: check for :
// - add GOMAXPROCS '1' to env
// - add import _ "go.uber.org/automaxprocs" to go app
module containerApp './modules/containerapp.module.bicep' = {
  name: 'containerAppDeploy'
  params: {
    prefix: prefix
    number: '001'
    location: location
    tags: tags
    managedIdentityId: identity.outputs.id
    containerAppEnvironmentId: containerAppEnvironment.id
    containerServer: containerRegistry.properties.loginServer
    containerImage: '/wildgrowth/silver:${version}'
    applicationPort: 8080
    minReplicas: 0
    maxReplicas: 6 // TODO: use variable
    scaleRules: [{
      custom: {
        type: 'azure-eventhub'
        identity: identity.outputs.id
        metadata: {
            eventHubNamespace: namespace.name
            eventHubName: 'telemetry-metrics'
            storageAccountName: storageAccount.name
            blobContainer: 'partition-checkpoints'
            checkpointStrategy: 'goSdk' // TODO: blobMetadata ?
            unprocessedEventThreshold: '400' // TODO: use variable: batch size * x (2?)
            activationUnprocessedEventThreshold: '200' // TODO: use variable: batch size * x (1?)
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
        value: 'silver-layer'
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
