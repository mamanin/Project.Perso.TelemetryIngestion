@description('Describes the script to run on Kusto Cluster Database creation.')
@export()
type KustoClusterDatabasesScriptDescription = {
  @description('The name of the script.')
  name: string

  @description('The force update tag to trigger script re-execution when changed.')
  scriptVersion: string

  @description('The content of the script to run.')
  @secure()
  script: string
}
