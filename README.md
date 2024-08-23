# THOR - TCP and HTTP ObserveR

`THOR` is a simple Go application with zero external dependencies that checks if a specified `TCP` or `HTTP` target is available. It continuously attempts to connect to the specified target at regular intervals until the target becomes available or the program is terminated.

## How It Works

`THOR` performs the following steps:

- **Configuration**: The application is configured using environment variables, allowing flexibility and easy integration into various environments like Docker or Kubernetes.
- **Target Connection Attempts**: It repeatedly attempts to connect to the specified `TCP` or `HTTP` target based on the configured `INTERVAL` and `DIAL_TIMEOUT`.
- **Logging**: `THOR` logs connection attempts, successes, and failures. You can enable additional logging fields to include more context in the logs.
- **Exit Status**:

  - If the target becomes available, `THOR` exits with a status code of `0` (success).
  - If the program is terminated before the target is ready, it exits with a non-zero status code, typically `1`, indicating failure or interruption.

## Environment Variables

`THOR` accepts the following environment variables:

### Common Variables

- `TARGET_ADDRESS`: The address of the target in the following format:
  - **TCP**: `host:port` (required). If `tcp://` is used as the scheme, the `CHECK_TYPE` can be omitted.
  - **HTTP**: `scheme://host:port` (required).
- `TARGET_NAME`: The name assigned to the target (optional, default: inferred from `TARGET_ADDRESS`). If not specified, the name will be derived from the host portion of the target address. For example, `http://postgres.default.svc.cluster.local:5432` would be inferred as `postgres.default.svc.cluster.local`.
- `INTERVAL`: The interval between connection attempts (optional, default: `2s`).
- `DIAL_TIMEOUT`: The maximum time allowed for each connection attempt (optional, default: `1s`).
- `CHECK_TYPE`: Specifies the type of check to perform: `tcp` or `http` (optional, default: inferred from `TARGET_ADDRESS`).
- `LOG_ADDITIONAL_FIELDS`: Enables logging of additional fields (optional, default: `false`).

### HTTP-Specific Variables

- `METHOD`: The HTTP method to use (optional, default: `GET`).
- `HEADERS`: Comma-separated list of HTTP headers to include in the request (optional).
  For Example:
  - `Authorization=Bearer token`
  - `Content-Type=application/json,Accept=application/json`
- `EXPECTED_STATUSES`: Comma-separated list of expected HTTP status codes or ranges (optional, default: `200`).
  `THOR` considers the check successful if the target returns any status code listed in `EXPECTED_STATUSES`. You can specify individual status codes or ranges of codes. For example:

  - Individual status codes: `200,301,404`
  - Ranges of status codes: `200,300-302`
  - Combination of both: `200,301-302,404,500-502`

  Examples:

  - `200,301-302,404`: The check will be considered successful if the target responds with `200`, `301`, `302`, or `404`.
  - `200,300-302,500-502`: The check will succeed if the target responds with `200`, any status in the range `300-302`, or any status in the range `500-502`.

    This flexibility allows you to precisely define what HTTP responses are acceptable for your service, ensuring that the application only proceeds when the target is in the desired state.

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
    image: ghcr.io/containeroo/thor:latest
    env:
      - name: TARGET_ADDRESS
        value: valkey.default.svc.cluster.local:6379
  - name: wait-for-postgres
    image: ghcr.io/containeroo/thor:latest
    env:
      - name: TARGET_NAME
        value: PostgreSQL
      - name: TARGET_ADDRESS
        value: postgres.default.svc.cluster.local:5432
      - name: CHECK_TYPE
        value: tcp # Specify the type of check, either tcp or http
      - name: INTERVAL
        value: "5s" # Specify the interval duration, e.g., 2 seconds
      - name: DIAL_TIMEOUT
        value: "5s" # Specify the dial timeout duration, e.g., 2 seconds
      - name: LOG_ADDITIONAL_FIELDS
        value: "true"
  - name: wait-for-webapp
    image: ghcr.io/containeroo/thor:latest
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

## Usage Scenarios

- Kubernetes initContainers: Use `THOR` to delay the start of a service until its dependencies are ready, ensuring reliable startup sequences.
- Startup Scripts: Include `THOR` in deployment scripts to ensure that services wait for dependencies before proceeding.
- CI/CD Pipelines: Use `THOR` in CI/CD pipelines to wait for services to be ready before running integration tests.
