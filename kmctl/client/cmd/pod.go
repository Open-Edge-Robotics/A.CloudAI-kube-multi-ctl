package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

var (
	podName      string
	podNamespace string
)

// podCmd represents the pod command
var podCmd = &cobra.Command{
	Use:   "pod",
	Short: "Get a pod by name and namespace",
	Long: `Get a pod what is same name and namespace from all Kubernetes clusters.
	
	For example:
	get pod -n my-pod
	get pod -n my-pod -s my-namespace`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Get a pod: %s (%s)\n", podName, podNamespace)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				getCon := controller.NewGet(&cluster.Host, &cluster.Port)
				getCon.GetPod(&podName, &podNamespace, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	podCmd.Flags().StringVarP(&podName, "name", "n", "", "The pod name")
	podCmd.Flags().StringVarP(&podNamespace, "namespace", "s", "default", "The pod namespace")
	podCmd.MarkFlagRequired("name")
}
