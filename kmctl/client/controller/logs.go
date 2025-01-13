package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"com.kubebackend/m/client/model"
	pb "com.kubebackend/m/proto"
	"google.golang.org/grpc"
)

type LogsController struct {
	client pb.KubeBackendClient
}

func NewLogs(host, port *string) *LogsController {
	return &LogsController{
		client: *GetClient(host, port),
	}
}

func (c *LogsController) GetPodLogsStream(name, namespace *string, lastLines *int, cluster *model.Cluster) {
	podLogs := &pb.GetPodLogsRequest{Name: *name, Namespace: *namespace}
	callOpts := grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024)
	stream, err := c.client.GetPodLogs(context.Background(), podLogs, callOpts)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to get pod logs: %v\n", err)
		return
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Logs for pod \"%s\" in namespace \"%s\":\n", *name, *namespace)

	for {
		logs, err := stream.Recv()
		if err == io.EOF {
			log.Printf("  End of log stream")
			break
		} else if err != nil {
			fmt.Printf("  There is no logs for pod \"%s\" in namespace \"%s\"\n", *name, *namespace)
			break
		}

		logsLine := strings.Split(logs.Log, "\n")
		if len(logsLine) > *lastLines {
			lastLogs := logsLine[len(logsLine)-*lastLines:]
			for _, line := range lastLogs {
				fmt.Printf("  %s\n", line)
			}
		} else {
			for _, line := range logsLine {
				fmt.Printf("  %s\n", line)
			}
		}
	}
}
