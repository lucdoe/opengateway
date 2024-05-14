# OpenGateway

[![codecov](https://codecov.io/gh/lucdoe/open-gateway/branch/main/graph/badge.svg?token=SDFO3CX9ZN)](https://codecov.io/gh/lucdoe/open-gateway)

Welcome to our API Gateway, an efficient solution written in Go for small to mid-sized projects. It offers a streamlined approach to managing and securing your APIs.
Leveraging Go's performance and simplicity, this gateway ensures quick response times and reliable performance without the complexity of larger systems. It is open-source, avoiding vendor lock-in and allowing complete control over deployment and customization.

You can easily extend its functionality with configurable plugins and the setup is straightforward, with minimal configuration needed.

## Local setup

1. **Clone the repository:**
   <br> SSH: `git clone git@github.com:lucdoe/open-gateway.git`
   <br>HTTPS: `git clone https://github.com/lucdoe/open-gateway.git`

### Docker

Make sure you have Docker with Docker Compose installed on your machine. Docker recommends [Docker Desktop](https://www.docker.com/products/docker-desktop/) for Windows and Mac users.

2. **Edit the `cmd/gateway/config.yaml` file to your needs**
3. **Run Docker compose with `docker compose up`**

### Manual

2. **Edit the `cmd/gateway/config.yaml` file to your needs**
3. **Run the gateway with `go run cmd/gateway/main.go`**

## Go - Grade

| Metric      | Value        |
| ----------- | ------------ |
| **Grade**   | **A+ 98.1%** |
| Files       | 962          |
| Issues      | 158          |
| go_vet      | 100%         |
| gofmt       | 100%         |
| ineffassign | 100%         |
| gocyclo     | 83%          |
| license     | 100%         |
| misspell    | 100%         |

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

These results provide a clear picture of the API Gateway's performance under load, highlighting its ability to handle a high number of requests efficiently.

## License

The code of this repository is licensed under Apache v2.0. [Read License tldr](<https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)>) for a quick summary.
