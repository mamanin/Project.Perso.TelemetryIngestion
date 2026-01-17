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
    .AddAzureKustoCluster("pocitpadx001")
    .RunAsEmulator(e => e.WithLifetime(ContainerLifetime.Persistent))
    .AddReadWriteDatabase("telemetries");
// kusto
//     .WithCreationScript(
//         ".create table metrics (timestamp: datetime, device_id: string, metric: string, value: dynamic, unit: string)"
//         + "\n" +
//         ".alter table metrics policy streamingingestion enable"
//     );

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
    .AddGolangApp("bronze", "../../src/cmd/bronze-worker")
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WithReference(eventHub)
    .WithReference(raw)
    .WithReference(metrics)
    .WithReference(blob)
    .WithOtlpExporter(OtlpProtocol.Grpc);

var silver = builder
    .AddGolangApp("silver", "../../src/cmd/silver-worker")
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WithReference(eventHub)
    .WithReference(metrics)
    .WithReference(data)
    .WithReference(cache)
    .WithReference(blob)
    .WithOtlpExporter(OtlpProtocol.Grpc);

var gold = builder
    .AddGolangApp("gold", "../../src/cmd/gold-worker")
    .WaitFor(kusto)
    .WaitFor(eventHub)
    .WaitFor(blob)
    .WithReference(eventHub)
    .WithReference(data)
    .WithReference(kusto)
    .WithReference(kusto)
    .WithReference(blob)
    .WithOtlpExporter(OtlpProtocol.Grpc);

// Add test sender
_ = builder
    .AddGolangApp("v2-sender", "../../src/cmd/telemetry-senders/v2-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc);

_ = builder
    .AddGolangApp("v1-sender", "../../src/cmd/telemetry-senders/v1-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc);

_ = builder
    .AddGolangApp("legacy-sender", "../../src/cmd/telemetry-senders/legacy-sender")
    .WithReference(eventHub)
    .WithReference(raw)
    .WithOtlpExporter(OtlpProtocol.Grpc);

await builder
    .Build()
    .RunAsync();
