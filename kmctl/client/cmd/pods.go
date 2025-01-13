package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

var namespace string

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Get all pods in namespace from all kubernetes clusters",
	Long: `Get all pods in namespace from all kubernetes clusters.
	
	For example:
	get pods
	get pods -s my-namespace`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Get pods in namespace: %s\n", namespace)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				getCon := controller.NewGet(&cluster.Host, &cluster.Port)
				getCon.GetPods(&namespace, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	podsCmd.Flags().StringVarP(&namespace, "namespace", "s", "default", "The pod namespace")
}
