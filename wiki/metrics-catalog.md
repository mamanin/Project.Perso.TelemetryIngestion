- [Metrics catalog](#metrics-catalog)

# Metrics catalog

> ℹ️ These rules and units are fictional and are only defined for the purpose of this project.

| Source        | Metric name        | Validation rules                                             | Conversion rules  |
| ------------- | ------------------ | ------------------------------------------------------------ | ----------------- |
| `device`      | `state`            | `on`, `warning`, `error`                                     | ø                 |
| `device`      | `uptime`           | `0 ≤ x`                                                      | Converts to `s`   |
| `device`      | `battery_level`    | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `device`      | `firmware_version` | string                                                       | ø                 |
| `cpu`         | `usage`            | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `cpu`         | `usage_user`       | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `cpu`         | `usage_system`     | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `cpu`         | `usage_idle`       | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `cpu`         | `memory`           | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `cpu`         | `temperature`      | `0 ≤ x ≤ 400`                                                | Converts to `K`   |
| `cpu`         | `voltage`          | `0.5 ≤ x ≤ 3`                                                | Converts to `V`   |
| `cpu`         | `fan_speed`        | `0 ≤ x ≤ 10 000`                                             | ø                 |
| `environment` | `temperature`      | `200 ≤ x ≤ 350`                                              | Converts to `K`   |
| `environment` | `humidity`         | `0 ≤ x ≤ 100`                                                | ø                 |
| `environment` | `pressure`         | `80 000 ≤ x ≤ 120 000`                                       | Converts to `Pa`  |
| `environment` | `wind_speed`       | `0 ≤ x ≤ 150`                                                | Converts to `m/s` |
| `environment` | `wind_direction`   | `0 ≤ x ≤ 360`                                                | ø                 |
| `environment` | `rainfall`         | `0 ≤ x`                                                      | Converts to `m`   |
| `environment` | `lux`              | `0 ≤ x`                                                      | ø                 |
| `environment` | `co2`              | `0 ≤ x ≤ 5 000`                                              | ø                 |
| `network`     | `bytes_received`   | `0 ≤ x`                                                      | Converts to `B`   |
| `network`     | `bytes_sent`       | `0 ≤ x`                                                      | Converts to `B`   |
| `network`     | `packets_dropped`  | `0 ≤ x`                                                      | ø                 |
| `network`     | `latency`          | `0 ≤ x`                                                      | Converts to `s`   |
| `network`     | `bandwidth_usage`  | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `storage`     | `total_size`       | `0 ≤ x`                                                      | Converts to `B`   |
| `storage`     | `used_size`        | `0 ≤ x`                                                      | Converts to `B`   |
| `storage`     | `free_size`        | `0 ≤ x`                                                      | Converts to `B`   |
| `storage`     | `usage_percent`    | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `storage`     | `read_throughput`  | `0 ≤ x`                                                      | Converts to `B/s` |
| `storage`     | `write_throughput` | `0 ≤ x`                                                      | Converts to `B/s` |
| `storage`     | `health`           | `good`, `warning`, `critical`                                | ø                 |
| `battery`     | `level`            | `0 ≤ x ≤ 100`                                                | Unit must be `%`  |
| `battery`     | `voltage`          | `0 ≤ x ≤ 100`                                                | Converts to `V`   |
| `battery`     | `current`          | `-1 000 ≤ x ≤ 1 000`                                         | Converts to `A`   |
| `battery`     | `power`            | `0 ≤ x`                                                      | Converts to `W`   |
| `battery`     | `temperature`      | `200 ≤ x ≤ 400` K                                            | Converts to `K`   |
| `battery`     | `status`           | `charging`, `discharging`, `full`, `not_charging`, `unknown` | ø                 |
| `battery`     | `cycle_count`      | `0 ≤ x`                                                      | ø                 |
| `battery`     | `time_to_empty`    | `0 ≤ x`                                                      | Converts to `s`   |