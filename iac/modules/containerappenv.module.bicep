// =======================================================================
// Container App Environment Module
// -----------------------------------------------------------------------
// Module: containerappenv.module.bicep
// Description: Deploys an Azure Container App Environment
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.app/managedenvironments
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Container App Environment')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('Log Analytics workspace name')
param logAnalyticsWorkspaceName string

@description('Enable OpenTelemetry endpoints')
param enableOTelEndpoints bool = false

@description('Application Insights Connection String (optional, required if OpenTelemetry is enabled)')
@secure()
param appInsightsConnectionString string = ''

resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2025-02-01' existing = {
  name: logAnalyticsWorkspaceName
}

resource containerAppEnvironment 'Microsoft.App/managedEnvironments@2025-02-02-preview' = {
  name: BuildResourceName(prefix, 'ace', number)
  location: location
  tags: tags
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logAnalytics.properties.customerId
        sharedKey: logAnalytics.listKeys().primarySharedKey
      }
    }
    zoneRedundant: false
    workloadProfiles: [
      {
        name: 'Consumption'
        workloadProfileType: 'Consumption'
      }
    ]
    appInsightsConfiguration: {
      connectionString: appInsightsConnectionString
    }
    openTelemetryConfiguration: enableOTelEndpoints
      ? {
          tracesConfiguration: {
            destinations: [
              'appInsights'
            ]
          }
          logsConfiguration: {
            destinations: [
              'appInsights'
            ]
          }
        }
      : null
  }
}

@description('The ID of the Container App Environment')
output id string = containerAppEnvironment.id

@description('The name of the Container App Environment')
output name string = containerAppEnvironment.name
