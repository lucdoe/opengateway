# OpenGateway

[![codecov](https://codecov.io/gh/lucdoe/opengateway/branch/main/graph/badge.svg?token=SDFO3CX9ZN)](https://codecov.io/gh/lucdoe/opengateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucdoe/opengateway)](https://goreportcard.com/report/github.com/lucdoe/opengateway)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Build](https://github.com/lucdoe/opengateway/actions/workflows/ci.yml/badge.svg)

A lightweight, open-source API gateway written in Go for small to mid-sized projects. You describe your services and the plugins they need in one YAML file; the gateway routes requests to them. No vendor lock-in, no control plane.

## Quick start

```bash
git clone https://github.com/lucdoe/opengateway.git
cd opengateway
```

Over SSH, clone `git@github.com:lucdoe/opengateway.git` instead.

Edit [`cmd/opengateway/config.yaml`](cmd/opengateway/config.yaml) to describe your services, then start the gateway on port 4000 in one of two ways.

**Docker** (needs Docker with Compose, for example [Docker Desktop](https://www.docker.com/products/docker-desktop/)). Starts the gateway and Redis, and sets `STORE_TYPE=redis`. The config file is copied into the image, so rebuild after changing it:

```bash
docker compose up --build
```

**Go** (needs Go 1.22 or newer; nothing else unless you choose the `redis` store):

```bash
go run ./cmd/opengateway
```

## A config

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
```

`GET /api/people/7` on the gateway is forwarded to `https://swapi.dev/api/people/7`. Check a config without starting the gateway, for example in CI:

```bash
go run ./cmd/opengateway check -c path/to/config.yaml
```

## Docs

| Page | What's in it |
| --- | --- |
| [Configuration](docs/configuration.md) | services, prefix routes and path rewrites, settings per service and endpoint, `${VAR}` placeholders, server and store settings |
| [Plugins](docs/plugins.md) | logger, CORS, rate limits per IP, header or user, the response cache |
| [Auth](docs/auth.md) | JWT with a shared secret or an identity provider's JWKS, scopes, warnings for risky settings |
| [Operations](docs/operations.md) | health check, request IDs, error responses, circuit breaker, Prometheus metrics, shutdown |
| [Architecture](docs/architecture.md) | how it's built, and how a request flows through it |

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
