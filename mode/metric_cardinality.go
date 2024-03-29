package mode

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pree-dew/metric-explorer/api_client/client_golang/api"
	v1 "github.com/pree-dew/metric-explorer/api_client/client_golang/api/prometheus/v1"

	apiclient "github.com/pree-dew/metric-explorer/api_client"
)

type CardinalityFlag struct {
	Metric                     string
	DumpAs                     string
	Label                      []string
	LabelCount                 int
	CardinalityPerDuration     int
	AllowedCardinalityLimit    int64
	FilterLabel                string
	Lag                        int
	RelativeLabelNo            int
	DropAction                 bool
	AggregateAction            bool
	SplitAction                bool
	DisableRelativeCardinality bool
}

type cardinalityDetails struct {
	cardinality uint64
	labelInfo   labelMap
}

type RWMap struct {
	sync.RWMutex
	m labelsCardinalityInfo
}

// Get is a wrapper for getting the value from the underlying map
func (r *RWMap) Get(key string) labelInfo {
	r.RLock()
	defer r.RUnlock()
	return r.m[key]
}

// Set is a wrapper for setting the value of a key in the underlying map
func (r *RWMap) Set(key string, val labelInfo) {
	r.Lock()
	defer r.Unlock()
	r.m[key] = val
}

// SetDropActionInfo is a wrapper for setting the value of a key in the underlying map
func (r *RWMap) SetDropActionInfo(key string, labelExists bool) {
	r.Lock()
	defer r.Unlock()
	v, _ := r.m[key]
	v.duplicateExists = labelExists
	r.m[key] = v
}

func createPairs(labels []string, labelCount int) []string {
	pairs := []string{}
	if labelCount == 1 {
		return labels
	}

	for i := 0; i < len(labels)-1; i++ {
		for j := i + 1; j < len(labels); j++ {
			pairs = append(pairs, labels[i]+", "+labels[j])
		}
	}

	return pairs
}

func CardinalityInvoke(dataSource string, cFlag CardinalityFlag) {
	var (
		wg = &sync.WaitGroup{}
		cd = cardinalityDetails{labelInfo: map[string]labelInfo{}}
	)

	client, err := api.NewClient(api.Config{
		Address: dataSource,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)

	// In somecases where cardinality is high it is important to filter on some labels
	// api gives specific filter values as per the filter specified
	if cFlag.FilterLabel != "" {
		focusLabel = cFlag.FilterLabel
	}

	// Make status call with focus variable and specific metric
	r, err := apiclient.MetricInfo(v1api, cFlag.Metric, focusLabel, topN, "")
	if err != nil {
		fmt.Println("Error while fetching cardinality info: ", err)
		return
	}

	// In case if series is incorrect
	if len(r.SeriesCountByMetricName) == 0 {
		fmt.Println("No series found")
		return
	}

	cardinality := r.SeriesCountByMetricName[0].Value
	if cFlag.DisableRelativeCardinality && cardinality > uint64(cFlag.AllowedCardinalityLimit) {
		fmt.Println("Cardinality is greater than allowed limit and relative cardinality flag is disable, can't process")
		return
	}

	// In case of high cardinality pick the filter variable with smallest Cardinality
	// use that as a filter in base metric to find cardinality contribution, if Cardinality
	// is with in limit then no need to use filter
	filter := ""
	if (cardinality > uint64(cFlag.AllowedCardinalityLimit)) && !cFlag.DisableRelativeCardinality {
		cFlag.RelativeLabelNo -= 1
		if len(r.SeriesCountByFocusLabelValue) < cFlag.RelativeLabelNo {
			cFlag.RelativeLabelNo = len(r.SeriesCountByFocusLabelValue) - 1
		}

		filter = fmt.Sprintf(`%s=""`, focusLabel)
		if len(r.SeriesCountByFocusLabelValue) != 0 {
			filter = fmt.Sprintf(`%s="%s"`, focusLabel, r.SeriesCountByFocusLabelValue[cFlag.RelativeLabelNo].Name)
		}

		modifiedMetric := strings.Replace(cFlag.Metric, "{", fmt.Sprintf("{%s,", filter), 1)
		if modifiedMetric == cFlag.Metric {
			cFlag.Metric = fmt.Sprintf("%s{%s}", cFlag.Metric, filter)
		} else {
			cFlag.Metric = modifiedMetric
		}
		r, err = apiclient.MetricInfo(v1api, cFlag.Metric, focusLabel, topN, "")
		if err != nil {
			fmt.Println("Error while fetching cardinality info: ", err)
			return
		}
	}

	// In case if series is incorrect
	if len(r.SeriesCountByMetricName) == 0 {
		return
	}

	if len(r.SeriesCountByFocusLabelValue) == 0 {
		return
	}

	if r.SeriesCountByFocusLabelValue[0].Value > uint64(cFlag.AllowedCardinalityLimit) {
		fmt.Println("Cardinality is greater than allowed limit even after applying relative cardinality, use different label for relative cardinality, can't process")
		return
	}

	// By default consider all labels for finding cardinality contribution
	labelsToConsider := []string{}
	cd.cardinality = r.SeriesCountByMetricName[0].Value
	for l := range r.LabelValueCountByLabelName {
		if r.LabelValueCountByLabelName[l].Name == "__name__" {
			continue
		}
		cd.labelInfo[r.LabelValueCountByLabelName[l].Name] = labelInfo{uniqueCount: int(r.LabelValueCountByLabelName[l].Value)}

		labelsToConsider = append(labelsToConsider, r.LabelValueCountByLabelName[l].Name)
	}

	if len(cFlag.Label) != 0 {
		labelsToConsider = cFlag.Label
	}

	// create unique pairs as per label count
	pairs := createPairs(labelsToConsider, cFlag.LabelCount)
	cMap := &RWMap{m: labelsCardinalityInfo{}}

	// if label count < no of explicit labels provided then include
	// a combination of all labels also
	if len(cFlag.Label) != 0 && cFlag.LabelCount < len(cFlag.Label) {
		pairs = append(pairs, strings.Join(labelsToConsider, ","))
	}

	for p := range pairs {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			// If the diff between current time and start of the day time in UTC is less than 12 hrs then use the diff instead of 12 hours
			now := time.Now().UTC()
			startOfDayTime := time.Date(
				now.Year(), now.Month(),
				now.Day(), 0, 0, 0, 0, time.UTC).Unix()
			currentTime := now.Unix()
			cardinalityDuration := int(currentTime - startOfDayTime)
			if cardinalityDuration > cFlag.CardinalityPerDuration {
				cardinalityDuration = cFlag.CardinalityPerDuration
			}

			r, err := apiclient.GetQueryResult(v1api, cFlag.Metric, cFlag.CardinalityPerDuration, cFlag.Lag, pairs[p], apiclient.LabelCardinalityStr)
			if err != nil {
				fmt.Println("Error while finding cardinality:", err)
			}

			per := int((cd.cardinality - uint64(r)) * 100 / cd.cardinality)

			cMap.Set(pairs[p], labelInfo{uniqueCount: cd.labelInfo[pairs[p]].uniqueCount, cardinalityPer: per})

			if cFlag.DropAction {
				r, err := apiclient.GetQueryResult(v1api, cFlag.Metric, cFlag.CardinalityPerDuration, cFlag.Lag, pairs[p], apiclient.DuplicatesLabelsStr)
				if err != nil {
					fmt.Println("Error while finding duplicate labels exists:", err)
				}

				cMap.SetDropActionInfo(pairs[p], r == 1)
			}
		}(p)

	}

	wg.Wait()

	action := ""
	if cFlag.DropAction {
		action = drop
	}

	if cFlag.LabelCount == 1 {
		dumpCardinalityInfoPerLabel(cFlag.Metric, cd.cardinality, cMap.m, action, cFlag.DumpAs)
	} else {
		dumpCardinalityInfoWithoutLabels(cFlag.Metric, cd.cardinality, cd.labelInfo, cFlag.DumpAs)
		dumpCardinalityPer(cFlag.Metric, cMap.m, action, cFlag.DumpAs)
	}
}
