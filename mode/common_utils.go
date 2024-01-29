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
	drop            = "action"
)

type labelsCardinalityInfo map[string]labelInfo

type stringIntMap struct {
	key   string
	value uint64
}

type labelInfo struct {
	uniqueCount     int
	cardinalityPer  int
	values          []string
	duplicateExists bool
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

func dumpCardinalityInfoWithLabels(metric string, cardinality uint64, labels labelMap, labelValues map[string][]map[string]uint64, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", metric})
	t.AppendHeader(table.Row{"Cardinality", cardinality})
	t.AppendHeader(table.Row{"Label", "Unique Value", "Label Values"})
	t.AppendSeparator()
	for k, v := range labels {
		lString := []string{}
		for _, values := range labelValues[k] {
			for lk, lv := range values {
				lString = append(lString, fmt.Sprintf("%s - %d", lk, lv))
			}
		}

		t.AppendRow([]interface{}{k, v.uniqueCount, strings.Join(lString, "\n")})
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

func sortMap(m labelsCardinalityInfo) []stringIntMap {
	lc := make([]stringIntMap, len(m))

	i := 0
	for k, v := range m {
		lc[i] = stringIntMap{k, uint64(v.uniqueCount)}
		i++
	}

	// Sort slice based on values
	sort.Slice(lc, func(i, j int) bool {
		return lc[i].value > lc[j].value
	})
	return lc
}

func getHeaders(noOfLabels int, action string) table.Row {
	if noOfLabels == 1 {
		if action == "" {
			return table.Row{"Label", "Unique Value", "Cardinality %"}
		}

		if action == drop {
			return table.Row{"Label", "Unique Value", "Cardinality %", "Duplicate Labels Exists"}
		}
	}

	if action == "" {
		return table.Row{"Label", "Cardinality %"}
	}

	return table.Row{"Label", "Cardinality %", "Duplicate Labels Exists"}
}

func dumpCardinalityInfoPerLabel(metric string, cardinality uint64, labelInfo labelsCardinalityInfo, action, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", metric})
	t.AppendHeader(table.Row{"Cardinality", cardinality})
	t.AppendHeader(getHeaders(1, action))
	t.AppendSeparator()
	lc := sortMap(labelInfo)
	for _, v := range lc {
		if action == "" {
			t.AppendRow([]interface{}{v.key, v.value, labelInfo[v.key].cardinalityPer})
		} else {
			t.AppendRow([]interface{}{v.key, v.value, labelInfo[v.key].cardinalityPer, labelInfo[v.key].duplicateExists})
		}
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

func dumpCardinalityPer(metric string, labelInfo labelsCardinalityInfo, action, format string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Cardinality % contribution"})
	t.AppendSeparator()
	t.AppendHeader(getHeaders(2, action))
	lc := sortMap(labelInfo)
	for _, v := range lc {
		if action == "" {
			t.AppendRow([]interface{}{strings.ReplaceAll(v.key, ",", " -"), labelInfo[v.key].cardinalityPer})
		} else {
			t.AppendRow([]interface{}{strings.ReplaceAll(v.key, ",", " -"), labelInfo[v.key].cardinalityPer, labelInfo[v.key].duplicateExists})
		}
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
