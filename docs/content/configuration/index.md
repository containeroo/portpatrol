# Configuration

N.E.V.E.R. supports global settings and any number of dynamically named targets.

## Global flags

| Flag                 | Environment variable      | Type     | Default | Description                                                    |
| -------------------- | ------------------------- | -------- | ------- | -------------------------------------------------------------- |
| `--default-interval` | `NEVER__DEFAULT_INTERVAL` | duration | `2s`    | Default delay between checks. Targets can override this value. |
| `--max-attempts`     | `NEVER__MAX_ATTEMPTS`     | int      | `0`     | Attempts before failure. `0` retries indefinitely.             |
| `--log-format`       | `NEVER__LOG_FORMAT`       | enum     | `json`  | Log format: `json` or `text`.                                  |
| `--version`          | —                         | bool     | `false` | Print the version and exit.                                    |
| `--help`, `-h`       | —                         | bool     | `false` | Print command help.                                            |

## Target settings

Target flags use this format:

```text
--<TYPE>.<IDENTIFIER>.<PROPERTY>=<VALUE>
```

Target environment variables use this format:

```text
NEVER__<TYPE>_<IDENTIFIER>_<PROPERTY>=<VALUE>
```

The identifier becomes the target name unless the target's `name` property overrides it.

- [HTTP and HTTPS](http.md)
- [TCP](tcp.md)
- [ICMP](icmp.md)
- [Resolved values](resolved-values.md)

{{subpages}}
