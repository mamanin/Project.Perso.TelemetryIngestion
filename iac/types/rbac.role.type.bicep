@description('Rbac Role Type')
@export()
type RbacRoleType = {
  @description('Identifier of the role')
  id: string

  @description('Name of the role')
  description: string
}
