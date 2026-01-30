// =======================================================================
// Event Hub Namespace Eventhub Module
// -----------------------------------------------------------------------
// Module: eventhub.module.bicep
// Description: Deploys Azure Event Hub Namespace Eventhub
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.eventhub/namespaces/eventhubs
// =======================================================================

@description('The name of the Event Hub')
@minLength(4)
param name string

@description('The name of the existing namespace')
param namespaceName string

@description('The number of partitions for the Event Hub')
@minValue(1)
param partitionCount int = 1

@description('The retention time in hours for the Event Hub')
@minValue(1)
param retentionTimeInHours int = 10

resource namespace 'Microsoft.EventHub/namespaces@2025-05-01-preview' existing = {
  name: namespaceName
}

resource eventHub 'Microsoft.EventHub/namespaces/eventhubs@2025-05-01-preview' = {
  parent: namespace
  name: name
  properties: {
    partitionCount: partitionCount
    status: 'Active'
    retentionDescription: {
      retentionTimeInHours: retentionTimeInHours
      cleanupPolicy: 'Delete'
    }
  }
}
