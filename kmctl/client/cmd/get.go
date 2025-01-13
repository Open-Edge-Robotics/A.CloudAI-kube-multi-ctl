package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get kubernetes resources",
	Long: `Get kubernetes resources.
	
	Node: Get a node what is same name from all Kubernetes clusters.
	Nodes: Get all nodes from all kubernetes clusters.
	Pod: Get a pod what is same name and namespace from all Kubernetes clusters.
	Pods: Get all pods in namespace from all kubernetes clusters.`,
}

func init() {
	getCmd.AddCommand(nodeCmd)
	getCmd.AddCommand(nodesCmd)
	getCmd.AddCommand(podCmd)
	getCmd.AddCommand(podsCmd)
}
