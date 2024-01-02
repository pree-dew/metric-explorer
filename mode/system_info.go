package mode

import (
	"fmt"
	"math"
	"os"

	"github.com/pree-dew/metric-explorer/api_client/client_golang/api"
	v1 "github.com/pree-dew/metric-explorer/api_client/client_golang/api/prometheus/v1"

	apiclient "github.com/pree-dew/metric-explorer/api_client"
)

type SystemFlag struct {
	Lag              int
	TopN             string
	Cardinality      string
	DumpAs           string
	ChurnRate        int
	IngestionRate    int
	ActiveTimeSeries int
	TopQueries       bool
	TopNMaxLifeTime  string
}

type metricSeriesCount struct {
	name       string
	series     uint64
	percentage float64
}
type systemInfo struct {
	totalSeries      uint64
	topMetrics       []metricSeriesCount
	topQueries       []map[string]interface{}
	ingestionRate    int
	churnRate        float64
	activeTimeSeries int
	alertRules       int32
}

func SystemInvoke(dataSource string, sFlag SystemFlag) {
	// call tsdb api to get top 20 metrics
	// collect series count
	client, err := api.NewClient(api.Config{
		Address: dataSource,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)
	l := systemInfo{topMetrics: []metricSeriesCount{}}

	if sFlag.Cardinality != "" {
		if sFlag.Cardinality == "today" {
			sFlag.Cardinality = ""
		}
		result, err := apiclient.TopMetrics(v1api, sFlag.TopN, sFlag.Cardinality)
		if err != nil {
			fmt.Println("Error faced while fetching top metrics:", err)
			return
		}

		l.totalSeries = result.TotalSeries
		for m := range result.SeriesCountByMetricName {
			mSeries := result.SeriesCountByMetricName[m].Value
			percent := (float64(mSeries) * 100) / float64(l.totalSeries)
			l.topMetrics = append(l.topMetrics, metricSeriesCount{name: result.SeriesCountByMetricName[m].Name, series: mSeries, percentage: math.Round(percent*100) / 100})
		}

		dumpSystemView(l.topMetrics, sFlag.DumpAs)
	}

	if sFlag.ChurnRate != 0 {
		l.churnRate, err = apiclient.SystemChurnRate(v1api, sFlag.Lag)
		if err != nil {
			fmt.Println("Error faced while finding churn rate:", err)
		} else {
			fmt.Println("Churn Rate from last", sFlag.ChurnRate, "secs :", l.churnRate)
		}
	}

	if sFlag.IngestionRate != 0 {
		l.ingestionRate, err = apiclient.SystemIngestionRate(v1api, sFlag.Lag)
		if err != nil {
			fmt.Println("Error faced while finding ingestion rate:", err)
		} else {
			fmt.Println("Ingestion Rate from last 1 hour:", l.ingestionRate)
		}
	}

	if sFlag.ActiveTimeSeries != 0 {
		l.activeTimeSeries, err = apiclient.SystemActiveTimeSeries(v1api, sFlag.Lag)
		if err != nil {
			fmt.Println("Error faced while finding active timeseries:", err)
		} else {
			fmt.Println("Active timeseries from last 1 hour:", l.activeTimeSeries)
		}
	}

	if sFlag.TopQueries {
		res, err := apiclient.TopQueries(v1api, sFlag.TopN, sFlag.TopNMaxLifeTime)
		if err != nil {
			fmt.Println("Error faced while fetching top queries:", err)
			return
		}

		dumpTopQueriesView(res.TopByAverageDuration, sFlag.DumpAs)
	}
}
