# spark-history-web-proxy

Helm chart for spark-history-web-proxy

## Introduction

Helm chart for [Spark History Web Proxy](../../) deployment using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Helm >= 3
- Kubernetes >= 1.19

## Installing the chart

To install the chart with the release name `my-release`:

```shell
$ helm install my-release oci://quay.io/okdp/charts/spark-history-web-proxy --version 0.1.0
```

This will create a release of `my-release` in the default namespace. To install in a different namespace:

```shell
$ helm install my-release oci://quay.io/okdp/charts/spark-history-web-proxy --version 0.1.0 \
       --namespace spark
```

Note that `helm` will fail to install if the namespace if doesn't exist. Either create the namespace beforehand or pass the `--create-namespace` flag to the `helm install` command.

## Uninstalling the chart  `my-release`

To uninstall `my-release`:

```shell
$ helm uninstall my-release -n spark
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Downloading the chart locally

To download the chart locally, use the following command:

```shell
$ helm pull oci://quay.io/okdp/charts/spark-history-web-proxy --version 0.1.0
```

##### 2. Deploy the Helm Chart

```shell
helm install my-release oci://quay.io/okdp/charts/spark-history-web-proxy --version 0.1.0 \
      --namespace spark \
      --values my-values.yaml \
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod scheduling. |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| configuration.logging.format | string | `"console"` |  |
| configuration.logging.level | string | `"debug"` |  |
| configuration.proxy.listenAddress | string | `"0.0.0.0"` | Specify the Proxy listen address. |
| configuration.proxy.mode | string | `"debug"` | Specify the Server Mode. One of `debug`, `release` or `test`. |
| configuration.proxy.port | int | `4040` | Specify the Proxy listen port. |
| configuration.security.cors.allowCredentials | bool | `false` | Determine whether cookies and authentication credentials should be included in cross-origin requests. |
| configuration.security.cors.allowedHeaders | list | `["Origin","Accept","Authorization","Content-Length","Content-Type"]` | List the headers that clients are allowed to include in requests. |
| configuration.security.cors.allowedMethods | list | `["GET","POST","PUT","DELETE","PATCH","OPTIONS","HEAD"]` | Define the HTTP methods permitted for CORS requests. |
| configuration.security.cors.allowedOrigins | list | `["*"]` | Specify the allowed origins for cross-origin requests. "*" allows all origins. |
| configuration.security.cors.exposedHeaders | list | `["Content-Length"]` | Specify which response headers should be exposed to the client. |
| configuration.security.cors.maxAge | int | `3600` | Define how long (in seconds) the results of a preflight request can be cached by the client. |
| configuration.security.headers | object | `{}` |  |
| configuration.spark.history.port | int | `18080` | Same as spark.history.ui.port |
| configuration.spark.history.scheme | string | `"http"` | Specify the Spark History listen address scheme. |
| configuration.spark.history.service | string | `nil` | Specify the Spark History listen kubernetes service name. |
| configuration.spark.jobNamespaces | list | `["default"]` | List of namespaces where the spark jobs run. If empty, all namespaces will be allowed. |
| configuration.spark.ui.port | int | `4040` | Same as spark.ui.port |
| configuration.spark.ui.proxyBase | string | `"/sparkui"` | When the proxyBase is set to a value other than `/proxy`, disable the property `spark.ui.reverseProxy=false` in your Spark job configuration if already set. |
| fullnameOverride | string | `""` | Overrides the release name. |
| image.pullPolicy | string | `"Always"` | Image pull policy. |
| image.repository | string | `"quay.io/okdp/spark-history-web-proxy"` | Docker image registry. |
| image.tag | string | `"0.1.0-snapshot"` | Image tag. |
| imagePullSecrets | list | `[]` | Secrets to be used for pulling images from private Docker registries. |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` | Specify the ingress class (Kubernetes >= 1.18). |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| livenessProbe | object | `{"httpGet":{"path":"/healthz","port":"http"},"initialDelaySeconds":60,"periodSeconds":30,"timeoutSeconds":10}` | Liveness probe for the okdp-server container. |
| nameOverride | string | `""` | Override for the `okdp-server.fullname` template, maintains the release name. |
| nodeSelector | object | `{}` | Node selector for pod scheduling. |
| podAnnotations | object | `{}` | Additional annotations for the okdp-server pod. |
| podLabels | object | `{}` | Additional labels for the okdp-server pod. |
| podSecurityContext | object | `{}` |  |
| rbac.annotations | object | `{}` | Specify annotations for the proxy. |
| rbac.create | bool | `true` | Specify whether a RBAC should be created |
| readinessProbe | object | `{"httpGet":{"path":"/readiness","port":"http"}}` | Readiness probe for the okdp-server container. |
| replicaCount | int | `1` | Desired number of okdp-server pods to run. |
| resources | object | `{}` |  |
| securityContext | object | `{}` | Security context for the container. |
| service.port | int | `4040` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specify whether a service account should be created |
| serviceAccount.name | string | `""` | If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` | Tolerations for pod scheduling. |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |

## Source Code

* <https://github.com/apache/spark>
* <https://github.com/OKDP/okdp-spark-auth-filter>
* <https://github.com/OKDP/spark-images>
* <https://github.com/OKDP/spark-history-server/tree/main/helm/spark-history-server>
* <https://github.com/OKDP/spark-history-web-proxy>

