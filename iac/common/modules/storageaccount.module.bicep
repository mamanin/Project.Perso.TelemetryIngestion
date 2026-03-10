// =======================================================================
// Storage Account Module
// -----------------------------------------------------------------------
// Module: storage.module.bicep
// Description: Deploys an Azure Storage Account
// See: https://learn.microsoft.com/en-us/azure/templates/microsoft.storage/storageaccounts
// =======================================================================

import { BuildResourceName } from '../functions/core.functions.bicep'

@description('The naming prefix for resource naming')
@minLength(4)
param prefix string

@description('The number for resource naming')
@minLength(3)
param number string = '001'

@description('The location of the storage account')
param location string = resourceGroup().location

@description('Resource tags')
param tags object = {}

@description('The SKU name for the storage account')
@allowed([
  'Standard_LRS'
  'Standard_GRS'
  'Standard_RAGRS'
  'Standard_ZRS'
  'Premium_LRS'
  'Premium_ZRS'
])
param skuName string = 'Standard_LRS'

@description('The kind of storage account')
@allowed([
  'StorageV2'
  'BlobStorage'
  'FileStorage'
  'BlockBlobStorage'
])
param kind string = 'StorageV2'

@description('Indicates whether to allow blob public access')
param blobPublicAccess bool = false

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-06-01' = {
  name: BuildResourceName(prefix, 'sto', number)
  location: location
  sku: {
    name: skuName
  }
  kind: kind
  identity: {
    type: 'SystemAssigned'
  }
  tags: tags
  properties: {
    accessTier: 'Hot'
    allowBlobPublicAccess: blobPublicAccess
    allowSharedKeyAccess: false
    supportsHttpsTrafficOnly: true
    minimumTlsVersion: 'TLS1_2'
    encryption: {
      services: {
        blob: {
          enabled: true
        }
        file: {
          enabled: true
        }
        table: {
          enabled: true
        }
        queue: {
          enabled: true
        }
      }
      requireInfrastructureEncryption: true
      keySource: 'Microsoft.Storage'
    }
  }
}

@description('The ID of the storage account')
output id string = storageAccount.id

@description('The name of the storage account')
output name string = storageAccount.name

@description('The tables endpoint of the storage account')
output tablesEnpoint string = storageAccount.properties.primaryEndpoints.table

@description('The blobs endpoint of the storage account')
output blobsEnpoint string = storageAccount.properties.primaryEndpoints.blob

@description('The files endpoint of the storage account')
output filesEnpoint string = storageAccount.properties.primaryEndpoints.file

@description('The queues endpoint of the storage account')
output queuesEnpoint string = storageAccount.properties.primaryEndpoints.queue
