package mode

import (
	"fmt"
	"os"
	"sync"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"

	apiclient "metric_explorer/api_client"
)

type MetricFlag struct {
	Metric           string
	DumpAs           string
	Cardinality      string
	ChurnRate        int
	ScrapeInterval   bool
	RespTime         int
	Loss             int
	SampleReceived   int
	IRate            int
	ResetTime        int
	Lag              int
	SparseDuration   int
	ActiveTimeSeries int
	LabelCount       string
}

type metricInfo struct {
	cardinality      int64
	scrapeInterval   int
	labelInfo        labelMap
	labelValues      map[string][]string
	respTime         float32
	loss             int
	sampleReceived   int64
	iRate            float64
	churnRate        float64
	resetTime        int
	activeTimeSeries int64
	isSparse         bool
}

func MInfoInvoke(dataSource string, m MetricFlag) {
	var (
		wg    = &sync.WaitGroup{}
		lock  = sync.RWMutex{}
		mInfo = metricInfo{labelInfo: labelMap{}, labelValues: map[string][]string{}}
	)

	client, err := api.NewClient(api.Config{
		Address: dataSource,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)

	// if cardinality information is asked then get cardinality with
	// label information
	if m.Cardinality != "" {
		r, err := apiclient.MetricInfo(v1api, m.Metric, focusLabel, topN, m.Cardinality)
		if err != nil {
			fmt.Println("Error while fetching cardinality info: ", err)
			return
		}

		if len(r.SeriesCountByMetricName) == 0 {
			fmt.Println("No series found")
			return
		}

		mInfo.cardinality = int64(r.SeriesCountByMetricName[0].Value)
		for l := range r.LabelValueCountByLabelName {
			label := r.LabelValueCountByLabelName[l].Name
			mInfo.labelInfo[label] = labelInfo{uniqueCount: int(r.LabelValueCountByLabelName[l].Value)}

			wg.Add(1)
			go func() {
				defer wg.Done()

				r, err := apiclient.MetricInfo(v1api, m.Metric, label, m.LabelCount, m.Cardinality)
				if err != nil {
					fmt.Println("Error while fetching focus label value: ", err)
					return
				}

				labelValues := []string{}
				for i := range r.SeriesCountByFocusLabelValue {
					labelValues = append(labelValues, r.SeriesCountByFocusLabelValue[i].Name)
				}

				lock.Lock()
				mInfo.labelValues[label] = labelValues
				lock.Unlock()
			}()
		}
	}

	if m.ScrapeInterval {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.ScrapeInterval(v1api, m.Metric)
			if err != nil {
				fmt.Println("Error while finding scrape interval: ", err)
			} else {
				mInfo.scrapeInterval = r
				fmt.Println("Scrape Interval:", mInfo.scrapeInterval)
			}
		}()
	}

	if m.ChurnRate != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.ChurnRate(v1api, m.Metric, m.ChurnRate, m.Lag)
			if err != nil {
				fmt.Println("Error while finding response time: ", err)
			} else {
				mInfo.churnRate = r
				fmt.Printf("Churn Rate [%ds]: %f\n", m.ChurnRate, mInfo.churnRate)
			}
		}()
	}

	if m.RespTime != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.ResponseTime(v1api, m.Metric, m.RespTime, m.Lag)
			if err != nil {
				fmt.Println("Error while finding response time: ", err)
			} else {
				mInfo.respTime = r
				fmt.Printf("Response Time [%ds]: %f\n", m.ResetTime, mInfo.respTime)

			}
		}()
	}

	if m.SparseDuration != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.MetricSparse(v1api, m.Metric, m.SparseDuration, m.Lag)
			if err != nil {
				fmt.Println("Error while finding response time: ", err)
			} else {
				mInfo.isSparse = false
				perGap := ((m.SparseDuration - r) * 100 / m.SparseDuration)
				if perGap > 10 {
					mInfo.isSparse = true
				}

				fmt.Printf("Sparseness %% over a duration of [%d]s: %d\n", m.SparseDuration, perGap)

			}
		}()
	}

	if m.Loss != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.LastLoss(v1api, m.Metric, m.Loss, m.Lag)
			if err != nil {
				fmt.Println("Error while finding last loss time: ", err)
			} else {
				mInfo.loss = r
				fmt.Printf("Last Loss [%ds]: %d\n", m.Loss, mInfo.loss)
			}
		}()
	}

	if m.SampleReceived != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.SampleReceived(v1api, m.Metric, m.SampleReceived, m.Lag)
			if err != nil {
				fmt.Println("Error while finding sample received: ", err)
			} else {
				mInfo.sampleReceived = r
				fmt.Printf("Sample Received [%ds]: %d\n", m.SampleReceived, mInfo.sampleReceived)
			}
		}()
	}

	if m.ActiveTimeSeries != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.ActiveTimeSeries(v1api, m.Metric, m.ActiveTimeSeries, m.Lag)
			if err != nil {
				fmt.Println("Error while finding sample received: ", err)
			} else {
				mInfo.activeTimeSeries = r
				fmt.Printf("Active Timeseries Received [%ds]: %d\n", m.ActiveTimeSeries, mInfo.activeTimeSeries)
			}
		}()
	}

	if m.IRate != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.IngestionRate(v1api, m.Metric, m.IRate, m.Lag)
			if err != nil {
				fmt.Println("Error while finding ingestion rate: ", err)
			} else {
				mInfo.iRate = r
				fmt.Printf("Ingestion Rate [%ds]: %f\n", m.IRate, mInfo.iRate)
			}
		}()
	}

	if m.ResetTime != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := apiclient.ResetTime(v1api, m.Metric, m.ResetTime, m.Lag)
			if err != nil {
				fmt.Println("Error while finding reset time: ", err)
			} else {
				mInfo.resetTime = r
				fmt.Printf("Resets Count for last [%ds]: %d\n", m.ResetTime, mInfo.resetTime)
			}
		}()
	}

	wg.Wait()

	if m.Cardinality != "" {
		dumpCardinalityInfoWithLabels(m.Metric, mInfo.cardinality, mInfo.labelInfo, mInfo.labelValues, m.DumpAs)
	}
}
