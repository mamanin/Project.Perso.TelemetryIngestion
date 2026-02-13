#:package Aspire.Hosting.Azure.EventHubs@13.1.1
#:package Aspire.Hosting.Azure.Kusto@13.1.1-preview.1.26105.8
#:package Aspire.Hosting.Azure.Storage@13.1.1
#:package Aspire.Hosting.Redis@13.1.1
#:sdk Aspire.AppHost.Sdk@13.1.1
#:package CommunityToolkit.Aspire.Hosting.Golang@13.0.0
#:package LupusBytes.Aspire.Hosting.Azure.EventHubs.LiveExplorer@2.0.0

var builder = DistributedApplication.CreateBuilder(args);

// Add Azure services
var cluster = builder
    .AddAzureKustoCluster("tispocadx001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent))
    .WithEndpoint("http", e => e.Port = 56700);
var metricsDb = cluster
    .AddReadWriteDatabase("telemetries")
//     .WithCreationScript(
// """
// .execute database script <|
//     .create-merge table metrics (
//         timestamp: datetime, 
//         device_id: string,
//         sensor: string, 
//         metric: string,
//         value: dynamic,
//         unit: string
//         );
//     .alter table metrics policy streamingingestion enable;
// """
//     )
    ;

var eventHub = builder
    .AddAzureEventHubs("tispocevh001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent));
var raw = eventHub.AddHub("telemetry-raw");
var metrics = eventHub.AddHub("telemetry-metrics");
var data = eventHub.AddHub("telemetry-data");

var storage = builder
    .AddAzureStorage("tispocsto001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent));
var blob = storage.AddBlobContainer("partition-checkpoints");

var cache = builder
    .AddRedis("tispocred001")
    .WithRedisInsight(ri => ri.WithLifetime(ContainerLifetime.Persistent))
    .WithLifetime(ContainerLifetime.Persistent);

// Add Live Explorer
// Enable Event Hubs Live Explorer for debugging
// _ = builder
//     .AddAzureEventHubsLiveExplorer("event-hubs-explorer")
//     .WithReference(metrics)
//     .WithReference(data)
//     .WithLifetime(ContainerLifetime.Persistent);

// Add Golang workers
var bronze = builder
    .AddGolangApp("bronze", "../src/service.ingestion/cmd/bronze-worker")
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WithReference(eventHub)
    .WithReference(raw)
    .WithReference(metrics)
    .WithReference(blob)
    .WithEnvironment("deployment-environment", "Aspire")
    .WithEnvironment("PROCESSOR_WORKERCOUNT", "3")
    .WithEnvironment("TELEMETRY_RAW_BATCHSIZE", "100")
    .WithOtlpExporter(OtlpProtocol.Grpc);

var silver = builder
    .AddGolangApp("silver", "../src/service.ingestion/cmd/silver-worker")
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WaitFor(cache)
    .WithReference(eventHub)
    .WithReference(metrics)
    .WithReference(data)
    .WithReference(blob)
    .WithReference(cache)
    .WithEnvironment("deployment-environment", "Aspire")
    .WithEnvironment("PROCESSOR_WORKERCOUNT", "4")
    .WithEnvironment("TELEMETRY_METRICS_BATCHSIZE", "400")
    .WithOtlpExporter(OtlpProtocol.Grpc);

var gold = builder
    .AddGolangApp("gold", "../src/service.ingestion/cmd/gold-worker")
    .WaitFor(metricsDb)
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WithReference(metricsDb)
    .WithReference(eventHub)
    .WithReference(data)
    .WithReference(blob)
    .WithEnvironment("deployment-environment", "Aspire")
    .WithEnvironment("PROCESSOR_WORKERCOUNT", "4")
    .WithEnvironment("TELEMETRY_DATA_BATCHSIZE", "400")
    .WithOtlpExporter(OtlpProtocol.Grpc);

// Add web server
_ = builder
    .AddGolangApp("web-server", "../src/service.data/cmd/server")
    .WaitFor(metricsDb)
    .WithReference(metricsDb)
    .WithEnvironment("deployment-environment", "Aspire")
    .WithHttpEndpoint(name: "localhost", port: 8080, isProxied: false)
    .WithOtlpExporter(OtlpProtocol.Grpc);

// Add test sender
_ = builder
    .AddGolangApp("v2-sender", "../src/service.ingestion/cmd/telemetry-senders/v2-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc)
    .WithExplicitStart();

_ = builder
    .AddGolangApp("v1-sender", "../src/service.ingestion/cmd/telemetry-senders/v1-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc)
    .WithExplicitStart();

_ = builder
    .AddGolangApp("legacy-sender", "../src/service.ingestion/cmd/telemetry-senders/legacy-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc)
    .WithExplicitStart();

await builder
    .Build()
    .RunAsync();
