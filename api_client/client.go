package apiclient

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type value struct {
	Timestamp float64
	Value     string
}

type queryParams struct {
	Metric     string
	LabelPair  string
	Duration   int
	MetricType string
}

const (
	apiTimeout               = 3 * 60 * time.Second
	scrapeIntervalQuery      = "scrape_interval(%s)"
	systemIngestionRateQuery = ""
	activeTimeSeriesQuery    = ""
	systemChurnRateQuery     = "count(count_over_time(scrape_samples_scraped{}[1h]))"
	metricResponseTimeTempl  = ""
)

const (
	LabelCardinalityStr = "labelCardinality"
	DuplicatesLabelsStr = "duplicateLabels"
)

const labelCardinalityTempl = `
count (
	group without ( {{.LabelPair}} ) (
		count_over_time ( {{.Metric}}[{{.Duration}}s] )
	)
)
`

const duplicateLabelsExistsTempl = `
sum( group without ( {{.LabelPair}} ) (
	count_over_time( {{.Metric}}[{{.Duration}}s] )
	)
) 
	!= bool 
sum(  count without ( {{.LabelPair}} ) (
	count_over_time( {{.Metric}}[{{.Duration}}s] )
	)
)
`

const metricChurnRateTempl = `
(
	count( count_over_time( {{.Metric}}[{{.Duration}}s])  ) offset 1h 
	- 
	count( count_over_time( {{.Metric}}[{{.Duration}}s] ) ) 
)*100 
/ count( count_over_time( {{.Metric}}[{{.Duration}}s] ) )`

const metricIngestionRateTempl = `
count(
	last_over_time( {{.Metric}}[{{.Duration}}s] )
) / scrape_interval( {{.Metric}} )
`

const metricSampleReceivedTempl = `
sum (
	count_over_time( {{.Metric}} [{{.Duration}}s] )
)
`

const metricActiveTimeSeriesTempl = `
count(
	last_over_time( {{.Metric}}[{{.Duration}}s] )
)
`

const metricLastResetTempl = `
count( resets( {{.Metric}}[{{.Duration}}s] ) )
`

const metricLastLossTempl = `
{{ if (eq .MetricType "Counter") }}

tlast_change_over_time(
	 sum({{.Metric}}[{{.Duration}}s]) 
)

{{ else }}

timestamp(
	sum({{.Metric}}[{{.Duration}}s])
)

{{ end }}
`

const metricSparseDurationTempl = `
avg(
	duration_over_time(
	   {{.Metric}}[{{.Duration}}s], 8m
        )
)
`

var tStore sync.Map

// LoadTmpl loads the query template
func LoadTmpl(tmplStr string) (*template.Template, error) {
	hasher := md5.New() // #nosec
	if _, err := hasher.Write([]byte(tmplStr)); err != nil {
		return nil, err
	}

	tHash := hex.EncodeToString(hasher.Sum(nil))

	t, ok := tStore.Load(tHash)
	if ok {
		return t.(*template.Template), nil
	}

	tmpl, err := template.New(tHash).Parse(tmplStr)
	if err != nil {
		return nil, err
	}

	tStore.Store(tHash, tmpl)
	return tmpl, nil
}

func createQuery(params queryParams, templ string) (string, error) {
	var q strings.Builder
	switch templ {
	case LabelCardinalityStr:
		templ = labelCardinalityTempl
	case DuplicatesLabelsStr:
		templ = duplicateLabelsExistsTempl
	}

	if t, err := LoadTmpl(templ); err != nil {
		return "", err
	} else if err := t.Execute(&q, params); err != nil {
		return "", err
	}

	return q.String(), nil
}

func UnmarshalJSON(b []byte) ([]value, error) {
	v := []map[string]interface{}{}
	sv := []value{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return sv, err
	}

	for i := range v {
		s := value{}
		for _, kv := range v[i] {
			if reflect.TypeOf(kv).String() == reflect.TypeOf(v[0]).String() {
				continue
			}

			for j := range kv.([]interface{}) {
				if j == 0 {
					s.Timestamp = kv.([]interface{})[j].(float64)
					continue
				}

				s.Value = kv.([]interface{})[j].(string)
			}
		}

		sv = append(sv, s)
	}

	return sv, err
}

func TopMetrics(v1api v1.API, topN, date string) (v1.TSDBResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	tsDBRes, err := v1api.TSDB(ctx, topN, date)
	if err != nil {
		return v1.TSDBResult{}, err
	}

	return tsDBRes, err
}

func TopQueries(v1api v1.API, topN, topNMaxLifeTime string) (v1.TopQueriesResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	topRes, err := v1api.TopQueries(ctx, topN, topNMaxLifeTime)
	if err != nil {
		return v1.TopQueriesResult{}, err
	}

	return topRes, err
}

func MetricInfo(v1api v1.API, metric, focusLabel string, topN, date string) (v1.TSDBWithMetricResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	return v1api.TSDBWithMetric(ctx, metric, focusLabel, topN, date)
}

func ScrapeInterval(v1api v1.API, metric string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	r, _, err := v1api.Query(ctx, fmt.Sprintf(scrapeIntervalQuery, metric), time.Now())
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.Atoi(values[0].Value)
}

func GetQueryResult(v1api v1.API, metric string, duration, offset int, lPair string, templType string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration, LabelPair: lPair}
	query, err := createQuery(params, templType)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.Atoi(values[0].Value)
}

func ResponseTime(v1api v1.API, metric string, duration, offset int) (float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricResponseTimeTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.QueryRange(ctx, query, v1.Range{Start: time.Now().Add(-time.Duration(offset) * time.Second), End: time.Now(), Step: time.Duration(60)})
	if err != nil {
		return 0, err
	}

	return float32(r.Type()), nil
}

func MetricSparse(v1api v1.API, metric string, duration, offset int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricSparseDurationTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	val, err := strconv.ParseFloat(values[0].Value, 64)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ResetTime(v1api v1.API, metric string, duration, offset int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricLastResetTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.Atoi(values[0].Value)
}

func SampleReceived(v1api v1.API, metric string, duration, offset int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricSampleReceivedTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.ParseInt(values[0].Value, 10, 64)
}

func ActiveTimeSeries(v1api v1.API, metric string, duration, offset int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricActiveTimeSeriesTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.ParseInt(values[0].Value, 10, 64)
}

func LastLoss(v1api v1.API, metric string, duration, offset int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration, MetricType: "Counter"}
	query, err := createQuery(params, metricLastLossTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	val, err := strconv.ParseFloat(values[0].Value, 64)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ChurnRate(v1api v1.API, metric string, duration, offset int) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricChurnRateTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.ParseFloat(values[0].Value, 64)
}

func IngestionRate(v1api v1.API, metric string, duration, offset int) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	params := queryParams{Metric: metric, Duration: duration}
	query, err := createQuery(params, metricIngestionRateTempl)
	if err != nil {
		return 0, err
	}

	r, _, err := v1api.Query(ctx, query, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.ParseFloat(values[0].Value, 64)
}

// Reference: https://www.robustperception.io/finding-churning-targets-in-prometheus-with-scrape_series_added
func SystemChurnRate(v1api v1.API, offset int) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	r, _, err := v1api.Query(ctx, systemChurnRateQuery, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.ParseFloat(values[0].Value, 64)
}

func SystemIngestionRate(v1api v1.API, offset int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	r, _, err := v1api.Query(ctx, systemIngestionRateQuery, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.Atoi(values[0].Value)
}

func SystemActiveTimeSeries(v1api v1.API, offset int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	r, _, err := v1api.Query(ctx, activeTimeSeriesQuery, time.Now().Add(-time.Duration(offset)*time.Second))
	if err != nil {
		return 0, err
	}

	by, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	values, err := UnmarshalJSON(by)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, nil
	}

	return strconv.Atoi(values[0].Value)
}
