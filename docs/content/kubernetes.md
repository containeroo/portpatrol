# Kubernetes

N.E.V.E.R. can run before an application container and prevent it from starting until its dependencies are reachable.

## HTTP and TCP dependencies

```yaml
initContainers:
  - name: wait-for-services
    image: ghcr.io/containeroo/never:latest
    args:
      - --http.api.address=http://api.default.svc.cluster.local:9000/healthz
      - --http.api.expected-status-codes=200,204
      - --tcp.database.address=postgres.default.svc.cluster.local:5432
      - --tcp.cache.address=valkey.default.svc.cluster.local:6379
      - --tcp.cache.interval=5s
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
```

## Secrets in headers

Use a resolved header value to avoid putting a secret directly in the pod arguments:

```yaml
initContainers:
  - name: wait-for-api
    image: ghcr.io/containeroo/never:latest
    args:
      - --http.api.address=http://api.default.svc.cluster.local/healthz
      - --http.api.header=Authorization=env:BEARER_TOKEN
    envFrom:
      - secretRef:
          name: api-credentials
```

## ICMP permissions

Only ICMP checks need additional permissions. Add the `NET_RAW` capability:

```yaml
initContainers:
  - name: wait-for-host
    image: ghcr.io/containeroo/never:latest
    args:
      - --icmp.host.address=hostname.example.com
      - --icmp.host.timeout=2s
    securityContext:
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        add: ["NET_RAW"]
```

HTTP and TCP checks do not require additional Linux capabilities.
