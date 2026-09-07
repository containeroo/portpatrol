# Runtime behavior

## Concurrent checks

N.E.V.E.R. starts all configured targets concurrently. A successful target stops checking while unsuccessful targets continue retrying. The process exits with status `0` only after every target succeeds.

## Attempts and failure

The global `--max-attempts` setting applies to targets without an override. A per-target `max-attempts` takes precedence. The value `0` means unlimited attempts.

If any target exhausts its attempts, N.E.V.E.R. cancels the remaining checks and exits with status `1`.

## Retry timing

Each target uses its own interval. An interval of `0` inherits `--default-interval`, which defaults to two seconds.

With linear backoff, checks retain the configured interval. Exponential backoff increases the delay after unsuccessful attempts. `max-interval` caps that growth when set.

## HTTP responses

HTTP checks validate the configured expected status codes. Redirects are followed by default and the final response is validated. Redirect following and the maximum number of hops are independently configurable. The request timeout covers the full operation, including redirects.

## Cancellation

Process cancellation stops active network operations through their request contexts. This includes HTTP requests, TCP connection attempts, ICMP operations, and retry waits.
