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
    processingWorkerCount: 16
    scalingEventThreshold: 50*3
    scalingActivationEventThreshold: 50*6
  }
  silver: {
    // ⚠️ There is a redis connection limitation, I would not go above 8 partitions with 7 workers. I haven't found a fix for this.
    partitionCount: 8
    processingBatchSize: 500
    processingWorkerCount: 7
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
