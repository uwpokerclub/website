package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "The UWPSC Admin API handles logic for managing the UWPSC events and semsters",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&PORT, "port", "p", "5000", "The port number for the server to run on.")
	startCmd.Flags().BoolVar(&RUN_MIGRATIONS, "run-migrations", false, "Run the SQL migrations on startup.")
}
