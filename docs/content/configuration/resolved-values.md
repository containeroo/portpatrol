# Resolved values

Addresses and HTTP header values can be loaded from environment variables or files instead of being supplied literally.

| Prefix  | Source               | Example                                  | Description                                              |
| ------- | -------------------- | ---------------------------------------- | -------------------------------------------------------- |
| `env:`  | Environment variable | `env:API_ADDRESS`                        | Reads the named environment variable.                    |
| `file:` | Key-value file       | `file:/config/app.txt//KeyName`          | Reads a key from a simple key-value file.                 |
| `json:` | JSON file            | `json:/config/app.json//database.host`   | Reads a dot-separated path from JSON.                     |
| `yaml:` | YAML file            | `yaml:/config/app.yaml//server.port`     | Reads a dot-separated path from YAML.                     |
| `ini:`  | INI file             | `ini:/config/app.ini//Section.Key`       | Reads a section and key from INI.                         |
| none    | Literal              | `https://example.com/healthz`            | Uses the supplied value directly.                        |

This resolution mechanism is separate from the `NEVER__...` environment-variable form used to configure flags.

## Examples

Resolve a target address:

```sh
never --http.api.address=env:API_ADDRESS
```

Resolve only a sensitive header value:

```sh
never \
  --http.api.address=https://example.com/healthz \
  --http.api.header="Authorization=env:API_TOKEN"
```
