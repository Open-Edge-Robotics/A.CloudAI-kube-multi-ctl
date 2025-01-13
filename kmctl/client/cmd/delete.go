package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete kubernetes yaml file from the all clusters",
	Long: `Delete kubernetes yaml file from the all clusters.
	
	For example:
	delete -f <yaml-file-path>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Delete: %s\n", yamlPath)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				yamlCon := controller.NewYaml(&cluster.Host, &cluster.Port)
				yamlCon.DeleteYaml(&yamlPath, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&yamlPath, "file", "f", "", "The yaml file path")
	deleteCmd.MarkFlagRequired("file")
}
