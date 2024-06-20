# OpenGateway - lightweight API Gateway

[![codecov](https://codecov.io/gh/lucdoe/opengateway/branch/main/graph/badge.svg?token=SDFO3CX9ZN)](https://codecov.io/gh/lucdoe/opengateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucdoe/opengateway)](https://goreportcard.com/report/github.com/lucdoe/opengateway)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Build](https://github.com/lucdoe/opengateway/actions/workflows/ci.yml/badge.svg)

Welcome to the OpenGateway repo, an lightweight API Gateway written in Go for small to mid-sized projects. Leveraging Go's performance and simplicity, this gateway ensures quick response times and reliable performance without the complexity of larger systems. It is open-source, avoiding vendor lock-in and allowing complete control over deployment and customization through simple YAMl config.

You can easily extend its functionality with configurable plugins, and the setup is straightforward, with minimal configuration needed.

Please also see the [architecture diagrams](https://github.com/lucdoe/opengateway/tree/main/docs) for an overview and better understanding of the architecture.

## Local setup

1. **Clone the repository:**
   <br> SSH: `git clone git@github.com:lucdoe/opengateway.git`
   <br>HTTPS: `git clone https://github.com/lucdoe/opengateway.git`

### Docker

Make sure you have Docker with Docker Compose installed on your machine. Docker recommends [Docker Desktop](https://www.docker.com/products/docker-desktop/) for Windows and Mac users.

2. **Edit the `cmd/gateway/config.yaml` file to your needs**
3. **Run Docker compose with `docker compose up`**

### Manual

2. **Edit the `cmd/gateway/config.yaml` file to your needs**
3. **Run the gateway with `go run cmd/gateway/main.go`**

## Metrics

### Benchmark Results

We conducted a performance benchmark using `hey` with the following parameters:

- **Total Requests**: 50,000
- **Concurrency Level**: 130

#### Summary

- **Total Time**: 5.9502 seconds
- **Requests per Second**: 8,389.5769
- **Slowest Request**: 0.1654 seconds
- **Average Request Time**: 0.0147 seconds
- **Total Data Transferred**: 63,811,840 bytes
- **Size per Request**: 1,278 bytes

#### Status Code Distribution

- **200 OK**: 49,853 responses

### Load testing (Locust)

We have conducted load testing on our API Gateway using Locust (200 peaks, 50 ramps, 5m).

| Type | Name           | Requests | Fails | Median (ms) | Average (ms) | Min (ms) | Max (ms) | Average size (bytes) | Current RPS | Current Failures/s |
| ---- | -------------- | -------- | ----- | ----------- | ------------ | -------- | -------- | -------------------- | ----------- | ------------------ |
| GET  | /some-endpoint | 15,532   | 0     | 4           | 4.93         | 1        | 76       | 1,280                | 66.8        | 0                  |

The gateway performs well under the tested load, with no failures and consistent response times.

These results clearly show the API Gateway's performance under load, highlighting its ability to handle many requests efficiently.

## License

The code of this repository is licensed under Apache v2.0. [Read License tldr](<https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)>) for a quick summary.
