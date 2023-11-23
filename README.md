<a href="https://last9.io"><img src="https://last9.github.io/assets/last9-github-badge.svg" align="right" /></a>

# metric-explorer
With the need to accommodate more granular information in metrics, the problem of cardinality is becoming more and more natural. We talk about cardinality often, but we don’t have the right ways to explore our time-series database to find helpful information to control or improve this problem.

Introducing **metric-explorer.** It provides three modes of operation:

1. System(system)
2. Explore(explore)
3. Cardinality Calculator(cc)

**metric-explorer** is compatible with:
- Victoriametrics
- Prometheus (WIP)

**Build the tool from source**
```
git clone git@github.com:pree-dew/metric-explorer.git
make my_binary
```


```
./bin/metric-explorer --help
A tool that helps in answering: I have detected high cardinality, what to do next?.

Provides capability to take decisions on how to control cardinality.
Supports three modes:
1. System(system): To get system wide information about cardinality.
2. Explore (explore): To know more indetails about specific metric.
3. Cardinality Control(cc): To make decision to control cardinality

Usage:
  metric-explorer [command]

Available Commands:
  cc          To understand the cardinality distribution of a metric
  completion  Generate the autocompletion script for the specified shell
  explore     Provide metrics information
  help        Help about any command
  system      Overview of your TSDB coverage

Flags:
      --config string   config file (default is $HOME/.metric_explorer.yaml)
  -h, --help            help for metric_explorer
  -v, --version         version for metric_explorer

Use "metric-explorer [command] --help" for more information about a command.
```
### System Mode:

Everything begins by identifying the spread first. When we have a database with many metrics, it’s essential to identify the most influential metrics in the system and their usage limits.
```
./bin/metric-explorer system --help
```
**Use Case 1**: Find topN cardinality metrics
```
./bin/metric-explorer system --config example/sample.yaml --cardinality
```
**Use Case 2**: Find my top N queries in decreasing order of average running time, over past x seconds
```
./bin/metric-explorer system --config example/sample.yaml --top-queries --topN=3 --top-query-max-lifetime=300 --dump-as=table
```
### Explore Mode:

After identifying the troublesome metric and queries, it’s important to understand other details about specific metrics in detail to narrow down the different kinds of problems whether it is cardinality, resource crunch, resets counts, loss of signal, sparseness, etc
```
./bin/metric-explorer explore --help
```
**Use Case 1**: Find the cardinality distribution of a specific metric, along with its label values in decreasing order of cardinality contribution
```
./bin/metric-explorer explore http_request_total --config example/sample.yaml 
--cardinality --label-count=3 --dump-as=table
```
**Use Case 2**: Find the last loss of signal for a metric (Supports both counter and gauge)
```
./bin/metric-explorer explore http_request_total --config example/sample.yaml --loss=360000 --lag=1800
```
**Use Case 3**: Find the sparseness % of a metric. Note: Sparseness here means metric is absent.
```
./bin/metric-explorer explore http_request_total --config example/sample.yaml --sparse=360000 --lag=1800
```
**Use Case 4**: Find no. of resets happend on a counter metric in past x seconds
```
./bin/metric-explorer explore http_request_total --config example/sample.yaml --reset-counts --lag=1800
```
**Use Case 5**: Find churn rate, ingestion rate, sample received, active time-series, scrape interval of a metric
```
./bin/metric-explorer explore http_request_total --config example/sample.yaml 
--active-timeseries --scrape-interval --ingestion-rate --sample-received --churn-rate  --lag=1800
```
### Cardinality Calculator Mode:

After finding that cardinality is the problem, we have to find/investigate which labels are the culprit and how to go about them be dropping a few of them to control the problem. It’s not easy to find this information for a very high cardinality metric, and mainly, the way cardinality has been considered so far as cartesian products of count of all unique labels is not the right way to think about it.
```
./bin/metric-explorer cc --help
```
**Use Case 1**: Get the cardinality contribution of each label

*Note*: Beyond cardinality limit **--allowed-cardinality-limit** tool automatically creates relative query; we can also control relative behavior by using flag `--filter-label`
```
./bin/metric-explorer cc http_request_total --config example/sample.yaml --allowed-cardinality-limit=30000
```
**Use Case 2**: Define the label on which you want to check cardinality
```
./bin/metric-explorer cc http_request_total --config example/sample.yaml --filter-label=cluster --dump-as=table
```
**Use Case 3**: Judge the cardinality % in pairs to identify the relation between labels is 1:1, 1:M or M:N. It is very critical to understand because if it is 1:1 then it may not result in bringing the cardinality so much.
```
./bin/metric-explorer cc http_request_total --config example/sample.yaml --filter-label=cluster --label-count=2 --dump-as=table
```

---
