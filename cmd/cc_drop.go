package cmd

import (
	"fmt"
	"os"

	"github.com/pree-dew/metric-explorer/mode"

	"github.com/spf13/cobra"
)

// ccDropCmd helps in analysis of metric if drop action is selected
var ccDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "To provide analysis if drop action is selected",
	Long: `Provides capability to find:

1. If any label or combination is dropped, is it going to result into duplicates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: fix this positional arg passing across sub-commands
		if cmd.Flags().Arg(0) == "" {
			fmt.Println("Metric name cannot be empty for metric info mode")
			os.Exit(1)
		}

		c.Metric = cmd.Flags().Arg(0)

		c.DropAction = true
		mode.CardinalityInvoke(config.DataSource, c)
	},
}

func init() {
	ccCmd.AddCommand(ccDropCmd)
}
