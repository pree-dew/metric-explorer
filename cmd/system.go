package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"metric_explorer/mode"
)

var (
	defaultSystemChurnRateDuration        = "3600"
	defaultSystemIngestionRateDuration    = "3600"
	defaultSystemActiveTimeseriesDuration = "3600"
)

var sFlag mode.SystemFlag

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Overview of your TSDB coverage",
	Long: `Provides system wise information and has options to get overview of:

- Metrics with high ingestion rates (Under Development).
- Metrics with high cardinality.
- Metrics with high churn rates (Under Development).
- Active Timeseries (Under Development).`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().NFlag() < 2 {
			fmt.Println("Please provide atleast 1 flag, refer --help for flag information")
			os.Exit(1)
		}

		isSet := cmd.PersistentFlags().Lookup("cardinality").Changed
		if !isSet {
			sFlag.Cardinality = ""
		}

		mode.SystemInvoke(config.DataSource, sFlag)
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.PersistentFlags().BoolVar(&sFlag.TopQueries, "top-queries", false, "Stats of top queries")
	systemCmd.PersistentFlags().StringVar(&sFlag.Cardinality, "cardinality", "",
		"Provide the date for which cardinality should be calculated, format is YYYY-MM-DD")
	systemCmd.PersistentFlags().StringVar(&sFlag.TopN, "topN", "20", "Details of top N metrics")
	systemCmd.PersistentFlags().StringVar(&sFlag.TopNMaxLifeTime, "top-query-max-lifetime", "3600", "Duration to check for top queries(in seconds)")
	systemCmd.PersistentFlags().StringVar(&sFlag.DumpAs, "dump-as", "csv", "Dump format, allowed values csv, table")
	systemCmd.PersistentFlags().Lookup("cardinality").NoOptDefVal = "today"

	// Churn Rate, Ingestion Rate and Active Timeseries are currently under development
	// for system mode
}
