[![ci](https://github.com/okdp/spark-web-proxy/actions/workflows/ci.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/ci.yml)
[![release-please](https://github.com/okdp/spark-web-proxy/actions/workflows/release-please.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/release-please.yml)
[![image-rebuild](https://github.com/okdp/spark-web-proxy/actions/workflows/docker-rebuild.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/docker-rebuild.yml)
[![License Apache2](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)


spark-web-proxy acts as a reverse proxy for [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html) and [Spark UI](https://spark.apache.org/docs/latest/web-ui.html). It completes [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html) by seamlessly integrating live (running) Spark applications UIs. The web proxy enables real-time dynamic discovery and monitoring of running spark applications (without delay) alongside completed applications, all within your existing Spark History Server Web UI.

The proxy is non-intrusive and independent of any specific version of Spark History Server or Spark. It supports all Spark application deployment modes, including Kubernetes Jobs, Spark Operator, Jupyter Spark notebooks, etc.

![Spark History](docs/images/spark-history.png)

## Requirements

- Kubernetes cluster
- [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html)
- [Helm](https://helm.sh/) installed

> [!NOTE]
> You can use the following [Spark History Server](https://github.com/OKDP/spark-history-server) helm chart.
> 

## Installation

The web proxy can be deployed either as a sidecar container alongside your Spark History Server or as an independent helm chart.

1. As a sidecar container:

Refer to [helm/spark-history-server](https://github.com/OKDP/spark-history-server/tree/main/helm/spark-history-server) repository for guidlines and examples.

2. As an independent chart:

Refer to [README](helm/spark-web-proxy/README.md) for customization options and installation guidelines.

In both cases, you need to use the web proxy ingress instead of your spark history ingress.

## Spark History and spark jobs Configuration

Both [Spark History and Spark jobs](https://spark.apache.org/docs/latest/monitoring.html) themselves must be configured to log events, and to log them to the same shared, writable directory.

### Spark History:

```console
spark.history.fs.logDirectory /path/to/the/same/shared/event/logs
```

### Spark Jobs:

```console
spark.eventLog.enabled true
spark.eventLog.dir /path/to/the/same/shared/event/logs
```

### Spark Reverse Proxy Support

The web proxy supports Spark Reverse Proxy feature for Spark web UIs by enabling the property `spark.ui.reverseProxy=true` in your spark jobs. In that case, the web proxy configuration property `configuration.spark.ui.proxyBase` should be set to `/proxy`

For more configuration properties, refer to [Spark Monitoring](https://spark.apache.org/docs/latest/monitoring.html) configuration page.

## Spark jobs deployment

In a cluster mode, spark by default adds the label `spark-role: driver` in the spark driver pods.

In a client mode, add the following label into your driver pods:

```yaml
kind: ...
metadata:
  labels:
    ...
    spark-role: driver
    ...
```

## Authentication

[Spark Authentication Filter](https://github.com/OKDP/okdp-spark-auth-filter) can be applied to both Spark History Server and Spark Jobs to enable user authentication and authorization.
