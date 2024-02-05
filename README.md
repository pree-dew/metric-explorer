<a href="https://last9.io"><img src="https://last9.github.io/assets/last9-github-badge.svg" align="right" /></a>

# metric-explorer

With the need to accommodate more granular information in metrics, the problem of high cardinality is becoming more and more natural. We talk about cardinality often, but we don’t have the right ways to explore our time-series database to find helpful information to control or improve this problem.

> Read more on [what is high cardinality](https://last9.io/blog/what-is-high-cardinality/).

Introducing **metric-explorer.** It provides three modes of operation:

1. System(system)
2. Explore(explore)
3. Cardinality Calculator(cc)

**metric-explorer** is compatible with:
- Victoriametrics
- Prometheus (WIP)

**Build the tool from source**

```shell
git clone git@github.com:pree-dew/metric-explorer.git
make my_binary
```

```shell
./bin/metric-explorer --help
A tool that helps answer: I have detected high cardinality; what to do next?

It provides the capability to make decisions on how to control cardinality.
It supports three modes:
1. System(system): To get system-wide information about cardinality.
2. Explore (explore): To know more about specific metrics.
3. Cardinality Control(cc): To decide to control cardinality

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

```shell
./bin/metric-explorer system --help
```
**Use Case 1**: Find topN cardinality metrics

```shell
./bin/metric-explorer system --config example/sample.yaml --cardinality --dump-as=table
╭───────────────────────────────────────┬─────────────╮
│ METRIC                                │ CARDINALITY │
├───────────────────────────────────────┼─────────────┤
│ http_request_total                    │         111 │
│ scrape_duration_seconds               │           1 │
│ scrape_samples_post_metric_relabeling │           1 │
│ scrape_samples_scraped                │           1 │
│ scrape_series_added                   │           1 │
│ scrape_timeout_seconds                │           1 │
│ up                                    │           1 │
╰───────────────────────────────────────┴─────────────╯
```
**Use Case 2**: Find my top N queries in decreasing order of average running time over the past x seconds

```shell
./bin/metric-explorer system --config example/sample.yaml --top-queries --topN=3 --top-query-max-lifetime=300 --dump-as=table
```
### Explore Mode:

After identifying the troublesome metric and queries, it’s essential to understand other details about specific metrics to narrow down the different kinds of problems, whether cardinality, resource crunch, reset counts, loss of signal, sparseness, etc.

```shell
./bin/metric-explorer explore --help
```
**Use Case 1**: Find the cardinality distribution of a specific metric, along with its label values in decreasing order of cardinality contribution

```shell
./bin/metric-explorer explore http_request_total --config example/sample.yaml 
--cardinality --label-count=5 --dump-as=table
╭───────────────────┬────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ METRIC            │ HTTP_REQUEST_TOTAL │                                                                                                                                                       │
│ CARDINALITY       │                111 │                                                                                                                                                       │
│ LABEL             │       UNIQUE VALUE │ LABEL VALUES                                                                                                                                          │
├───────────────────┼────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ endpoint          │                  8 │ /api/v1/label/env/values,/api/v1/label/scrape_job/values,/api/v1/label/service_name/values,/api/v1/label/handler/values,/api/v1/label/instance/values │
│ status_code       │                  5 │ 422,500,400,200,503                                                                                                                                   │
│ method            │                  3 │ PUT,GET,POST                                                                                                                                          │
│ __name__          │                  1 │ http_request_total                                                                                                                                    │
│ instance          │                  1 │ localhost:9100                                                                                                                                        │
│ job               │                  1 │ vmagent-01                                                                                                                                            │
│ exported_instance │                  9 │ 10.16.130.145:9100,10.16.128.122:8482,10.16.131.243:9153,10.16.131.183:8482,10.16.128.129:9100                                                        │
│ host              │                  9 │ 10.16.130.145,10.16.128.122,10.16.131.243,10.16.131.183,10.16.128.129                                                                                 │
╰───────────────────┴────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```
**Use Case 2**: Find the last loss of signal for a metric (Supports both counter and gauge)

```shell
./bin/metric-explorer explore http_request_total --config example/sample.yaml --loss=360000 --lag=1800
```
**Use Case 3**: Find the sparseness % of a metric. Note: Sparseness here means metric is absent.

```shell
./bin/metric-explorer explore http_request_total --config example/sample.yaml --sparse=360000 --lag=1800
```
**Use Case 4**: Find no. of resets that happened on a counter metric in the past x seconds

```shell
./bin/metric-explorer explore http_request_total --config example/sample.yaml --reset-counts --lag=1800
```
**Use Case 5**: Find churn rate, ingestion rate, sample received, active time series, scrape interval of a metric

```shell
./bin/metric-explorer explore http_request_total --config example/sample.yaml 
--active-timeseries --scrape-interval --ingestion-rate --sample-received --churn-rate  --lag=1800
```

### Cardinality Calculator Mode:

After finding that cardinality is the problem, we have to find/investigate which labels are the culprit and how to go about them be dropping a few to control the problem. It’s not easy to find this information for a very high cardinality metric, and mainly, the way cardinality has been considered so far as cartesian products of count of all unique labels is not the right way to think about it.

```shell
./bin/metric-explorer cc --help
```

**Use Case 1**: Get the cardinality contribution of each label

*Note*: Beyond cardinality limit **--allowed-cardinality-limit** tool automatically creates relative query; we can also control relative behavior by using flag `--filter-label`

```shell
./bin/metric-explorer cc http_request_total --config example/sample.yaml --allowed-cardinality-limit=30000 --dump-as=table
╭───────────────────┬────────────────────┬───────────────╮
│ METRIC            │ HTTP_REQUEST_TOTAL │               │
│ CARDINALITY       │                111 │               │
│ LABEL             │       UNIQUE VALUE │ CARDINALITY % │
├───────────────────┼────────────────────┼───────────────┤
│ exported_instance │                  9 │            17 │
│ host              │                  9 │            17 │
│ endpoint          │                  8 │            37 │
│ status_code       │                  5 │            29 │
│ method            │                  3 │            25 │
│ job               │                  1 │            17 │
│ instance          │                  1 │            17 │
╰───────────────────┴────────────────────┴───────────────╯
```

**Use Case 2**: Define the label on which you want to check cardinality

```shell
 ./bin/metric-explorer cc http_request_total --config example/sample.yaml --filter-label=endpoint --allowed-cardinality-limit=100 --dump-as=table 
╭───────────────────┬──────────────────────────────────────────────────────────────────┬───────────────╮
│ METRIC            │ HTTP_REQUEST_TOTAL{ENDPOINT="/API/V1/LABEL/SERVICE_NAME/VALUES"} │               │
│ CARDINALITY       │                                                               16 │               │
│ LABEL             │                                                     UNIQUE VALUE │ CARDINALITY % │
├───────────────────┼──────────────────────────────────────────────────────────────────┼───────────────┤
│ host              │                                                                8 │            18 │
│ exported_instance │                                                                8 │            18 │
│ status_code       │                                                                5 │            25 │
│ method            │                                                                3 │            18 │
│ endpoint          │                                                                1 │            18 │
│ instance          │                                                                1 │            18 │
│ job               │                                                                1 │            18 │
╰───────────────────┴──────────────────────────────────────────────────────────────────┴───────────────╯
```

**Use Case 3**: Judge the cardinality % in pairs to identify the relation between labels is 1:1, 1:M or M:N. It is very critical to understand because if it is 1:1, it may not bring the cardinality so much.

```shell
./bin/metric-explorer cc http_request_total --config example/sample.yaml --filter-label=cluster --label-count=2 --dump-as=table
```

**Use Case 4**: Find out dropping a label or pair of labels is going to result into duplicates or not

```shell
./bin/metric-explorer cc dropmetric --config  example/sample.yaml 
```
To find for specific labels

```shell
./bin/metric-explorer cc dropmetric --config  example/sample.yaml  --labels=pod --labels=host --labels=instance
```

