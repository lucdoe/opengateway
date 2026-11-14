# Operations

## Health check

`GET /__health` answers `200 {"status":"ok"}` while the gateway runs, without plugins or contacting a service, for load balancers and container health checks.

## Request IDs

Every request carries an `X-Request-ID`: the one the client sent, if it is 1 to 128 visible ASCII characters, or else a new random one. The gateway sends it to the service, writes it in the log line and returns it on the response, also on errors it answers itself.

## Errors

When the gateway answers a request itself, rather than passing on the service's response, the body is JSON with a `message`, for example `{"message":"Rate limit exceeded"}`:

| Status | When |
| --- | --- |
| `404`, `405` | no route matches the path, or the method |
| `401`, `403`, `429` | a plugin rejects the request |
| `502`, `504` | the service is unreachable, or too slow to start answering |
| `503` | the service's circuit breaker is open |

## Circuit breaker

Every service has its own circuit breaker. A failure is a service that cannot be reached, times out, or answers `500`, `502`, `503` or `504`. After `ConsecutiveFailures` in a row the breaker opens, and the gateway answers `503` with a `Retry-After` header straight away, without contacting the service. After `OpenTimeout`, trial requests decide whether it closes again. The thresholds are under `Server.CircuitBreaker`, see [Configuration](configuration.md#server-and-store).

## Metrics

With `Server.MetricsPort` set, the gateway serves Prometheus metrics at `/metrics` on that port. It's a port of its own, so you can keep it inside your network: the metrics name every service and route.

| Metric | Type | Labels |
| --- | --- | --- |
| `opengateway_requests_total` | counter | `service`, `route`, `method`, `status` |
| `opengateway_request_duration_seconds` | histogram, 1 ms to 10 s | `service`, `route` |
| `opengateway_circuit_breaker_state` | gauge: `0` closed, `1` half-open, `2` open | `service` |

Every answer counts, including ones from the gateway itself, such as a `429` or `503`. `route` is the path as configured, like `/api/users/{id}`, not the request path.

## Shutdown

On `SIGINT` or `SIGTERM` the gateway stops accepting new connections, lets requests in flight finish for up to `Server.ShutdownTimeout`, and exits.
