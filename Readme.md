
spark-history-web-proxy acts as a reverse proxy for [Spark History](https://spark.apache.org/docs/latest/monitoring.html) and [Spark UI](https://spark.apache.org/docs/latest/web-ui.html), completing existing Spark History by seamlessly integrating live (running) Spark applications. The web proxy enables real-time dynamic discovery and monitoring of running applications without delay.

The proxy is non-intrusive and independent from any Spark History or Spark version. It supports all Spark application deployment modes, including Kubernetes Jobs, Spark Operator, Jupyter Spark notebooks, and more.

# Installation

Refer to [README](helm/spark-history-web-proxy/README.md) for the customization and installation guide.

## Requirements

- Kubernetes cluster
- [Spark History Server](https://spark.apache.org/docs/latest/monitoring.html)
- [Helm](https://helm.sh/) installed

> [!NOTE]
> You can use the following [Spark History Server](https://github.com/OKDP/spark-history-server) helm chart.
> 