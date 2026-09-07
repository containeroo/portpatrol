<p align="center">
  <img src="./docs/assets/never.svg" alt="N.E.V.E.R." width="520">
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/containeroo/never.svg)](https://pkg.go.dev/github.com/containeroo/never)
[![Release](https://img.shields.io/github/release/containeroo/never.svg?style=flat-square)](https://github.com/containeroo/never/releases/latest)
[![Unit Tests](https://github.com/containeroo/never/actions/workflows/tests.yml/badge.svg)](https://github.com/containeroo/never/actions/workflows/tests.yml)
[![Lint](https://github.com/containeroo/never/actions/workflows/lint.yml/badge.svg)](https://github.com/containeroo/never/actions/workflows/lint.yml)
[![Build](https://github.com/containeroo/never/actions/workflows/build.yml/badge.svg)](https://github.com/containeroo/never/actions/workflows/build.yml)
[![License](https://img.shields.io/github/license/containeroo/never.svg?style=flat-square)](LICENSE)

**Network Endpoint Validation with Endless Retries** is a small readiness checker for HTTP, TCP, and ICMP targets. It waits for every configured dependency before exiting successfully, making it especially useful as a Kubernetes init container.

## Why N.E.V.E.R.?

- Check multiple HTTP, TCP, and ICMP targets concurrently.
- Configure each target independently with flags or environment variables.
- Retry indefinitely or stop after a configured number of attempts.
- Use linear or exponential backoff.
- Run as a compact container with no service dependencies.

## Quick start

```sh
docker run --rm ghcr.io/containeroo/never:latest \
  --http.web.address=https://example.com \
  --tcp.database.address=database.example.com:5432
```

Target flags follow this pattern:

```text
--<TYPE>.<IDENTIFIER>.<PROPERTY>=<VALUE>
```

The equivalent environment variable starts with `NEVER__`:

```text
NEVER__HTTP_WEB_ADDRESS=https://example.com
```

## Kubernetes

```yaml
initContainers:
  - name: wait-for-dependencies
    image: ghcr.io/containeroo/never:latest
    args:
      - --http.api.address=http://api.default.svc.cluster.local/healthz
      - --tcp.database.address=postgres.default.svc.cluster.local:5432
```

## Documentation

Read the complete configuration and deployment guide at **[containeroo.github.io/never](https://containeroo.github.io/never/)**.

## License

Licensed under the [Apache License 2.0](LICENSE).
