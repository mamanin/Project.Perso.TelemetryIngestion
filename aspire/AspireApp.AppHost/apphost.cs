#:package Aspire.Hosting.Azure.EventHubs@13.1.0
#:package Aspire.Hosting.Azure.Kusto@13.1.0-preview.1.25616.3
#:package Aspire.Hosting.Azure.Storage@13.1.0
#:package Aspire.Hosting.Redis@13.1.0
#:sdk Aspire.AppHost.Sdk@13.1.0
#:package CommunityToolkit.Aspire.Hosting.Golang@13.0.0
#:package LupusBytes.Aspire.Hosting.Azure.EventHubs.LiveExplorer@2.0.0

var builder = DistributedApplication.CreateBuilder(args);

// Add Azure services
var kusto = builder
    .AddAzureKustoCluster("pocitpdex001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent));
kusto.AddReadWriteDatabase("telemetries");

var eventHub = builder
    .AddAzureEventHubs("pocitpevh001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent));
var raw = eventHub.AddHub("telemetry-raw");
var metrics = eventHub.AddHub("telemetry-metrics");
var data = eventHub.AddHub("telemetry-data");

var storage = builder
    .AddAzureStorage("pocitpsto001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent));
var blob = storage.AddBlobContainer("partition-workers");

var cache = builder
    .AddRedis("pocitpred001")
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
    .AddGolangApp("bronze", "../../src/Service.Ingestion/cmd/bronze-worker")
    .WithReference(raw)
    .WithReference(metrics)
    .WithOtlpExporter(OtlpProtocol.Grpc);

var silver = builder
    .AddGolangApp("silver", "../../src/Service.Ingestion/cmd/silver-worker")
    .WithReference(eventHub)
    .WithReference(metrics)
    .WithReference(data)
    .WithReference(cache)
    .WithReference(blob)
    .WithOtlpExporter(OtlpProtocol.Grpc);

var gold = builder
    .AddGolangApp("gold", "../../src/Service.Ingestion/cmd/gold-worker")
    .WithReference(kusto)
    .WithReference(data)
    .WithOtlpExporter(OtlpProtocol.Grpc);

// Add test sender
_ = builder
    .AddGolangApp("test-sender", "../../src/Service.Ingestion/cmd/test-sender")
    .WithReference(eventHub)
    .WithReference(metrics)
    .WithReference(bronze)
    .WithReference(silver)
    .WithReference(gold);

await builder
    .Build()
    .RunAsync();
