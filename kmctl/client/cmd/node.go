package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

var name string

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Get a node what is same name from all Kubernetes clusters",
	Long: `Get a node what is same name from all Kubernetes clusters.
	
	For example:
	get node -n <node-name>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Get a node: %s\n", name)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				getCon := controller.NewGet(&cluster.Host, &cluster.Port)
				getCon.GetNode(&name, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	nodeCmd.Flags().StringVarP(&name, "name", "n", "", "The node name")
	nodeCmd.MarkFlagRequired("name")
}
