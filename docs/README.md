# Architecture

OpenGateway is a single Go binary. It reads one YAML file, builds a route for every configured endpoint, and reverse-proxies matching requests to the service behind it. Redis is only needed by the `cache` and `rate-limit` plugins.

The first two diagrams follow the [C4 model](https://c4model.com/); their sources are in [diagrams-as-code](diagrams-as-code).

## Context

Who uses the gateway and what it talks to.

![C4 context diagram](diagrams-as-image/context.png)

## Containers

The main parts inside the gateway.

![C4 container diagram](diagrams-as-image/container.png)

## Packages

How the code is organised. Arrows point from a package to what it uses.

```mermaid
flowchart LR
    main["cmd/opengateway<br/>wires everything, handles signals"]
    config["internal/config<br/>load, environment, defaults, validation"]
    server["internal/server<br/>routes, plugin chains, HTTP server"]
    proxy["internal/proxy<br/>one reverse proxy per service"]
    plugins["internal/plugins<br/>logger, cors, rate-limit, auth, cache"]
    yaml[("config.yaml")]
    redis[("Redis")]
    upstream[("Services")]

    main --> config --> yaml
    main --> plugins
    main --> server
    main --> proxy
    server --> config
    plugins -.->|"cache, rate-limit"| redis
    proxy --> upstream
```

- **`cmd/opengateway`** loads the config, builds only the plugins that some endpoint uses, and starts the server.
- **`internal/config`** owns the config file format. Nothing else reads files or environment variables.
- **`internal/server`** registers one route per endpoint and wraps it in that endpoint's plugins.
- **`internal/plugins/<name>`** holds one plugin each: its logic and its HTTP middleware.
- **`internal/proxy`** forwards a request to a service and maps failures to `502` and `504`.

## Request flow

A request to an endpoint that uses every plugin. Each plugin can answer on its own and end the request early.

```mermaid
flowchart TB
    client(["Client"]) --> router{"route for<br/>method and path?"}
    router -->|no| notfound["404 or 405"]
    router -->|yes| logger["logger<br/>writes a log line"]
    logger --> cors{"cors<br/>origin allowed?"}
    cors -->|no| forbidden["403"]
    cors -->|yes| ratelimit{"rate-limit<br/>under the limit?"}
    ratelimit -->|no| toomany["429"]
    ratelimit -->|yes| auth{"auth<br/>valid JWT?"}
    auth -->|no| unauthorized["401"]
    auth -->|yes| cache{"cache<br/>stored response?"}
    cache -->|hit| cached["stored response"]
    cache -->|miss| proxy["proxy"]
    proxy --> upstream[("Service")]
    upstream -->|response| stored["response to client,<br/>stored if cacheable"]
    proxy -.->|"unreachable or too slow"| gatewayerror["502 or 504"]
```

Plugins always run in this order: `logger`, `cors`, `rate-limit`, `auth`, `cache`. Auth runs before the cache so a stored response is never served to a caller that has not been authenticated.
