# Mondiff

Application to show differences between Grafana and Prometheus objects. Its main purpose is to detect any difference between internal objects grouped by some kind of parent object, like comparing Rules from the same RuleGroup in Prometheus, for example.

## Packages used

- [go-cmp](https://github.com/google/go-cmp): For showing the diff between Go objects
- [grafana-sdk](https://github.com/grafana-tools/sdk): Interacting with Grafana's HTTP API
- [prometheus-client-go-api](https://github.com/prometheus/client_golang/tree/master/api): Interacting with Prometheus' HTTP API

## Usage

``` sh
./mondiff <profile1> <profile2>
```

Where `<profile#>` are [config](#config) profiles for Prometheus or Grafana.

## Config

The config file stores the profiles with connection information and is written in toml format. The file should be called `config.toml` and is automatically searched for in `~/.mondiff/config.toml` or the current directory where the app runs.

Here's an example:

``` toml
[grafana-1]
url = "https://grafana1.example.com"
auth = "api_token"

[grafana-2]
url = "https://grafana2.example.com"
auth = "user:password"

[prom-1]
url = "https://prometheus1.example.com"

[prom-2]
url = "https://prometheus2.example.com"
```

Grafana urls require auth info, either an API token or `user:password` basic auth info as shown in the example. Prometheus doesn't require any auth info by default.

The application will automatically discover what comparison to run depending on the url pattern "grafana" or "prometheus" respectively.

## Features

### Grafana

- Compare Dashboards that have the same Title in both instances.

### Prometheus

- Compare Rules, alert and recording, for RuleGroups with same Name in both instances.

## Example Report

Report output is almost the same for both Prometheus and Grafana

``` text
########################
# Dashboards only in A #
########################

Kubernetes / Compute Resources / Cluster
Kubernetes / Compute Resources / Namespace (Pods)
Kubernetes / Compute Resources / Namespace (Workloads)
Kubernetes / Compute Resources / Pod
Kubernetes / Compute Resources / Workload
Kubernetes / Nodes
Kubernetes / Persistent Volumes
Kubernetes / Pods
Kubernetes / StatefulSets
Kubernetes / USE Method / Cluster
Kubernetes / USE Method / Node
########################
# Dashboards only in B #
########################

Alertmanager capacity planning
Cluster Usage By Namespace
Prometheus
Prometheus Benchmark - 2.3.x
Prometheus Benchmark - 2.7.x
Prometheus Metrics
Prometheus Stats New
Prometheus exporters status
###################################
# Diff between dashboards in both #
###################################

# Prometheus capacity planning:
	{sdk.Board}.Time.From:
		-: now-15d
		+: now-30m
        
        .
        .
        .
# OPA:
    (*{sdk.Board}.Panels[2].CustomPanel)["targets"].([]interface {})[0].(map[string]interface {})["expr"].(string):
		-: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="opa"}[5m])) by (le))
		+: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="opa"}[5m])) by (le) )
    (*{sdk.Board}.Panels[6].CustomPanel)["targets"].([]interface {})[0].(map[string]interface {})["expr"].(string):
		-: sum(irate(http_request_duration_seconds_bucket{namespace="opa", handler="v1/policies"}[1m])) by (le)
		+: sum(irate(http_request_duration_seconds_bucket{namespace="opa", handler="v1/data"}[1m])) by (le)
    .
    .
    .
```
