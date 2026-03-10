// =======================================================================
// Log Analytics Module
// -----------------------------------------------------------------------
// Module: loganalytics.module.bicep
// Description: Deploys a Log Analytics workspace
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.operationalinsights/workspaces
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Log Analytics workspace')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The sku of the Log Analytics workspace')
@allowed([
  'PerGB2018'
  'Free'
  'Standalone'
  'PerNode'
  'CapacityReservation'
])
param sku string = 'PerGB2018'

@description('The number of days to retain logs')
param retentionInDays int = 30

resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2025-07-01' = {
  name: BuildResourceName(prefix, 'law', number)
  location: location
  tags: tags
  properties: {
    sku: {
      name: sku
    }
    retentionInDays: retentionInDays
    workspaceCapping: {
      dailyQuotaGb: 1
    }
  }
}

@description('The ID of the Log Analytics workspace')
output id string = logAnalytics.id

@description('The name of the Log Analytics workspace')
output name string = logAnalytics.name
