// =======================================================================
// Application Insights Module
// -----------------------------------------------------------------------
// Module: appinsights.module.bicep
// Description: Deploys Azure Application Insights
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.insights/components
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Application Insights resource')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The ID of the Log Analytics workspace to associate with Application Insights')
param workspaceId string

resource appInsights 'Microsoft.Insights/components@2020-02-02' = {
  name: BuildResourceName(prefix, 'ain', number)
  location: location
  kind: 'web'
  tags: tags
  properties: {
    Application_Type: 'web'
    WorkspaceResourceId: workspaceId
    IngestionMode: 'LogAnalytics'
    publicNetworkAccessForIngestion: 'Enabled'
    publicNetworkAccessForQuery: 'Enabled'
  }
}

@description('The ID of the Application Insights resource')
output id string = appInsights.id

@description('The name of the Application Insights resource')
output name string = appInsights.name

@description('The name of the Application Insights resource')
output connectionString string = appInsights.properties.ConnectionString
