# OpenGateway

[![codecov](https://codecov.io/gh/lucdoe/opengateway/branch/main/graph/badge.svg?token=SDFO3CX9ZN)](https://codecov.io/gh/lucdoe/opengateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucdoe/opengateway)](https://goreportcard.com/report/github.com/lucdoe/opengateway)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Build](https://github.com/lucdoe/opengateway/actions/workflows/ci.yml/badge.svg)

A lightweight, open-source API gateway written in Go for small to mid-sized projects. You describe your services and the plugins they need in one YAML file; the gateway routes requests to them. No vendor lock-in, no control plane.

See the [architecture overview](docs/README.md) for how it is built.

## Quick start

```bash
git clone https://github.com/lucdoe/opengateway.git
cd opengateway
```

Over SSH, clone `git@github.com:lucdoe/opengateway.git` instead.

Edit [`cmd/opengateway/config.yaml`](cmd/opengateway/config.yaml) to describe your services, then start the gateway on port 4000 in one of two ways.

**Docker** (needs Docker with Compose, for example [Docker Desktop](https://www.docker.com/products/docker-desktop/)). Starts the gateway and Redis, and sets `STORE_TYPE=redis`:

```bash
docker compose up --build
```

The config file is copied into the image, so rebuild after changing it.

**Go** (needs Go 1.22 or newer; nothing else unless you choose the `redis` store):

```bash
go run ./cmd/opengateway
```

## Configuration

Everything is configured in `config.yaml`. The gateway reads it from the path given with `-c`, else from `CONFIG_PATH`, else from `./cmd/opengateway/config.yaml`. The file is validated at startup: an unknown key or an invalid value stops the gateway with a list of every problem.

To validate a config without starting the gateway, for example in CI, run `check`. It lists every problem and exits with `1`, or prints a summary and exits with `0`. It contacts neither Redis nor the upstreams.

```bash
go run ./cmd/opengateway check -c path/to/config.yaml
```

Both `check` and startup print warnings for settings that are valid but risky. They don't stop the gateway or fail `check`. All of them only apply when auth is used:

| Warning | Why |
| --- | --- |
| `SecretKey` is the example secret, or shorter than 32 bytes | Anyone can forge tokens with a known secret. HS256 needs 256 bits to resist guessing. |
| `Issuer` or `Audience` is empty | Tokens from other issuers, or meant for other services that share the secret, are accepted. |
| A service with auth endpoints has a plain `http://` URL to an outside host | The token is forwarded unencrypted. Loopback and private addresses, names without a dot (like a Compose service) and `.local`, `.internal`, `.lan`, `.svc` and `.localhost` names count as inside. |

### Services

Each service has an upstream `URL`, a list of `Endpoints`, and optionally a `Subpath` (starting with `/`, not ending with one) and `Plugins`.

```yaml
Services:
  starwars:
    URL: "https://swapi.dev"
    Subpath: "/api"
    Plugins: ["logger", "rate-limit"] # for every endpoint of this service
    Endpoints:
      - Name: "Get Person"
        HTTPMethod: "GET"
        Path: "/people/{id}"
        Plugins: ["cache"] # in addition, for this endpoint only
        Auth:
          ApplyAuth: true # require a valid JWT
          Scope: "people:read" # optional: scopes the token must carry, space-separated
```

With this config, `GET /api/people/7` on the gateway is forwarded to `https://swapi.dev/api/people/7`.

### Plugins

Plugins run in the order of this table, whatever order you list them in. Their settings live under `Plugins` in the config file.

| Plugin       | What it does                                                                                                        | Settings          |
| ------------ | ------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `logger`     | Writes one JSON line per answered request: method, path, route, status, duration, bytes, client IP and request ID (no query string). To stdout, or appended to `FilePath`. Lines are buffered and written every second; if the output can't keep up, lines are dropped and counted rather than slowing requests. | `LoggerConfig`    |
| `cors`       | Answers preflights and sets the CORS headers; rejects origins that are not listed with `403`. Requests without an `Origin` header (curl, other servers) pass untouched. `Origins: "*"` allows every origin but without credentials (cookies); list origins to allow them. | `CORSConfig`      |
| `rate-limit` | Allows each client IP (IPv6: each /64) `Limit` requests per `Window`, then answers `429` with `Retry-After`. Sends `X-RateLimit-Limit`, `-Remaining` and `-Reset` (seconds). If the store is unavailable, requests pass without a limit. | `RateLimitConfig` |
| `auth`       | Requires a `Bearer` JWT signed with HS256 and carrying `exp` (`401` otherwise). Checks `Issuer` and `Audience` when set. Tokens need every scope in `Scope`, and in the endpoint's `Auth.Scope`, in their space-separated `scope` claim (`403` otherwise). Enabled per endpoint with `Auth.ApplyAuth: true`. | `JWTConfig`       |
| `cache`      | Stores `200` responses to `GET` requests in the store and shares them between clients, for the upstream's `s-maxage` or `max-age`, at most 10 minutes. Requests with an `Authorization` header are only cached when the upstream marks the response `public` or `s-maxage`; it then goes to every caller that auth lets through. Requests with a `Cookie` header, streamed responses and bodies over 1 MB are never cached. | none              |

### Store

The `cache` and `rate-limit` plugins keep their state in a store. Pick one with `Store.Type`:

| `Store.Type`       | Where state lives          | Use it when                                                                 |
| ------------------ | -------------------------- | --------------------------------------------------------------------------- |
| `memory` (default) | in the gateway process     | you run one instance. Nothing to set up; state is lost on restart.          |
| `redis`            | a Redis server, 7.0 or newer | you run several instances that should share limits and cache. Valkey, KeyDB and Dragonfly work too. |

With `memory`, every instance counts its own limits: three instances with `Limit: 100` let a client through 300 times per window. The memory store holds at most 100,000 cached responses and 100,000 rate-limit counters; once either is full, new responses are not cached, or new clients pass without a limit, until old keys expire.

### Environment variables in the config

Any value can come from the environment, so secrets and per-environment URLs stay out of the file:

```yaml
Services:
  users:
    URL: "${USERS_URL:-http://localhost:3001}" # the fallback after :- is used when USERS_URL is unset or empty
Plugins:
  JWTConfig:
    SecretKey: "${JWT_SECRET}" # the gateway does not start while JWT_SECRET is unset
```

- A variable that is unset or empty, and has no fallback, stops the gateway with its line. `check` reports it too.
- The value is used exactly as it is, never read as YAML, so it can't change the structure of the config.
- Unquoted, a placeholder takes the type of its value: `Port: ${PORT}` is a number. Inside `{ }` flow style, quote it, since its `}` would close the mapping.
- Placeholders work in values, not in keys or comments. Write `$${` for a literal `${`.

### Server and environment

These settings are optional. An environment variable, where one exists, overrides the config file. Durations are written like `500ms`, `10s` or `2m`.

| Config key                    | Environment variable | Default     | Meaning                                                         |
| ----------------------------- | -------------------- | ----------- | --------------------------------------------------------------- |
| `Server.Port`                 | `PORT`               | `4000`      | Port the gateway listens on                                     |
| `Server.ReadHeaderTimeout`    |                      | `10s`       | Time a client has to send its request headers                   |
| `Server.IdleTimeout`          |                      | `120s`      | How long an idle keep-alive connection stays open               |
| `Server.ShutdownTimeout`      |                      | `15s`       | Time requests in flight get to finish when the gateway stops    |
| `Server.UpstreamTimeout`      |                      | `30s`       | Time an upstream has to start answering before the gateway responds `504` |
| `Server.CircuitBreaker.ConsecutiveFailures` |        | `5`         | Failures in a row that open a service's circuit breaker         |
| `Server.CircuitBreaker.OpenTimeout`         |        | `30s`       | How long it stays open before trial requests go through again   |
| `Server.CircuitBreaker.HalfOpenRequests`    |        | `1`         | Trial requests allowed at once while it recovers                |
| `Store.Type`                  | `STORE_TYPE`         | `memory`    | Store of the `cache` and `rate-limit` plugins: `memory` or `redis` |
| `Store.Redis.Host`            | `REDIS_HOST`         | `localhost` | Redis server, only used when `Store.Type` is `redis`            |
| `Store.Redis.Port`            | `REDIS_PORT`         | `6379`      |                                                                 |
| `Store.Redis.Password`        | `REDIS_PASSWORD`     | empty       |                                                                 |
| `Plugins.JWTConfig.SecretKey` | `JWT_SECRET`         | none        | Secret for validating tokens, only needed when auth is used     |

Every service has its own circuit breaker. A failure is an upstream that cannot be reached, times out, or answers `500`, `502`, `503` or `504`. While the breaker is open the gateway answers `503` with a `Retry-After` header straight away, without contacting the service.

When the gateway answers a request itself, rather than passing on the upstream's response, the body is JSON with a `message`, for example `{"message":"Rate limit exceeded"}`. This covers unknown routes (`404`, `405`), plugin rejections (`401`, `403`, `429`) and upstream failures (`502`, `503`, `504`).

`GET /__health` answers `200 {"status":"ok"}` while the gateway runs, without plugins or contacting a service, for load balancers and container health checks. Paths starting with `/__` belong to the gateway, so services can't route them.

Every request carries an `X-Request-ID`: the one the client sent, if it is 1 to 128 visible ASCII characters, or else a new random one. The gateway sends it to the service, writes it in the log line and returns it on the response, also on errors it answers itself.

On `SIGINT` or `SIGTERM` the gateway stops accepting new connections, lets requests in flight finish, and exits.

## Performance

One gateway capped at **2 CPU cores**, with every plugin on (logger, CORS, rate limit, JWT auth, cache):

| | |
| --- | --- |
| **Throughput** | **16,000 requests/s** sustained, with no errors added by the gateway |
| **Added latency** | **+0.3 ms** median over calling the service directly, at 500 requests/s |
| **Cache hits** | answered in **0.13 ms** median at 16,000 requests/s |
| **Memory** | 16 MB at 500 requests/s, about 100 MB at 16,000 |
| **A service goes down** | its requests get `503` in 0.3 ms, other services keep their latency |

Measured with [k6](https://k6.io/) at a constant arrival rate (open loop) against mock services with realistic latency (median 8–40 ms) and payloads (0.1–4 KB). The traffic mix is 50% reads by ID, 20% cacheable lists, 20% a second service and 10% authenticated `POST`s, on an Apple M1 Pro.

## License

Apache 2.0, see [LICENSE](LICENSE) or this [short summary](<https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)>). Licenses of the dependencies are in [LICENSES](LICENSES).
