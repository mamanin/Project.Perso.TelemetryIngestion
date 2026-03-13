@description('Constants for ingestion service')
@export()
var ingestionConstants = {
  eventhub: {
    rawName: 'telemetry-raw'
    metricsName: 'telemetry-metrics'
    dataName: 'telemetry-data'
  }
  consumerGroups: {
    bronze: 'cg-bronze'
    silver: 'cg-silver'
    gold: 'cg-gold'
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
    processingBatchSize: 50
    processingWorkerCount: 20
    scalingEventThreshold: 50*3
    scalingActivationEventThreshold: 50*6
  }
  silver: {
    partitionCount: 8
    processingBatchSize: 500
    processingWorkerCount: 12
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
