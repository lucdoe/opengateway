---
title: Benchmark
description: Throughput, added latency, cache hits and memory of one gateway on 2 CPU cores, with every plugin on.
---

One gateway capped at **2 CPU cores**, with every plugin on (logger, CORS, rate limit, JWT auth, cache):

| | |
| --- | --- |
| **Throughput** | **16,000 requests/s** sustained, with no errors added by the gateway |
| **Added latency** | **+0.3 ms** median over calling the service directly, at 500 requests/s |
| **Cache hits** | answered in **0.13 ms** median at 16,000 requests/s |
| **Memory** | 16 MB at 500 requests/s, about 100 MB at 16,000 |
| **A service goes down** | its requests get `503` in 0.3 ms, other services keep their latency |

## How it was measured

Measured with [k6](https://k6.io/) at a constant arrival rate (open loop) against mock services with realistic latency (median 8–40 ms) and payloads (0.1–4 KB). The traffic mix is 50% reads by ID, 20% cacheable lists, 20% a second service and 10% authenticated `POST`s, on an Apple M1 Pro.
