# N.E.V.E.R.

**Network Endpoint Validation with Endless Retries** waits for network dependencies to become ready. It checks HTTP, TCP, and ICMP targets concurrently and exits only when every target succeeds.

N.E.V.E.R. is designed for startup coordination, particularly as a Kubernetes init container. It is also useful in scripts, container entrypoints, and deployment pipelines that must wait for external services.

## What it does

- Runs multiple named targets concurrently.
- Retries indefinitely by default.
- Supports global and per-target attempt limits.
- Supports linear and exponential retry backoff.
- Accepts command-line flags and environment variables.
- Exits with status `0` when every target is ready and `1` when a target exhausts its attempts.

## Start here

- [Getting started](getting-started.md) covers the target syntax and a first run.
- [Configuration](configuration/index.md) is the complete flag and environment reference.
- [Kubernetes](kubernetes.md) provides init-container examples and ICMP permissions.
- [Runtime behavior](behavior.md) explains concurrency, retries, redirects, and exit conditions.

{{subpages}}
