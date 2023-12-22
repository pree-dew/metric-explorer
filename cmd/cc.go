package cmd

import (
	"fmt"
	"os"

	"github.com/pree-dew/metric-explorer/mode"

	"github.com/spf13/cobra"
)

var c mode.CardinalityFlag

// ccCmd represents the cc command
var ccCmd = &cobra.Command{
	Use:   "cc [metric]",
	Short: "To understand the cardinality distribution of a metric",
	Long: `Provides capability to find:

1. Cardinality of a metric.
2. Find cardinality as per specific filter.
3. Find unique counts of labels.
4. Find cardinality contribution of each label or pair of labels.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Arg(0) == "" {
			fmt.Println("Metric name cannot be empty for metric info mode")
			os.Exit(1)
		}

		c.Metric = cmd.Flags().Arg(0)

		mode.CardinalityInvoke(config.DataSource, c)
	},
}

func init() {
	rootCmd.AddCommand(ccCmd)
	ccCmd.PersistentFlags().StringVar(&c.FilterLabel, "filter-label", "", "Use filter to consider to get limited timeseries")
	ccCmd.PersistentFlags().IntVar(&c.LabelCount, "label-count", 1, "No. of labels to consider for cardinality, currently supports 1 and 2")
	ccCmd.PersistentFlags().IntVar(&c.CardinalityPerDuration, "cc-duration", 43200,
		"Cardinality duration for labels contribution. [Note]: this is not for unique label count of each label")
	ccCmd.PersistentFlags().IntVar(&c.Lag, "lag", 60, "Lag to consider from current time to calculate cardinality")
	ccCmd.PersistentFlags().IntVar(&c.RelativeLabelNo, "relative-label-no", 3,
		"Which label value should be used to create a relative query, this number is in decreasing order of cardinality contribution")
	ccCmd.PersistentFlags().Int64Var(&c.AllowedCardinalityLimit, "allowed-cardinality-limit", 30000,
		"Specify the acceptable limit, if the cardinality execeeds beyond this then relative cardinality kicks in.")
	ccCmd.PersistentFlags().StringVar(&c.DumpAs, "dump-as", "csv", "Dump format, allowed values csv, table")
	ccCmd.PersistentFlags().StringArrayVar(&c.Label, "labels", []string{}, "Labels to consider for cardinality")
}
