package controller

import (
	"fmt"
	"log"

	pb "com.kubebackend/m/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetClient(host, port *string) *pb.KubeBackendClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	addr := fmt.Sprintf("%s:%s", *host, *port)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Printf("Failed to connect to server: %v\n", err)
	}

	client := pb.NewKubeBackendClient(conn)

	return &client
}
