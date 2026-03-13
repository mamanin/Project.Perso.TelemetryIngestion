// =======================================================================
// Event Hub Namespace Module
// -----------------------------------------------------------------------
// Module: eventhub.namespace.module.bicep
// Description: Deploys Azure Event Hub Namespace
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.eventhub/namespaces
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the Event Hub Namespace resource')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('SKU for Event Hub Namespace')
@allowed([
  'Basic'
  'Standard'
  'Premium'
])
param sku string = 'Basic'

@description('Capacity for Event Hub Namespace SKU')
@minValue(1)
param capacity int = 1

@description('Whether to enable auto-inflate for the Event Hub Namespace')
param isAutoInflateEnabled bool = false

@description('Maximum throughput units for auto-inflate (applicable only if auto-inflate is enabled)')
param maximumThroughputUnits int = 0

resource namespace 'Microsoft.EventHub/namespaces@2025-05-01-preview' = {
  name: BuildResourceName(prefix, 'ehn', number)
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  tags: tags
  sku: {
    name: sku
    tier: sku
    capacity: capacity
  }
  properties: {
    minimumTlsVersion: '1.2'
    isAutoInflateEnabled: sku == 'Premium' ? false : isAutoInflateEnabled
    maximumThroughputUnits: sku == 'Premium' ? 0 : maximumThroughputUnits
    kafkaEnabled: true
    zoneRedundant: false
    disableLocalAuth: true
  }
}

@description('The ID of the Event Hub Namespace resource')
output id string = namespace.id

@description('The name of the Event Hub Namespace resource')
output name string = namespace.name
