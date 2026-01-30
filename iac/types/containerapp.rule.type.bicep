@description('Container App Rule Type')
@export()
type ContainerAppScaleRule = {
  custom: ContainerAppCustomRuleType
}

@description('Container App Custom Rule Type')
@export()
type ContainerAppCustomRuleType = {
  @description('Type of the scaling rule')
  // For 'azure-eventhub', refer to https://keda.sh/docs/2.18/scalers/azure-event-hub/
  type: 'http' | 'cpu' | 'memory' | 'custom' | 'azure-eventhub'

  @description('The identity used to authenticate the scaling rule')
  identity: string

  @description('Metadata for the scaling rule')
  metadata: object
}
