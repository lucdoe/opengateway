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

Everything is configured in `config.yaml`. The gateway reads it from the path in `CONFIG_PATH`, or from `./cmd/opengateway/config.yaml` if that is not set. The file is validated at startup: an unknown key or an invalid value stops the gateway with a list of every problem.

### Services

Each service has an upstream `URL`, a list of `Endpoints`, and optionally a `Subpath` and `Plugins`.

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
```

With this config, `GET /api/people/7` on the gateway is forwarded to `https://swapi.dev/api/people/7`.

### Plugins

Plugins run in the order of this table, whatever order you list them in. Their settings live under `Plugins` in the config file.

| Plugin       | What it does                                                                                                        | Settings          |
| ------------ | ------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `logger`     | Writes one JSON line per answered request: method, path, route, status, duration, bytes and client IP (no query string). To stdout, or appended to `FilePath`. Lines are buffered and written every second; if the output can't keep up, lines are dropped and counted rather than slowing requests. | `LoggerConfig`    |
| `cors`       | Rejects requests from origins that are not allowed and sets the CORS response headers.                              | `CORSConfig`      |
| `rate-limit` | Allows each client IP `Limit` requests per `Window`, then answers `429` with `Retry-After`. Sends `X-RateLimit-Limit`, `-Remaining` and `-Reset` (seconds). If the store is unavailable, requests pass without a limit. | `RateLimitConfig` |
| `auth`       | Requires a `Bearer` JWT signed with HS256. Enabled per endpoint with `Auth.ApplyAuth: true`.                        | `JWTConfig`       |
| `cache`      | Stores `200` responses to `GET` requests in the store for 10 minutes and shares them between clients. Requests with an `Authorization` header are never cached. | none              |

### Store

The `cache` and `rate-limit` plugins keep their state in a store. Pick one with `Store.Type`:

| `Store.Type`       | Where state lives          | Use it when                                                                 |
| ------------------ | -------------------------- | --------------------------------------------------------------------------- |
| `memory` (default) | in the gateway process     | you run one instance. Nothing to set up; state is lost on restart.          |
| `redis`            | a Redis server, 7.0 or newer | you run several instances that should share limits and cache. Valkey, KeyDB and Dragonfly work too. |

With `memory`, every instance counts its own limits: three instances with `Limit: 100` let a client through 300 times per window. The memory store holds at most 100,000 keys; once full, new responses are not cached until old ones expire.

### Server and environment

These settings are optional. An environment variable, where one exists, overrides the config file, so secrets can stay out of it. Durations are written like `500ms`, `10s` or `2m`.

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
