@description('Container App Probes Type')
@export()
type ContainerAppProbesType = {
  @description('Type of the probe')
  type: 'Startup' | 'Liveness' | 'Readiness'

  @description('Probe scheme - determines probe type')
  scheme: 'HTTP' | 'TCP'

  @description('Port for TCP probe')
  port: int?
  
  @description('Path for HTTP probe')
  path: string?

  @description('Period in seconds between probe checks')
  period: int
}
