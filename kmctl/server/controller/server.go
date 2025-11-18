package controller

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"com.kubebackend/m/client/controller"
	pb "com.kubebackend/m/proto"
)

type server struct {
	kubeCon *KubeController
	pb.UnimplementedKubeBackendServer
}

func (s *server) GetNodes(ctx context.Context, in *pb.GetNodesRequest) (*pb.NodeList, error) {
	nodes, err := s.kubeCon.GetNodes()
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
		return nil, err
	}

	var nodeList pb.NodeList
	for _, node := range nodes.Items {
		nodeList.Nodes = append(nodeList.Nodes, &pb.Node{
			Name:          node.Name,
			Ip:            node.Annotations["k3s.io/internal-ip"],
			Arch:          node.Labels["kubernetes.io/arch"],
			KernelVersion: node.Status.NodeInfo.KernelVersion,
		})
	}

	log.Printf("GetNodesResponse: all")

	return &nodeList, nil
}

func (s *server) GetNode(ctx context.Context, in *pb.GetNodeRequest) (*pb.Node, error) {
	node, err := s.kubeCon.GetNode(in.Name)
	if err != nil {
		log.Printf("Failed to get node: %v", err)
		return nil, err
	}

	log.Printf("GetNodeResponse: %s", in.Name)

	return &pb.Node{
		Name:          node.Name,
		Ip:            node.Annotations["k3s.io/internal-ip"],
		Arch:          node.Labels["kubernetes.io/arch"],
		KernelVersion: node.Status.NodeInfo.KernelVersion,
	}, nil
}

func (s *server) GetPods(ctx context.Context, in *pb.GetPodsRequest) (*pb.PodList, error) {
	log.Printf("GetPodsRequest: %v", in)
	pods, err := s.kubeCon.GetPods(&in.Namespace)
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
		return nil, err
	}

	var podList pb.PodList
	for _, pod := range pods.Items {
		podList.Pods = append(podList.Pods, &pb.Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Label:     pod.Labels["app"],
			Image:     pod.Spec.Containers[0].Image,
		})
	}

	log.Printf("GetPodsResponse: %s", in.Namespace)

	return &podList, nil
}

func (s *server) GetPod(ctx context.Context, in *pb.GetPodRequest) (*pb.Pod, error) {
	pod, err := s.kubeCon.GetPod(in.Namespace, in.Name)
	if err != nil {
		log.Printf("Failed to get pod: %v", err)
		return nil, err
	}

	log.Printf("GetPodResponse: %s", in.Name)

	return &pb.Pod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Label:     pod.Labels["app"],
		Image:     pod.Spec.Containers[0].Image,
	}, nil
}

func (s *server) GetPodLogs(in *pb.GetPodLogsRequest, stream pb.KubeBackend_GetPodLogsServer) error {
	logs, err := s.kubeCon.GetPodLogs(in.Namespace, in.Name)
	if err != nil {
		log.Printf("Failed to get pod logs: %v", err)
		return err
	}

	if err := stream.Send(&pb.GetPodLogsResponse{
		Log: *logs,
	}); err != nil {
		log.Printf("Failed to send logs: %v", err)
		return err
	}

	log.Printf("GetPodLogsResponse: %s", in.Name)

	return nil
}

func (s *server) ApplyYaml(ctx context.Context, in *pb.ApplyYamlRequest) (*pb.ApplyYamlResponse, error) {
	message, err := s.kubeCon.ApplyYaml(in.Yaml)
	if err != nil {
		log.Printf("Failed to apply yaml: %v", err)
		return nil, err
	}

	log.Printf("ApplyYamlResponse: %s", *message)

	return &pb.ApplyYamlResponse{
		Message: *message,
	}, nil
}

func (s *server) DeleteYaml(ctx context.Context, in *pb.ApplyYamlRequest) (*pb.ApplyYamlResponse, error) {
	message, err := s.kubeCon.DeleteYaml(in.Yaml)
	if err != nil {
		log.Printf("Failed to delete yaml: %v", err)
		return nil, err
	}

	log.Printf("DeleteYamlResponse: %s", *message)

	return &pb.ApplyYamlResponse{
		Message: *message,
	}, nil
}

func (s *server) UpgradeYaml(ctx context.Context, in *pb.UpgradeYamlRequest) (*pb.UpgradeYamlResponse, error) {
	message, err := s.kubeCon.ApplyYaml(in.Yaml)
	if err != nil {
		log.Printf("Failed to upgrade yaml: %v", err)
		return nil, err
	}

	log.Printf("UpgradeYamlResponse: %s", *message)
	major, minor1, minor2, err := parseVersion(in.Version)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	databasePath := "/database/database.db"
	dbCon := controller.NewDB(&databasePath)

	repoNames := []string{"MICOM_MANAGER", "DEVICE_BRINGUP", "NAVIGATION", "MIDDLEWARE"}
	tableNames := []string{"micom_managers", "device_bringups", "navigations", "middlewares"}

	if in.Type < 0 || in.Type >= int32(len(repoNames)) {
		log.Printf("Invalid upgrade type")
		return nil, fmt.Errorf("invalid upgrade type")
	}

	log.Printf("Upgrade %s Ver %d.%d.%d\n", repoNames[in.Type], major, minor1, minor2)
	repo := controller.Repo{
		Repo_name:   repoNames[in.Type],
		Ver_major:   major,
		Ver_minor_1: minor1,
		Ver_minor_2: minor2,
		Updated_at:  "CURRENT_TIMESTAMP",
	}
	dbCon.InsertRepo(&tableNames[in.Type], &repo)

	log.Printf("UpgradeYamlResponse: %s", *message)

	return &pb.UpgradeYamlResponse{
		Message: *message,
	}, nil
}

func parseVersion(version string) (int, int, int, error) {
	versions := strings.Split(version, ".")
	if len(versions) != 3 {
		return 0, 0, 0, fmt.Errorf("version must be in the format of <00.00.00>")
	}

	major, err := strconv.Atoi(versions[0])
	if err != nil {
		return 0, 0, 0, err
	}

	minor1, err := strconv.Atoi(versions[1])
	if err != nil {
		return 0, 0, 0, err
	}

	minor2, err := strconv.Atoi(versions[2])
	if err != nil {
		return 0, 0, 0, err
	}

	return major, minor1, minor2, nil
}

func NewServer(kubeconfig string) *server {
	kubeCon, err := NewKubeController(&kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create kube controller: %v", err)
	}

	return &server{
		kubeCon: kubeCon,
	}
}
