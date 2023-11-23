/*
Copyright © 2023 PREETI DEWANI preetidewani1990@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	DataSource string `yaml:"datasource"`
}

var (
	cfgFile string
	config  Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "metric-explorer",
	Version: "0.0.1",
	Short:   "Metric Explorer for finding the details about your TSDB",
	Long: `A tool that helps in answering: I have detected high cardinality, what to do next?.
	
Provides capability to take decisions on how to control cardinality.
Supports three modes:
1. System(system): To get system wide information about cardinality.
2. Explore (explore): To know more details about specific metric.
3. Cardinality Control(cc): To make decision to control cardinality`,
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.metric_explorer.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		fmt.Println("using default one .metric_explorer.yaml")
		viper.SetConfigFile(".metric_explorer.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error while reading configuration", err)
		os.Exit(1)
	}

	if err = viper.Unmarshal(&config); err != nil {
		fmt.Println("Error while unmarshalling config file:", err)
		os.Exit(1)
	}
}
