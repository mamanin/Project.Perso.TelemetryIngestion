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

  @description('Number of consecutive failures for the probe to be considered failed after having succeeded')
  failureThreshold: int?

  @description('Period in seconds between probe checks')
  periodSeconds: int

  @description('Number of seconds after the container has started before the probe is initiated')
  timeoutSeconds: int?

  @description('Number of seconds after the container has started before liveness probes are initiated')
  initialDelaySeconds: int?
}
