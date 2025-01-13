package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

var yamlPath string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply kubernetes yaml file to the all clusters",
	Long: `Apply kubernetes yaml file to the all clusters.
	
	For example:
	apply -f <yaml-file-path>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Apply: %s\n", yamlPath)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				yamlCon := controller.NewYaml(&cluster.Host, &cluster.Port)
				yamlCon.ApplyYaml(&yamlPath, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	applyCmd.Flags().StringVarP(&yamlPath, "file", "f", "", "The yaml file path")
	applyCmd.MarkFlagRequired("file")
}
