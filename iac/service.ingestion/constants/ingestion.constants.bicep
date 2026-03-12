@description('Constants for ingestion service')
@export()
var ingestionConstants = {
  eventhub: {
    rawName: 'telemetry-raw'
    metricsName: 'telemetry-metrics'
    dataName: 'telemetry-data'
  }
  dataExplorer: {
    telemetryDatabaseName: 'telemetries'
    telemetryTableName: 'metrics'
  }
  tableStorage: {
    checkpointsTableName: 'partition-checkpoints'
  }
  bronze: {
    partitionCount: 6
    processingBatchSize: 75
    processingWorkerCount: 8
    scalingEventThreshold: 75*3
    scalingActivationEventThreshold: 75*6
  }
  silver: {
    partitionCount: 8
    processingBatchSize: 500
    processingWorkerCount: 6
    scalingEventThreshold: 500*3
    scalingActivationEventThreshold: 500*6
  }
  gold: {
    partitionCount: 6
    processingBatchSize: 500
    processingWorkerCount: 4
    scalingEventThreshold: 500*3
    scalingActivationEventThreshold: 500*6
  }
}
