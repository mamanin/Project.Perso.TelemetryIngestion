@description('Constants for ingestion service')
@export()
var ingestionConstants = {
  eventhub: {
    rawName: 'telemetry-raw'
    metricsName: 'telemetry-metrics'
    dataName: 'telemetry-data'
  }
  bronze: {
    partitionCount: 4
    processingBatchSize: 100
    processingWorkerCount: 4
    scalingEventThreshold: 100*3
    scalingActivationEventThreshold: 100*6
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
    processingWorkerCount: 6
    scalingEventThreshold: 500*3
    scalingActivationEventThreshold: 500*6
  }
}
