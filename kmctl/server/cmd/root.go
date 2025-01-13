package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kube multi CLI server",
	Short: "This is a simple CLI server to operate multiple clusters",
	Long: `This is a simple CLI server to operate multiple clusters.
	It can be used to serve get nodes, pods, etc. Also, it can be used to apply deployment YAML and delete.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
