package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

// nodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Get all nodes from all kubernetes clusters.",
	Long: `Get all nodes from all kubernetes clusters.
	
	For example:
	get nodes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Get nodes\n")
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				getCon := controller.NewGet(&cluster.Host, &cluster.Port)
				getCon.GetNodes(&cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	// getCmd.AddCommand(nodesCmd)
}
