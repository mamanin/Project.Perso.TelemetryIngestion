// =======================================================================
// Diagnostic Infrastructure Deployment
// -----------------------------------------------------------------------
// Module: main.bicep
// Description: Deploys the ingestion diagnostic infrastructure resources 
//       to monitor and log the other services.
// =======================================================================

import { BuildResourceName } from '../common/functions/core.functions.bicep'

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
  service: 'diagnostic'
}

// -----------------------------------------------------------------------
// Existing Resources
// -----------------------------------------------------------------------

resource appInsights 'Microsoft.Insights/components@2020-02-02' existing = {
  name: BuildResourceName(prefix, 'ain', '001')
}

resource eventHubNamespace 'Microsoft.EventHub/namespaces@2025-05-01-preview' existing = {
  name: BuildResourceName(prefix, 'ehn', '001')
}

resource bronzeContainerApp 'Microsoft.App/containerApps@2025-10-02-preview' existing = {
  name: BuildResourceName(prefix, 'aca', '001')
}

resource silverContainerApp 'Microsoft.App/containerApps@2025-10-02-preview' existing = {
  name: BuildResourceName(prefix, 'aca', '002')
}

// -----------------------------------------------------------------------
// Service Core
// -----------------------------------------------------------------------

module ingestionWorkbook '../common/modules/workbook.module.bicep' = {
  name: 'ingestionWorkbookDeploy'
  params: {
    displayName: 'Ingestion Diagnostic Workbook'
    location: location
    tags: tags
    workbookTemplate: loadTextContent('./workbooks/ingestion-workbook.json')
    replaceTokens: [
      {
        typeTemplate: 'AppInsights'
        key: 'ingestion'
        id: appInsights.id
      }
      {
        typeTemplate: 'EventHub'
        key: 'ingestion-telemetry'
        id: eventHubNamespace.id
      }
      {
        typeTemplate: 'ContainerApp'
        key: 'ingestion-bronze'
        id: bronzeContainerApp.id
      }
      {
        typeTemplate: 'ContainerApp'
        key: 'ingestion-silver'
        id: silverContainerApp.id
      }
    ]
  }
}

