# Configuration

Everything is configured in one YAML file. The gateway reads it from the path given with `-c`, else from `CONFIG_PATH`, else from `./cmd/opengateway/config.yaml`. It's validated at startup: an unknown key or an invalid value stops the gateway with a list of every problem, and a key that has been renamed or removed says what to write instead.

`check` validates a config without starting the gateway. It lists every problem and exits with `1`, or prints a summary and exits with `0`. It contacts neither Redis nor the services, and prints the same [warnings](auth.md#warnings) as startup.

```bash
go run ./cmd/opengateway check -c path/to/config.yaml
```

## Services

Each service has an upstream `URL`, a list of `Endpoints`, and optionally a `Subpath` (starting with `/`, not ending with one), `Plugins` and [`Settings`](#settings-per-service-and-endpoint).

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
        Plugins: ["cache", "auth"] # in addition, for this endpoint only; auth requires a valid JWT
        Scope: "people:read" # optional, with auth: scopes the token must carry, space-separated
```

`GET /api/people/7` on the gateway is forwarded to `https://swapi.dev/api/people/7`. Paths starting with `/__` belong to the gateway, so services can't route them.

## Prefix routes and upstream paths

You don't have to list every endpoint:

- A `Path` ending in `/*` matches that path and every path below it: `/files/*` matches `/files`, `/files/a` and `/files/a/b`.
- `HTTPMethod: "*"` matches every method.

When routes overlap, the most specific one wins, whatever their order in the file: exact paths before prefixes, longer prefixes before shorter ones, a named method before `*`.

`UpstreamPath` changes the path sent to the service, after the service URL's own path:

```yaml
Services:
  accounts:
    URL: "https://accounts.internal"
    Subpath: "/api"
    Endpoints:
      - HTTPMethod: "GET"
        Path: "/users/{id}"
        UpstreamPath: "/v2/accounts/{id}" # GET /api/users/7 → /v2/accounts/7
      - HTTPMethod: "*"
        Path: "/files/*"
        UpstreamPath: "/storage/*" # GET /api/files/a/b.txt → /storage/a/b.txt
      - HTTPMethod: "*"
        Path: "/*" # everything else under /api, forwarded as it is
```

`{name}` takes the value of the same variable in `Path`, and a final `/*` takes what came after the prefix. Without `UpstreamPath`, the full gateway path is forwarded.

## Settings per service and endpoint

A service or an endpoint can change the rate limit, the CORS policy and, for a service, the upstream timeout, in a `Settings` block. Only the fields it sets change. An endpoint's settings go over its service's, which go over `Plugins` and `Server`:

```yaml
Plugins:
  RateLimit: {Limit: 1000, Window: 1m, Key: "jwt:sub"} # the default
Services:
  accounts:
    URL: "https://accounts.internal"
    Plugins: ["rate-limit"]
    Settings:
      UpstreamTimeout: 5s
      RateLimit: {Limit: 100} # 100 a minute per user, for every endpoint of the service
    Endpoints:
      - Name: "Login"
        HTTPMethod: "POST"
        Path: "/login"
        Settings:
          RateLimit: {Limit: 5, Key: "ip"} # 5 a minute per IP, for this endpoint only
```

Each `Settings` block that changes the rate limit has its own count: the 5 logins above don't use up the 100 of the other accounts endpoints, or the global 1000. Endpoints that inherit a service's limit share one count. `UpstreamTimeout` can only be set for a service.

## Environment variables

Any value can come from the environment, so secrets and per-environment URLs stay out of the file:

```yaml
Services:
  users:
    URL: "${USERS_URL:-http://localhost:3001}" # the fallback after :- is used when USERS_URL is unset or empty
Plugins:
  JWT:
    SecretKey: "${JWT_SECRET}" # the gateway does not start while JWT_SECRET is unset
```

- A variable that is unset or empty, and has no fallback, stops the gateway with its line. `check` reports it too.
- The value is used exactly as it is, never read as YAML, so it can't change the structure of the config.
- Unquoted, a placeholder takes the type of its value: `Port: ${PORT}` is a number. Inside `{ }` flow style, quote it, since its `}` would close the mapping.
- Placeholders work in values, not in keys or comments. Write `$${` for a literal `${`.

The variables in the [table below](#server-and-store) also override the file when set, without a placeholder.

## Server and store

These settings are optional. Durations are written like `500ms`, `10s` or `2m`.

| Config key | Environment variable | Default | Meaning |
| --- | --- | --- | --- |
| `Server.Port` | `PORT` | `4000` | Port the gateway listens on |
| `Server.MetricsPort` | | `0` (off) | Port that serves [Prometheus metrics](operations.md#metrics) at `/metrics` |
| `Server.ReadHeaderTimeout` | | `10s` | Time a client has to send its request headers |
| `Server.IdleTimeout` | | `120s` | How long an idle keep-alive connection stays open |
| `Server.ShutdownTimeout` | | `15s` | Time requests in flight get to finish when the gateway stops |
| `Server.UpstreamTimeout` | | `30s` | Time a service has to start answering before the gateway responds `504` |
| `Server.CircuitBreaker.ConsecutiveFailures` | | `5` | Failures in a row that open a service's [circuit breaker](operations.md#circuit-breaker) |
| `Server.CircuitBreaker.OpenTimeout` | | `30s` | How long it stays open before trial requests go through again |
| `Server.CircuitBreaker.HalfOpenRequests` | | `1` | Trial requests allowed at once while it recovers |
| `Store.Type` | `STORE_TYPE` | `memory` | Store of the `cache` and `rate-limit` plugins: `memory` or `redis` |
| `Store.Redis.Host` | `REDIS_HOST` | `localhost` | Redis server, only used when `Store.Type` is `redis` |
| `Store.Redis.Port` | `REDIS_PORT` | `6379` | |
| `Store.Redis.Password` | `REDIS_PASSWORD` | empty | |
| `Plugins.JWT.SecretKey` | `JWT_SECRET` | none | Secret for validating tokens, only needed when auth is used |

### Store

The `cache` and `rate-limit` plugins keep their state in the store:

| `Store.Type` | Where state lives | Use it when |
| --- | --- | --- |
| `memory` (default) | in the gateway process | you run one instance. Nothing to set up; state is lost on restart. |
| `redis` | a Redis server, 7.0 or newer | you run several instances that should share limits and cache. Valkey, KeyDB and Dragonfly work too. |

With `memory`, every instance counts its own limits: three instances with `Limit: 100` let a client through 300 times per window. The memory store holds at most 100,000 cached responses and 100,000 rate-limit counters; once either is full, new responses are not cached, or new clients pass without a limit, until old keys expire.
