@description('Storage Account Container Type')
@export()
type StorageAccountContainerType = {
  @description('Name of the container')
  name: string

  @description('Access level for the container')
  publicAccess: 'None' | 'Blob' | 'Container'
}
