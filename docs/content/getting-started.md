# Getting started

N.E.V.E.R. is distributed as a container image and as release binaries.

## Run the container

Wait for an HTTP endpoint and a TCP service in parallel:

```sh
docker run --rm ghcr.io/containeroo/never:latest \
  --http.web.address=https://example.com \
  --tcp.database.address=database.example.com:5432
```

The process exits successfully when both targets are ready. By default, unsuccessful checks repeat every two seconds without an attempt limit.

## Define targets

Every target has a type, an identifier, and properties:

```text
--<TYPE>.<IDENTIFIER>.<PROPERTY>=<VALUE>
```

Supported target types are `http`, `tcp`, and `icmp`. The identifier distinguishes targets of the same type.

```sh
never \
  --http.frontend.address=https://frontend.example.com/ready \
  --http.backend.address=https://backend.example.com/ready \
  --tcp.database.address=postgres.example.com:5432
```

## Use environment variables

Flags have equivalent environment variables. Remove the leading `--`, replace dots and hyphens with underscores, uppercase the name, and add the `NEVER__` prefix:

```text
--http.web.address=https://example.com
NEVER__HTTP_WEB_ADDRESS=https://example.com
```

Command-line flags take precedence over environment variables. See [Configuration](configuration/index.md) for the complete mapping.

## Control retries

Use a global limit or override it for one target:

```sh
never \
  --max-attempts=20 \
  --http.api.address=https://example.com/ready \
  --http.api.max-attempts=5 \
  --tcp.database.address=database.example.com:5432 \
  --tcp.database.backoff=exponential \
  --tcp.database.max-interval=30s
```

An attempt limit of `0` means unlimited attempts.
