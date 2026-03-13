@description('Describes the SKU of the Redis cache to deploy.')
@export()
type KustoClusterSkuDescription = {
  @description('The type of Kusto cluster to deploy.')
  // See: https://azure.microsoft.com/en-us/pricing/details/data-explorer
  name: 'Dev(No SLA)_Standard_E2a_v4' | 'Standard_E2ads_v5' | 'Standard_E8ads_v5' // To be extended with more SKUs as needed (as this is kinda expensive :3)

  @description('The tier of the Kusto cluster to deploy.')
  tier: 'Basic' | 'Standard'

  @description('The capacity of the Kusto cluster to deploy.')
  @minValue(1)
  capacity: int
}
