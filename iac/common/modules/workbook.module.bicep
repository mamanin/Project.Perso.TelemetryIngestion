// =======================================================================
// Workbook Module
// -----------------------------------------------------------------------
// Module: workbook.module.bicep
// Description: Deploys Azure Workbooks
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.insights/workbooks
// =======================================================================

import { WorkbookTemplating } from '../types/workbook.templating.type.bicep'
import { toUpperSnakeCase } from '../functions/tools.functions.bicep'

@description('Location where the workbook will be deployed')
param location string = resourceGroup().location

@description('Display name of the workbook shown in the Azure Portal')
param displayName string

@description('List of Resource to map in the templated workbook')
param replaceTokens WorkbookTemplating[]

@description('Workbook JSON template (placeholders will be replaced with resource IDs)')
param workbookTemplate string

@description('Tags to apply to the workbook resource')
param tags object = {}

@description('Workbook category (default: workbook)')
param category string = 'workbook'

var resourceTypeToPrefix = {
  Workbook: 'WORKBOOK'
  AppInsights: 'APP_INSIGHTS'
  LogAnalytics: 'LOG_ANALYTICS'
  EventHub: 'EVENT_HUB'
  ContainerApp: 'CONTAINER_APP'
  WebPlan: 'WEB_PLAN'
  ServiceBus: 'SERVICE_BUS'
  CosmosDb: 'COSMOS_DB'
}

@description('List of placeholders and their corresponding resource IDs to replace in the workbook template')
var tokens = [
  for res in replaceTokens: {
    placeholder: '${resourceTypeToPrefix[res.typeTemplate]}_${toUpperSnakeCase(res.key)}'
    resourceId: res.id
  }
]

@description('Processed workbook template with resource ID placeholders replaced')
var processedTemplate = reduce(
  array(tokens),
  workbookTemplate,
  (current, replacement) => replace(current, '{{${replacement.placeholder}}}', replacement.resourceId)
)

// Determine the default sourceId for the workbook
// The sourceId defines the default monitoring context for the workbook:
// - If there’s exactly one App Insights, use its ID
// - If there’s exactly one LAW, use its ID
// - If multiple sources or none are provided, use 'azure monitor' (global workbook)
var appInsights = [for res in replaceTokens: res.typeTemplate == 'AppInsights' ? res : null]
var logAnalyticsWorkspaces = [for res in replaceTokens: res.typeTemplate == 'LogAnalytics' ? res : null]

@description('Default sourceId for the workbook, determined based on the provided resources')
var defaultSourceId = (length(appInsights) + length(logAnalyticsWorkspaces)) > 1
  ? 'azure monitor'
  : length(appInsights) == 1
      ? appInsights[0].?id
      : length(logAnalyticsWorkspaces) == 1 ? logAnalyticsWorkspaces[0].?id : 'azure monitor'

resource workbook 'Microsoft.Insights/workbooks@2023-06-01' = {
  name: guid(resourceGroup().id, 'Microsoft.Insights/workbooks', displayName)
  location: location
  tags: tags
  kind: 'shared'
  properties: {
    displayName: displayName
    serializedData: processedTemplate
    category: category
    sourceId: defaultSourceId
  }
}

@description('The ID of the deployed workbook')
output workbookId string = workbook.id

@description('The name of the deployed workbook')
output workbookName string = workbook.name
