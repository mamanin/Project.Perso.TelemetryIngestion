@description('Describes the SKU of the Redis cache to deploy.')
@export()
type RedisSkuDescription = {
  @description('The type of Redis cache to deploy.')
  name: 'Basic' | 'Standard' | 'Premium'

  @minValue(0)
  @maxValue(6)
  @description('The size of the Redis cache to deploy (C:[0, 6] P:[1, 4]).')
  // See: https://azure.microsoft.com/en-us/pricing/details/cache
  capacity: int
}
