- [Telemetry formats](#telemetry-formats)
  - [Metric Format](#metric-format)
  - [Telemetry v2 format](#telemetry-v2-format)
  - [Telemetry v1 format](#telemetry-v1-format)
  - [Telemetry legacy format](#telemetry-legacy-format)

# Telemetry formats

This document defines the telemetry formats expected by the bronze layer of the telemetry ingestion workflow.

## Metric Format

The metric format represents a single metric data point produced internally by the bronze layer after processing raw telemetry events. It is the format used for all events passed to the next layers of the workflow.

> The naming of the fields is the smallest possible to reduce the size of the events. <br>
> As the metric format is only used internally in the ingestion workflow, we can afford to use a less readable format that is more efficient in terms of size. (To be honest, this is a micro-optimization that happened to be implemented, so it is documented here 😊)

Json example:
```json
{
    "sp": "/device/12345678/cpu",
    "n": "temperature",
    "v": 72.4,
    "u": "°C",
    "_ts": 1743580800
}
```

Json schema:
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "MetricEvent",
    "type": "object",
    "properties": {
        "sp": {
            "type": "string",
            "pattern": "^/[^/]+/[0-9]{4,8}(/[^/]+)?$",
            "description": "Sensor path. Format: /device/{device_id} or /{device}/{device_id}/{sensor}"
        },
        "n": {
            "type": "string",
            "description": "Name of the metric"
        },
        "v": {
            "anyOf": [
                { "type": "string" },
                { "type": "number" }
            ],
            "description": "Value of the metric"
        },
        "u": {
            "type": "string",
            "description": "Unit of the metric"
        },
        "_ts": {
            "type": "integer",
            "minimum": 0,
            "description": "Unix epoch timestamp in seconds"
        }
    },
    "required": [
        "sp",
        "n",
        "v",
        "_ts"
    ],
    "additionalProperties": false
}
```

## Telemetry v2 format

This input format is the one used in the IoT-oriented professional project I was working on while developing this project. It was designed to handle multiple metrics with multiple timestamps in a single event. It is a flexible format that makes it easy to configure the telemetry update frequency defined by the device without risking the loss of data points.

Json example:
```json
{
    "version": "v2.0",
    "device_id": "12345678",
    "timestamp": 1743580800,
    "metrics": [
        {
            "name": "temperature",
            "origin": "cpu",
            "unit": "°C",
            "measures": [
                { "timestamp": 1743580800, "value": 72.4 },
                { "timestamp": 1743580860, "value": 73.1 }
            ]
        },
        {
            "name": "usage",
            "origin": "cpu",
            "unit": "%",
            "measures": [
                { "timestamp": 1743580800, "value": 45.2 },
                { "timestamp": 1743580860, "value": 47.8 }
            ]
        },
        {
            "name": "latency",
            "origin": "network",
            "unit": "µs",
            "measures": [
                { "timestamp": 1743580800, "value": 1250 },
                { "timestamp": 1743580860, "value": 980 }
            ]
        }
    ]
}
```

Json schema:
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "TelemetryV2Event",
    "type": "object",
    "properties": {
        "version": {
            "type": "string",
            "const": "v2.0"
        },
        "device_id": {
            "type": "string",
            "pattern": "^[0-9]{4,8}$",
            "description": "Device identifier (4–8 digits)"
        },
        "timestamp": {
            "type": "integer",
            "minimum": 0,
            "description": "Unix epoch timestamp in seconds."
        },
        "metrics": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "name": {
                        "type": "string"
                    },
                    "origin": {
                        "type": "string"
                    },
                    "unit": {
                        "type": "string"
                    },
                    "measures": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "properties": {
                                "timestamp": {
                                    "type": "integer"
                                },
                                "value": {
                                    "anyOf": [
                                        { "type": "string" },
                                        { "type": "number" }
                                    ],
                                    "description": "Value of the telemetry data point"
                                }
                            },
                            "required": [
                                "timestamp",
                                "value"
                            ],
                            "additionalProperties": false
                        }
                    }
                },
                "required": [
                    "name",
                    "origin",
                    "measures"
                ],
                "additionalProperties": false
            }
        }
    },
    "required": [
        "version",
        "device_id",
        "timestamp",
        "metrics"
    ],
    "additionalProperties": false
}
```

## Telemetry v1 format

This input format does not exist, but I created it as an example of how the bronze layer can handle multiple telemetry formats. It is defined as a simpler version of the telemetry v2 format, where each metric is represented by a single value instead of an array of measures.
> This is a just-for-fun format 🥰

Json example:
```json
{
    "version": "v1.0",
    "device_id": "12345678",
    "timestamp": 1743580800,
    "data": [
        {
            "name": "temperature",
            "origin": "cpu",
            "unit": "°C",
            "value": 72.4
        },
        {
            "name": "usage",
            "origin": "cpu",
            "unit": "%",
            "value": 45.2
        },
        {
            "name": "latency",
            "origin": "network",
            "unit": "µs",
            "value": 1250
        }
    ]
}
```

Json schema:
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "TelemetryV1Event",
    "type": "object",
    "properties": {
        "version": {
            "type": "string",
            "const": "v1.0",
            "description": "Version of the telemetry format (must be 'v1.0')"
        },
        "device_id": {
            "type": "string",
            "pattern": "^[0-9]{4,8}$",
            "description": "Device identifier (4–8 digits)"
        },
        "timestamp": {
            "type": "integer",
            "minimum": 0,
            "description": "Unix epoch timestamp in seconds."
        },
        "data": {
            "type": "array",
            "description": "Array of telemetry data points",
            "items": {
                "type": "object",
                "description": "Telemetry data points",
                "properties": {
                    "name": {
                        "type": "string",
                        "description": "Name of the telemetry data point"
                    },
                    "origin": {
                        "type": "string",
                        "enum": [
                            "device",
                            "cpu",
                            "environment",
                            "network",
                            "storage",
                            "battery"
                        ],
                        "description": "Origin of the telemetry data point"
                    },
                    "unit": {
                        "type": "string",
                        "description": "Unit of the telemetry data point"
                    },
                    "value": {
                        "anyOf": [
                            { "type": "string" },
                            { "type": "number" }
                        ],
                        "description": "Value of the telemetry data point"
                    }
                },
                "required": [
                    "name",
                    "origin",
                    "value"
                ],
                "additionalProperties": false
            }
        }
    },
    "required": [
        "version",
        "device_id",
        "timestamp",
        "data"
    ],
    "additionalProperties": false
}
```

## Telemetry legacy format

This input format is the legacy format produced by devices in the IoT-oriented professional project I was working on while developing this project. It is highly specific and not flexible, but it is still used by some legacy devices that we need to support in the bronze layer.

Json example:
```json
{
    "device_id": "12345678",
    "timestamp": 1743580800,
    "state": "on",
    "uptime": 1440.0,
    "cpu_usage": 45.2,
    "cpu_memory": 68.5,
    "cpu_temperature": 72.4,
    "network_latency": 1250.0
}
```

Json schema:
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "TelemetryLegacyEvent",
    "type": "object",
    "properties": {
        "device_id": {
            "type": "string",
            "pattern": "^[0-9]{4,8}$",
            "description": "Device identifier (4–8 digits)"
        },
        "timestamp": {
            "type": "integer",
            "minimum": 0,
            "description": "Unix epoch timestamp in seconds"
        },
        "state": {
            "type": "string",
            "description": "Device state"
        },
        "uptime": {
            "type": "number",
            "description": "Uptime in minutes"
        },
        "cpu_usage": {
            "type": "number",
            "description": "CPU usage percent"
        },
        "cpu_memory": {
            "type": "number",
            "description": "CPU memory percent"
        },
        "cpu_temperature": {
            "type": "number",
            "description": "CPU temperature in °C"
        },
        "network_latency": {
            "type": "number",
            "description": "Network latency in microseconds"
        }
    },
    "required": [
        "device_id",
        "timestamp",
        "state",
        "uptime",
        "cpu_usage",
        "cpu_memory",
        "cpu_temperature",
        "network_latency"
    ],
    "additionalProperties": false
}
```
