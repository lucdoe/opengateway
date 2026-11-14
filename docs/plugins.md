---
title: Plugins
description: Logging, CORS, rate limits per IP, header or user, and a shared response cache.
---

A service's `Plugins` run for each of its endpoints, an endpoint's in addition. They run in the order of this table, whatever order you list them in, and their settings live under `Plugins` in the config. A service or endpoint can change a plugin's settings for itself with [`Settings`](configuration.md#settings-per-service-and-endpoint).

| Plugin | What it does | Settings |
| --- | --- | --- |
| `logger` | Writes one JSON line per answered request: method, path, route, status, duration, bytes, client IP and request ID (no query string). To stdout, or appended to `FilePath`. Lines are buffered and written every second; if the output can't keep up, lines are dropped and counted rather than slowing requests. | `Logger` |
| `cors` | Answers preflights and sets the CORS headers; rejects origins that are not listed with `403`. Requests without an `Origin` header (curl, other servers) pass untouched. `Origins: "*"` allows every origin but without credentials (cookies); list origins to allow them. | `CORS` |
| `rate-limit` | Allows each client `Limit` requests per `Window`, then answers `429` with `Retry-After`. A client is an IP (IPv6: a /64) unless [`Key`](#rate-limit-keys) says otherwise. Sends `X-RateLimit-Limit`, `-Remaining` and `-Reset` (seconds). If the store is unavailable, requests pass without a limit. | `RateLimit` |
| `auth` | Requires a valid JWT, see [Auth](auth.md). | `JWT` |
| `cache` | Stores `200` responses to `GET` requests and shares them between clients, for the upstream's `s-maxage` or `max-age`, at most 10 minutes. Requests with an `Authorization` header are only cached when the upstream marks the response `public` or `s-maxage`; it then goes to every caller that auth lets through. Requests with a `Cookie` header, streamed responses and bodies over 1 MB are never cached. | none |

`cache` and `rate-limit` keep their state in the [store](configuration.md#store).

## Rate-limit keys

`Plugins.RateLimit.Key` sets what counts as one client:

| `Key` | Counts per | Use it when |
| --- | --- | --- |
| `ip` (default) | client IP, IPv6 per /64 | clients reach the gateway directly |
| `header:Name`, e.g. `header:CF-Connecting-IP` | value of that header | a CDN or load balancer in front sets the real client address. Only name a header it sets: clients can send any value. |
| `jwt:sub` | verified token subject | users share IPs, as behind a company NAT. Rate-limit then runs right after auth, so forged tokens get `401` and are never counted. |

A request without the header, or without a token, counts by its IP, so leaving it out doesn't escape the limit. Header values and subjects are stored hashed.
