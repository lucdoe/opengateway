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

**Docker** (needs Docker with Compose, for example [Docker Desktop](https://www.docker.com/products/docker-desktop/)). Starts the gateway and Redis:

```bash
docker compose up --build
```

The config file is copied into the image, so rebuild after changing it.

**Go** (needs Go 1.22 or newer, and Redis on `localhost:6379` if you use the `cache` or `rate-limit` plugin):

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
| `logger`     | Writes one line per request to a file.                                                                              | `LoggerConfig`    |
| `cors`       | Rejects requests from origins that are not allowed and sets the CORS response headers.                              | `CORSConfig`      |
| `rate-limit` | Allows each client IP `Limit` requests per `Window`, then answers `429`. Needs Redis.                               | `RateLimitConfig` |
| `auth`       | Requires a `Bearer` JWT signed with HS256. Enabled per endpoint with `Auth.ApplyAuth: true`.                        | `JWTConfig`       |
| `cache`      | Stores `200` responses to `GET` requests in Redis for 10 minutes and shares them between clients. Requests with an `Authorization` header are never cached. | none              |

### Server and environment

These settings are optional. An environment variable, where one exists, overrides the config file, so secrets can stay out of it. Durations are written like `500ms`, `10s` or `2m`.

| Config key                    | Environment variable | Default     | Meaning                                                         |
| ----------------------------- | -------------------- | ----------- | --------------------------------------------------------------- |
| `Server.Port`                 | `PORT`               | `4000`      | Port the gateway listens on                                     |
| `Server.ReadHeaderTimeout`    |                      | `10s`       | Time a client has to send its request headers                   |
| `Server.IdleTimeout`          |                      | `120s`      | How long an idle keep-alive connection stays open               |
| `Server.ShutdownTimeout`      |                      | `15s`       | Time requests in flight get to finish when the gateway stops    |
| `Server.UpstreamTimeout`      |                      | `30s`       | Time an upstream has to start answering before the gateway responds `504` |
| `Redis.Host`                  | `REDIS_HOST`         | `localhost` | Redis server, only used by the `cache` and `rate-limit` plugins |
| `Redis.Port`                  | `REDIS_PORT`         | `6379`      |                                                                 |
| `Redis.Password`              | `REDIS_PASSWORD`     | empty       |                                                                 |
| `Plugins.JWTConfig.SecretKey` | `JWT_SECRET`         | none        | Secret for validating tokens, only needed when auth is used     |

On `SIGINT` or `SIGTERM` the gateway stops accepting new connections, lets requests in flight finish, and exits.

## Performance

Benchmark with [`hey`](https://github.com/rakyll/hey), 50,000 requests at a concurrency of 130:

| Total time | Requests/s | Average | Slowest  | `200 OK` responses | Size per request | Total data       |
| ---------- | ---------- | ------- | -------- | ------------------ | ---------------- | ---------------- |
| 5.9502 s   | 8,389.5769 | 14.7 ms | 165.4 ms | 49,853             | 1,278 bytes      | 63,811,840 bytes |

Load test with [Locust](https://locust.io/), 200 peak users, ramp of 50, 5 minutes, `GET /some-endpoint`:

| Requests | Fails | Median | Average | Min  | Max   | Average size | Requests/s | Failures/s |
| -------- | ----- | ------ | ------- | ---- | ----- | ------------ | ---------- | ---------- |
| 15,532   | 0     | 4 ms   | 4.93 ms | 1 ms | 76 ms | 1,280 bytes  | 66.8       | 0          |

## License

Apache 2.0, see [LICENSE](LICENSE) or this [short summary](<https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)>). Licenses of the dependencies are in [LICENSES](LICENSES).
