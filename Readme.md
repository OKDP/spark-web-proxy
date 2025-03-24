[![ci](https://github.com/okdp/spark-web-proxy/actions/workflows/ci.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/ci.yml)
[![release-please](https://github.com/okdp/spark-web-proxy/actions/workflows/release-please.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/release-please.yml)
[![image-rebuild](https://github.com/okdp/spark-web-proxy/actions/workflows/docker-rebuild.yml/badge.svg)](https://github.com/okdp/spark-web-proxy/actions/workflows/docker-rebuild.yml)
[![License Apache2](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)


spark-web-proxy acts as a reverse proxy for [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html) and [Spark UI](https://spark.apache.org/docs/latest/web-ui.html). It completes [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html) by seamlessly integrating live (running) Spark applications UIs. The web proxy enables real-time dynamic discovery and monitoring of running spark applications (without delay) alongside completed applications, all within your existing Spark History Server Web UI.

The proxy is non-intrusive and independent of any specific version of Spark History Server or Spark. It supports all Spark application deployment modes, including Kubernetes Jobs, Spark Operator, notebooks (Jupyter, etc), etc.

![Spark History](docs/images/spark-history.png)

## Requirements

- Kubernetes cluster
- [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html)
- [Helm](https://helm.sh/) installed

> [!NOTE]
> You can use the following [Spark History Server](https://github.com/OKDP/spark-history-server) helm chart.
> 

## Installation

To deploy the spark web proxy, refer to helm chart [README](helm/spark-web-proxy/README.md) for customization options and installation guidelines.

The web proxy can also be deployed as a sidecar container alongside your existing Spark History Server. Ensure to set the property `configuration.spark.service` to `localhost`.

In both cases, you need to use the spark web proxy ingress instead of your spark history ingress.

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

### Cluster mode

In a cluster mode, `no additional configuration` is needed as spark by default adds the label `spark-role: driver` and the `spark-ui` port in the spark driver pods as shown in the following:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    ...
    spark-role: driver
spec:
  containers:
  - args:
    - driver
    name: spark-kubernetes-driver
    ports:
    ...
    - containerPort: 4040
      name: spark-ui
      protocol: TCP
```

### Notebooks and Client mode

In a client mode, the web proxy relies on [/api/v1/applications/[app-id]/environment](https://spark.apache.org/docs/latest/monitoring.html) Spark History Rest API to get the Spark driver IP and UI port and [/api/v1/applications/[app-id]](https://spark.apache.org/docs/latest/monitoring.html) to get the application status.

By default, Spark does not render the property `spark.ui.port` in the environment properties. So, you should set the property during the job submission or using a listener.

Here is an example of how to set the `spark.ui.port` on a jupyter notebook:

```python
import socket
def find_available_port(start_port=4041, max_port=4100):
    """Find the next available port starting from start_port."""
    for port in range(start_port, max_port):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            if s.connect_ex(("localhost", port)) != 0:
                return port
    raise Exception(f"No available ports found in range {start_port}-{max_port}")
```

```python
conf.set("spark.ui.port", find_available_port())
```

## Authentication

[Spark Authentication Filter](https://github.com/OKDP/okdp-spark-auth-filter) can be applied to both Spark History Server and Spark Jobs to enable user authentication and authorization.
