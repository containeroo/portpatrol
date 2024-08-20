# TOAST - Trigger Operations After Successful Test

`TOAST` is a simple Go application with zero external dependencies that checks if a specified TCP or HTTP target is available. It continuously attempts to connect to the specified target at regular intervals until the target becomes available or the program is terminated.

## Environment Variables

`TOAST` accepts the following environment variables:

### Common Variables

- `TARGET_ADDRESS`: The address of the target in the following format:
  - **TCP**: `host:port` (required).
  - **HTTP**: `host:port` (required).
- `TARGET_NAME`: The name of the target to check (optional, default: inferred from `TARGET_ADDRESS`)\*.
- `INTERVAL`: The interval between connection attempts (optional, default: `2s`).
- `DIAL_TIMEOUT`: The timeout for each connection attempt (optional, default: `1s`).
- `CHECK_TYPE`: The type of check to perform, either `tcp` or `http` (optional, default: inferred from `TARGET_ADDRESS`).
- `LOG_ADDITIONAL_FIELDS`: Log additional fields (optional, default: `false`).

### HTTP-Specific Variables

- `METHOD`: The HTTP method to use (optional, default: `GET`).
- `HEADERS`: Comma-separated list of HTTP headers to include in the request (optional).
- `EXPECTED_STATUSES`: Comma-separated list of expected HTTP status codes or ranges (optional, default: `200`).

**\*** If `TARGET_NAME` is not set, the name will be inferred from the host part of the target address as follows: `postgres.default.svc.cluster.local:5432` will be inferred as `postgres`.

## Behavior Flowchart

```mermaid
graph TD;
    A[Start] --> B[Attempt to connect to TARGET_ADDRESS];
    B -->|Connection successful| C[Target is ready];
    B -->|Connection failed| D[Wait for retry INTERVAL];
    D --> B;
    C --> E[End];
    F[Program terminated or canceled] --> E;
```

## Logging

With the `LOG_ADDITIONAL_FIELDS` environment variable set to true, additional fields will be logged.

### With additional fields

```text
ts=2024-07-05T13:08:20+02:00 level=INFO msg="Waiting for PostgreSQL to become ready..." dial_timeout="1s" interval="2s" target_address="postgres.default.svc.cluster.local:5432" target_name="PostgreSQL" version="0.0.22"
ts=2024-07-05T13:08:21+02:00 level=WARN msg="PostgreSQL is not ready ✗" dial_timeout="1s" error="dial tcp: lookup postgres.default.svc.cluster.local: i/o timeout" interval="2s" target_address="postgres.default.svc.cluster.local:5432" target_name="PostgreSQL" version="0.0.22"
ts=2024-07-05T13:08:24+02:00 level=WARN msg="PostgreSQL is not ready ✗" dial_timeout="1s" error="dial tcp: lookup postgres.default.svc.cluster.local: i/o timeout" interval="2s" target_address="postgres.default.svc.cluster.local:5432" target_name="PostgreSQL" version="0.0.22"
ts=2024-07-05T13:08:27+02:00 level=WARN msg="PostgreSQL is not ready ✗" dial_timeout="1s" error="dial tcp: lookup postgres.default.svc.cluster.local: i/o timeout" interval="2s" target_address="postgres.default.svc.cluster.local:5432" target_name="PostgreSQL" version="0.0.22"
ts=2024-07-05T13:08:27+02:00 level=INFO msg="PostgreSQL is ready ✓" dial_timeout="1s" error="dial tcp: lookup postgres.default.svc.cluster.local: i/o timeout" interval="2s" target_address="postgres.default.svc.cluster.local:5432" target_name="PostgreSQL" version="0.0.22"
```

### Without additional fields

```text
time=2024-07-12T12:44:41.494Z level=INFO msg="Waiting for PostgreSQL to become ready..."
time=2024-07-12T12:44:41.512Z level=WARN msg="PostgreSQL is not ready ✗"
time=2024-07-12T12:44:43.532Z level=WARN msg="PostgreSQL is not ready ✗"
time=2024-07-12T12:44:45.552Z level=INFO msg="PostgreSQL is ready ✓"
```

## Kubernetes initContainer Configuration

Configure your Kubernetes deployment to use this init container:

```yaml
initContainers:
  - name: wait-for-valkey
    image: ghcr.io/containeroo/toast:latest
    env:
      - name: TARGET_NAME
        value: valkey
      - name: TARGET_ADDRESS
        value: valkey.default.svc.cluster.local:6379
      - name: CHECK_TYPE
        value: tcp # Specify the type of check, either tcp or http
      - name: INTERVAL
        value: "2s" # Specify the interval duration, e.g., 2 seconds
      - name: DIAL_TIMEOUT
        value: "2s" # Specify the dial timeout duration, e.g., 2 seconds
      - name: LOG_ADDITIONAL_FIELDS
        value: "true"
  - name: wait-for-postgres
    image: ghcr.io/containeroo/toast:latest
    env:
      - name: TARGET_NAME
        value: PostgreSQL
      - name: TARGET_ADDRESS
        value: postgres.default.svc.cluster.local:5432
      - name: CHECK_TYPE
        value: tcp # Specify the type of check, either tcp or http
      - name: INTERVAL
        value: "2s" # Specify the interval duration, e.g., 2 seconds
      - name: DIAL_TIMEOUT
        value: "2s" # Specify the dial timeout duration, e.g., 2 seconds
      - name: LOG_ADDITIONAL_FIELDS
        value: "true"
  - name: wait-for-webapp
    image: ghcr.io/containeroo/toast:latest
    env:
      - name: TARGET_NAME
        value: webapp
      - name: TARGET_ADDRESS
        value: webapp.default.svc.cluster.local:8080
      - name: CHECK_TYPE
        value: http # Specify the type of check, either tcp or http
      - name: METHOD
        value: "GET"
      - name: HEADERS
        value: "Authorization=Bearer token"
      - name: EXPECTED_STATUSES
        value: "200,202"
      - name: INTERVAL
        value: "2s" # Specify the interval duration, e.g., 2 seconds
      - name: DIAL_TIMEOUT
        value: "2s" # Specify the dial timeout duration, e.g., 2 seconds
      - name: LOG_ADDITIONAL_FIELDS
        value: "true"
```
