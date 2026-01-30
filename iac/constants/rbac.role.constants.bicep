// https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles
@description('The built-in role definition IDs to assign to the principal')
@export()
var rbacRoles = {
  appconfig: {
    'App Configuration Contributor': {
      id: 'fe86443c-f201-4fc4-9d2a-ac61149fbda0'
      description: 'Grants permission for all management operations, except purge, for App Configuration resources.'
    }
    'App Configuration Data Owner': {
      id: '5ae67dd6-50cb-40e7-96ff-dc2bfa4b606b'
      description: 'Allows full access to App Configuration data.'
    }
    'App Configuration Data Reader': {
      id: '516239f1-63e1-4d78-a4de-a74fb236a071'
      description: 'Allows read access to App Configuration data.'
    }
    'App Configuration Reader': {
      id: '175b81b9-6e0d-490a-85e4-0d422273c10c'
      description: 'Grants permission for read operations for App Configuration resources.'
    }
  }
  containerregistry : {
    'Acr Delete': {
      id: 'c2f4ef07-c644-48eb-af81-4b1b4947fb11'
      description: 'Delete repositories, tags, or manifests from a container registry.'
    }
    'Acr Image Signer': {
      id: '6cef56e8-d556-48e5-a04f-b8e64114680f'
      description: 'Push trusted images to or pull trusted images from a container registry enabled for content trust.'
    }
    'Acr Pull': {
      id: '7f951dda-4ed3-4680-a7ca-43fe172d538d'
      description: 'Pull artifacts from a container registry.'
    }
    'Acr Push': {
      id: '8311e382-0749-4cb8-b61a-304f252e45ec'
      description: 'Push artifacts to or pull artifacts from a container registry.'
    }
  }
  cosmos: {
    'Cosmos DB Account Reader Role': {
      id: 'fbdf93bf-df7d-467e-a4d2-9458aa1360c8'
      description: 'Can read account data but cannot read or write container or database data.'
    }
    'Cosmos DB Operator': {
      id: '230815da-be43-4aae-9cb4-875f7bd000aa'
      description: 'Can provision Cosmos DB accounts but cannot access account keys or connection strings.'
    }
    'DocumentDB Account Contributor': {
      id: '5bd9cd88-fe45-4216-938b-f97437e15450'
      description: 'Can manage Cosmos DB accounts, but not access to them.'
    }
  }
  eventhub: {
    'Azure Event Hubs Data Owner': {
      id: 'f526a384-b230-433a-b45c-95f59c4a2dec'
      description: 'Allows for full access to Azure Event Hubs resources.'
    }
    'Azure Event Hubs Data Receiver': {
      id: 'a638d3c7-ab3a-418d-83e6-5f17a39d4fde'
      description: 'Allows receive access to Azure Event Hubs resources.'
    }
    'Azure Event Hubs Data Sender': {
      id: '2b629674-e913-4c01-ae53-ef4638d8f975'
      description: 'Allows send access to Azure Event Hubs resources.'
    }
  }
  keyvault: {
    'Key Vault Administrator': {
      id: '00482a5a-887f-4fb3-b363-3b7fe8e74483'
      description: 'Perform all data plane operations on key vault and all objects in it.'
    }
    'Key Vault Certificates Officer': {
      id: 'a4417e6f-fecd-4de8-b567-7b0420556985'
      description: 'Perform any action on certificates, except manage permissions.'
    }
    'Key Vault Contributor': {
      id: 'f25e0fa2-a7c8-4377-a976-54943a77a395'
      description: 'Manage key vaults, but does not allow assigning roles or accessing secrets, keys, or certificates.'
    }
    'Key Vault Crypto Officer': {
      id: '14b46e9e-c2b7-41b4-b07b-48a6ebf60603'
      description: 'Perform any action on keys, except manage permissions.'
    }
    'Key Vault Crypto User': {
      id: '12338af0-0e69-4776-bea7-57ae8d297424'
      description: 'Perform cryptographic operations using keys.'
    }
    'Key Vault Reader': {
      id: '21090545-7ca7-4776-b22c-e363652d74d2'
      description: 'Read metadata of key vaults and its objects. Cannot read sensitive values.'
    }
    'Key Vault Secrets Officer': {
      id: 'b86a8fe4-44ce-4948-aee5-eccb2c155cd7'
      description: 'Perform any action on secrets, except manage permissions.'
    }
    'Key Vault Secrets User': {
      id: '4633458b-17de-408a-b874-0445c86b69e6'
      description: 'Read secret contents.'
    }
  }
  servicebus: {
    'Azure Service Bus Data Owner': {
      id: '090c5cfd-751d-490a-894a-3ce6f1109419'
      description: 'Allows full access to Azure Service Bus resources.'
    }
    'Azure Service Bus Data Receiver': {
      id: '4f6d3b9b-027b-4f4c-9142-0e5a2a2247e0'
      description: 'Allows receive access to Azure Service Bus resources.'
    }
    'Azure Service Bus Data Sender': {
      id: '69a216fc-b8fb-44d8-bc22-1f3c2cd27a39'
      description: 'Allows send access to Azure Service Bus resources.'
    }
  }
  storageaccount: {
    'Storage Account Contributor': {
      id: '17d1049b-9a84-46fb-8f53-869881c3d3ab'
      description: 'Manage storage accounts and access account keys for Shared Key authorization.'
    }
    'Storage Blob Data Contributor': {
      id: 'ba92f5b4-2d11-453d-a403-e96b0029c9fe'
      description: 'Read, write, and delete Azure Storage containers and blobs.'
    }
    'Storage Blob Data Owner': {
      id: 'b7e6dc6d-f1e8-4753-8033-0f276bb0955b'
      description: 'Full access to Storage blob containers and data, including POSIX access control.'
    }
    'Storage Blob Data Reader': {
      id: '2a2b9908-6ea1-4ae2-8e65-a410df84e7d1'
      description: 'Read and list Azure Storage containers and blobs.'
    }
    'Storage Queue Data Contributor': {
      id: '974c5e8b-45b9-4653-ba55-5f855dd0fb88'
      description: 'Read, write, and delete Azure Storage queues and queue messages.'
    }
    'Storage Queue Data Message Processor': {
      id: '8a0f0c08-91a1-4084-bc3d-661d67233fed'
      description: 'Peek, retrieve, and delete messages from Azure Storage queues.'
    }
    'Storage Queue Data Message Sender': {
      id: 'c6a89b2d-59bc-44d0-9896-0f6e12d7b80a'
      description: 'Add messages to Azure Storage queues.'
    }
    'Storage Queue Data Reader': {
      id: '19e7f393-937e-4f77-808e-94535e297925'
      description: 'Read and list Azure Storage queues and queue messages.'
    }
    'Storage Table Data Contributor': {
      id: '0a9a7e1f-b9d0-4cc4-a60d-0319b160aaa3'
      description: 'Read, write, and delete Azure Storage tables and entities.'
    }
    'Storage Table Data Reader': {
      id: '76199698-9eea-4c19-bc75-cec21354c6b6'
      description: 'Read access to Azure Storage tables and entities.'
    }
  }
  webpubsub: {
    'Web PubSub Service Owner': {
      id: '12cf5a90-567b-43ae-8102-96cf46c7d9b4'
      description: 'Full access to Azure Web PubSub Service REST APIs.'
    }
    'Web PubSub Service Reader': {
      id: 'bfb1c7d2-fb1a-466b-b2ba-aee63b92deaf'
      description: 'Read-only access to Azure Web PubSub Service REST APIs.'
    }
    'SignalR/Web PubSub Contributor': {
      id: '8cf5e20a-e4b2-4e9d-b3a1-5ceb692c2761'
      description: 'Create, Read, Update, and Delete SignalR and Web PubSub service resources.'
    }
  }
}
