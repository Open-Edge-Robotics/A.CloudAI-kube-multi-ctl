package controller

import (
	"context"
	"fmt"
	"log"
	"os"

	"com.kubebackend/m/client/model"
	pb "com.kubebackend/m/proto"
)

type YamlController struct {
	client pb.KubeBackendClient
}

func NewYaml(host, port *string) *YamlController {
	return &YamlController{
		client: *GetClient(host, port),
	}
}

func (c *YamlController) ApplyYaml(path *string, cluster *model.Cluster) error {
	yamlFile, err := os.ReadFile(*path)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to read yaml file: %v\n", err)
		return err
	}

	applyYaml := &pb.ApplyYamlRequest{Yaml: string(yamlFile)}
	_, err = c.client.ApplyYaml(context.Background(), applyYaml)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to apply yaml: %v\n", err)
		return err
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Apply Yaml Response: %s\n", *path)

	return nil
}

func (c *YamlController) DeleteYaml(path *string, cluster *model.Cluster) error {
	yamlFile, err := os.ReadFile(*path)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to read yaml file: %v\n", err)
		return err
	}

	applyYaml := &pb.ApplyYamlRequest{Yaml: string(yamlFile)}
	_, err = c.client.DeleteYaml(context.Background(), applyYaml)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to delete yaml: %v\n", err)
		return err
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Delete Yaml Response: %s\n", *path)

	return nil
}

func (c *YamlController) UpgradeYaml(updateType *int, version *string, path *string, cluster *model.Cluster) error {
	yamlFile, err := os.ReadFile(*path)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to read yaml file: %v\n", err)
		return err
	}

	upgradeYaml := &pb.UpgradeYamlRequest{Yaml: string(yamlFile), Version: *version, Type: int32(*updateType)}
	_, err = c.client.UpgradeYaml(context.Background(), upgradeYaml)
	if err != nil {
		fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
		log.Printf("Failed to upgrade yaml: %v\n", err)
		return err
	}

	fmt.Printf("Cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("  Upgrade Yaml Response: %s\n", *path)

	return nil
}
