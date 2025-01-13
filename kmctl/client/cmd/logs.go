package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
)

var (
	logsPodName      string
	logsPodNamespace string
	logsLines        int
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from a pod what is same name and namespace from all Kubernetes clusters",
	Long: `Get logs from a pod what is same name and namespace from all Kubernetes clusters.
	
	For example:
	logs -n <pod-name> -s <pod-namespace>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Get logs: %s\n", logsPodName)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				logsCon := controller.NewLogs(&cluster.Host, &cluster.Port)
				logsCon.GetPodLogsStream(&logsPodName, &logsPodNamespace, &logsLines, &cluster)
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsPodName, "name", "n", "", "The pod name")
	logsCmd.Flags().StringVarP(&logsPodNamespace, "namespace", "s", "default", "The pod namespace")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "l", 30, "The number of lines to show from the end of the logs")
	logsCmd.MarkFlagRequired("name")
}
