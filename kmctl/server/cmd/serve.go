package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "com.kubebackend/m/proto"
	"com.kubebackend/m/server/controller"
)

var (
	host       string
	port       string
	kubeconfig string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the gRPC server for the CLI",
	Long: `Serve the gRPC server for the CLI.
	You can set the host and port to listen on.
	Also, you can set the kubeconfig path to connect to the Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting server...")

		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		log.Printf("Listening on %s:%s", host, port)

		s := controller.NewServer(kubeconfig)
		grpcServer := grpc.NewServer()
		pb.RegisterKubeBackendServer(grpcServer, s)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	},
}

func init() {
	serveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to listen on")
	serveCmd.Flags().StringVarP(&port, "port", "P", "50051", "Port to listen on")
	serveCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "K", "", "Path to kubeconfig")
}
