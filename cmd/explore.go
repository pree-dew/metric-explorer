package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pree-dew/metric-explorer/mode"

	"github.com/spf13/cobra"
)

var m mode.MetricFlag

// minfoCmd represents the minfo command
var minfoCmd = &cobra.Command{
	Use:   "explore [metric]",
	Short: "Provide metrics information",
	Long: `It accepts a series and provide options to:

- Cardinality.
- Scrape interval.
- Top N contributing labels, count and few samples.
- Last loss of signal over in past x seconds.
- Average sample received in past x seconds.
- Ingestion rate in past x seconds.
- Sparse Percentage of a metric. Average duration for which metric is absent.
- Active timeseries in past x seconds
- If counter, last reset times.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Arg(0) == "" {
			fmt.Println("Metric name cannot be empty for metric info mode")
			os.Exit(1)
		}

		m.Metric = cmd.Flags().Arg(0)

		if cmd.Flags().NFlag() < 2 {
			fmt.Println("Please provide atleast 1 flag, refer --help for flag information")
			os.Exit(1)
		}

		isSet := cmd.PersistentFlags().Lookup("churn-rate").Changed
		if !isSet {
			m.ChurnRate = 0
		}

		isSet = cmd.PersistentFlags().Lookup("ingestion-rate").Changed
		if !isSet {
			m.IRate = 0
		}

		isSet = cmd.PersistentFlags().Lookup("sample-received").Changed
		if !isSet {
			m.SampleReceived = 0
		}

		isSet = cmd.PersistentFlags().Lookup("response-time").Changed
		if !isSet {
			m.RespTime = 0
		}

		isSet = cmd.PersistentFlags().Lookup("loss").Changed
		if !isSet {
			m.Loss = 0
		}

		isSet = cmd.PersistentFlags().Lookup("reset-counts").Changed
		if !isSet {
			m.ResetTime = 0
		}

		isSet = cmd.PersistentFlags().Lookup("sparse").Changed
		if !isSet {
			m.SparseDuration = 0
		}

		isSet = cmd.PersistentFlags().Lookup("active-timeseries").Changed
		if !isSet {
			m.ActiveTimeSeries = 0
		}

		cardinalityFlag := cmd.PersistentFlags().Lookup("cardinality")
		isSet = cardinalityFlag.Changed
		if !isSet {
			m.Cardinality = ""
		}

		if cardinalityFlag.Value.String() == "today" {
			m.Cardinality = time.Now().UTC().Format("2006-01-02")
		}

		mode.MInfoInvoke(config.DataSource, m)
	},
}

func init() {
	rootCmd.AddCommand(minfoCmd)

	// minfoCmd.PersistentFlags().StringVar(&m.Metric, "metric", "", "Metric name under consideration")
	minfoCmd.PersistentFlags().StringVar(&m.Cardinality, "cardinality", "", "Cardinality of metric")
	minfoCmd.PersistentFlags().BoolVar(&m.ScrapeInterval, "scrape-interval", false, "Scrape Interval of metric")
	minfoCmd.PersistentFlags().IntVar(&m.RespTime, "response-time", 300, "Response time for x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.Loss, "loss", 3600, "Last loss of signal for a duration of x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.SampleReceived, "sample-received", 1800, "Average samples received in past x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.IRate, "ingestion-rate", 300, "Ingestion rate past x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.ChurnRate, "churn-rate", 3600, "Churn rate past x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.SparseDuration, "sparse", 3600, "Check sparness for duration over x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.ResetTime, "reset-counts", 3600, "No. of times counter reset in x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.ActiveTimeSeries, "active-timeseries", 3600, "No. of active timeseries in duration of x seconds")
	minfoCmd.PersistentFlags().IntVar(&m.Lag, "lag", 60, "Lag to consider for collecting stats")
	minfoCmd.PersistentFlags().StringVar(&m.DumpAs, "dump-as", "csv", "Dump format, allowed values csv, table")
	minfoCmd.PersistentFlags().StringVar(&m.LabelCount, "label-count", "5", "No. of label values to present for each label along with cardinality information, arranged in decreassing order")

	minfoCmd.PersistentFlags().Lookup("response-time").NoOptDefVal = "300"
	minfoCmd.PersistentFlags().Lookup("loss").NoOptDefVal = "3600"
	minfoCmd.PersistentFlags().Lookup("sample-received").NoOptDefVal = "1800"
	minfoCmd.PersistentFlags().Lookup("ingestion-rate").NoOptDefVal = "300"
	minfoCmd.PersistentFlags().Lookup("churn-rate").NoOptDefVal = "3600"
	minfoCmd.PersistentFlags().Lookup("sparse").NoOptDefVal = "3600"
	minfoCmd.PersistentFlags().Lookup("reset-counts").NoOptDefVal = "3600"
	minfoCmd.PersistentFlags().Lookup("active-timeseries").NoOptDefVal = "3600"
	minfoCmd.PersistentFlags().Lookup("cardinality").NoOptDefVal = "today"
}
