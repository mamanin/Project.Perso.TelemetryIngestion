@description('Workbook Templating Type')
@export()
type WorkbookTemplating = {
  @description('Type of the resource to map in the template placeholders')
  typeTemplate: 'Workbook' | 'AppInsights' | 'LogAnalytics' | 'EventHub' | 'ContainerApp' | 'WebPlan' | 'ServiceBus' | 'CosmosDb'

  @description('Key of the resource to be used in the template placeholders')
  key: string

  @description('Identifier of the resource')
  id: string
}
