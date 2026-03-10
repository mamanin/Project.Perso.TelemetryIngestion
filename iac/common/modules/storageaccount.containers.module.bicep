// =======================================================================
// Storage Containers Module
// -----------------------------------------------------------------------
// Module: storage.containers.module.bicep
// Description: Deploys Azure Storage Blob Containers
// See: 
//    - https://learn.microsoft.com/en-us/azure/templates/microsoft.storage/storageaccounts/blobservices
//    - https://learn.microsoft.com/en-us/azure/templates/microsoft.storage/storageaccounts/blobservices/containers
// =======================================================================

import { StorageAccountContainerType } from '../types/storageaccount.container.type.bicep'

@description('The name of the existing storage account')
param storageAccountName string

@description('Array of blob to create')
param containers StorageAccountContainerType[]

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-06-01' existing = {
  name: storageAccountName
}

resource blobService 'Microsoft.Storage/storageAccounts/blobServices@2025-06-01' = {
  parent: storageAccount
  name: 'default'
}

resource blobContainers 'Microsoft.Storage/storageAccounts/blobServices/containers@2025-06-01' = [for container in containers: {
  parent: blobService
  name: container.name
  properties: {
    publicAccess: container.publicAccess
  }
}]

@description('The endpoint URL for accessing the storage blobs')
output storageBlobsEndpoint string = storageAccount.properties.primaryEndpoints.blob
