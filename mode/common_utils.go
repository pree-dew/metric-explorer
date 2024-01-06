package mode

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	topN            = "20"
	relativeLabelNo = 10
	focusLabel      = "job"
)

type cardinalityPer map[string]uint64

type stringIntMap struct {
	key   string
	value uint64
}

type labelInfo struct {
	uniqueCount    int
	cardinalityPer int
	values         []string
}

type labelMap map[string]labelInfo

func dumpTopQueriesView(topQueries []map[string]interface{}, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"top queries"})
	t.AppendHeader(table.Row{"Query", "Time Range(in seconds)", "Average Response Time(in seconds)", "Count"})
	t.AppendSeparator()
	for i := range topQueries {
		in := make([]interface{}, 4)
		for k, v := range topQueries[i] {
			switch k {
			case "query":
				in[0] = v
			case "timeRangeSeconds":
				in[1] = v
			case "avgDurationSeconds":
				in[2] = v
			case "count":
				in[3] = v
			}
		}
		t.AppendRow(in)
		t.AppendSeparator()
	}

	t.AppendSeparator()
	t.SetStyle(table.StyleRounded)
	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}

func dumpCardinalityInfoWithLabels(metric string, cardinality int64, labels labelMap, labelValues map[string][]string, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", metric})
	t.AppendHeader(table.Row{"Cardinality", cardinality})
	t.AppendHeader(table.Row{"Label", "Unique Value", "Label Values"})
	t.AppendSeparator()
	for k, v := range labels {
		t.AppendRow([]interface{}{k, v.uniqueCount, strings.Join(labelValues[k], ",")})
	}

	t.AppendSeparator()
	t.SetStyle(table.StyleRounded)
	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}

func sortLabelMap(labels labelMap) []stringIntMap {
	lc := make([]stringIntMap, len(labels))

	i := 0
	for k, v := range labels {
		lc[i] = stringIntMap{k, uint64(v.uniqueCount)}
		i++
	}

	// Sort slice based on values
	sort.Slice(lc, func(i, j int) bool {
		return lc[i].value > lc[j].value
	})
	return lc
}

func sortMap(m map[string]uint64) []stringIntMap {
	lc := make([]stringIntMap, len(m))

	i := 0
	for k, v := range m {
		lc[i] = stringIntMap{k, v}
		i++
	}

	// Sort slice based on values
	sort.Slice(lc, func(i, j int) bool {
		return lc[i].value > lc[j].value
	})
	return lc
}

func dumpCardinalityInfoPerLabel(metric string, cardinality uint64, labels labelMap, cPer cardinalityPer, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", metric})
	t.AppendHeader(table.Row{"Cardinality", cardinality})
	t.AppendHeader(table.Row{"Label", "Unique Value", "Cardinality %"})
	t.AppendSeparator()
	lc := sortLabelMap(labels)
	for _, v := range lc {
		t.AppendRow([]interface{}{v.key, v.value, cPer[v.key]})
	}

	t.AppendSeparator()
	t.SetStyle(table.StyleRounded)
	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}

func dumpCardinalityInfoWithoutLabels(metric string, cardinality uint64, labels labelMap, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", metric})
	t.AppendHeader(table.Row{"Cardinality", cardinality})
	t.AppendHeader(table.Row{"Label", "Unique Value"})
	t.AppendSeparator()
	lc := sortLabelMap(labels)
	for _, v := range lc {
		t.AppendRow([]interface{}{v.key, v.value})
	}

	t.AppendSeparator()
	t.SetStyle(table.StyleRounded)
	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}

func dumpCardinalityPer(metric string, cPer cardinalityPer, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Cardinality % contribution"})
	t.AppendSeparator()
	t.AppendHeader(table.Row{"Label", "Unique Value"})
	lc := sortMap(cPer)
	for _, v := range lc {
		t.AppendRow([]interface{}{strings.ReplaceAll(v.key, ",", " -"), v.value})
	}

	t.SetStyle(table.StyleRounded)
	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}

func dumpSystemView(totalSeries uint64, arr []metricSeriesCount, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Total Timeseries", totalSeries})
	t.AppendHeader(table.Row{"Metric", "Cardinality", "Cardinality %"})
	for i := range arr {
		t.AppendRow([]interface{}{arr[i].name, arr[i].series, arr[i].percentage})
	}

	t.SetStyle(table.StyleRounded)

	if format == "csv" {
		t.RenderCSV()
	} else {
		t.Render()
	}
	fmt.Println()
}
