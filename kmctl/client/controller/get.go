package controller

import (
	"context"
	"fmt"

	"com.kubebackend/m/client/model"
	pb "com.kubebackend/m/proto"
)

type GetController struct {
	client pb.KubeBackendClient
}

func NewGet(host, port *string) *GetController {
	return &GetController{
		client: *GetClient(host, port),
	}
}

func (c *GetController) GetNode(name *string, cluster *model.Cluster) {
	node := &pb.GetNodeRequest{Name: *name}
	nodeInfo, err := c.client.GetNode(context.Background(), node)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		fmt.Println("  There is no node with that name")
		return
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Name: %s\n", nodeInfo.Name)
	fmt.Printf("  IP: %s\n", nodeInfo.Ip)
	fmt.Printf("  Architecture: %s\n", nodeInfo.Arch)
	fmt.Printf("  Kernel Version: %s\n", nodeInfo.KernelVersion)
}

func (c *GetController) GetNodes(cluster *model.Cluster) {
	nodes := &pb.GetNodesRequest{}
	nodeList, err := c.client.GetNodes(context.Background(), nodes)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		fmt.Println("  There are no nodes")
		return
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	for _, node := range nodeList.Nodes {
		fmt.Printf("  %s: %s\n", node.Name, node.Ip)
	}
}

func (c *GetController) GetPod(name *string, namespace *string, cluster *model.Cluster) {
	pod := &pb.GetPodRequest{Name: *name, Namespace: *namespace}
	podInfo, err := c.client.GetPod(context.Background(), pod)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		fmt.Println("  There is no pod with that name")
		return
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Pod: %s\n", podInfo.Name)
	fmt.Printf("  Namespace: %s\n", podInfo.Namespace)
	fmt.Printf("  Label: %v\n", podInfo.Label)
	fmt.Printf("  Status: %s\n", podInfo.Status)
	fmt.Printf("  Image: %s\n", podInfo.Image)
}

func (c *GetController) GetPods(namespace *string, cluster *model.Cluster) {
	pods := &pb.GetPodsRequest{Namespace: *namespace}
	podList, err := c.client.GetPods(context.Background(), pods)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		fmt.Println("  There are no pods")
		return
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	maxNameLength := 0
	for _, pod := range podList.Pods {
		if len(pod.Name) > maxNameLength {
			maxNameLength = len(pod.Name)
		}
	}

	for _, pod := range podList.Pods {
		fmt.Printf("  %-*s\t%s\n", maxNameLength, pod.Name, pod.Status)
	}
}
