package controller

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubeyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"gopkg.in/yaml.v3"
)

type kubeController struct {
	clientset *kubernetes.Clientset
}

func GetKubeClient(path *string) (*kubernetes.Clientset, error) {
	kubeHostEnv := os.Getenv("KUBERNETES_SERVICE_HOST")

	var clientset *kubernetes.Clientset

	if kubeHostEnv == "" {
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			kubeconfig = *path
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Printf("Failed to get k8s config: %v", err)
			return nil, err
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Printf("Failed to create k8s client: %v", err)
			return nil, err
		}
	}
	return clientset, nil
}

func NewKubeController(path *string) (*kubeController, error) {
	clientset, err := GetKubeClient(path)
	if err != nil {
		log.Printf("Failed to create k8s client: %v", err)
		return nil, err
	}

	return &kubeController{
		clientset: clientset,
	}, nil
}

func (k *kubeController) GetNodes() (*corev1.NodeList, error) {
	nodes, err := k.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list nodes: %v", err)
		return nil, err
	}

	return nodes, nil
}

func (k *kubeController) GetNode(name string) (*corev1.Node, error) {
	node, err := k.clientset.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Failed to get node: %v", err)
		return &corev1.Node{}, err
	}

	return node, nil
}

func (k *kubeController) GetPods(namespace *string) (*corev1.PodList, error) {
	pods, err := k.clientset.CoreV1().Pods(*namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return nil, err
	}

	return pods, nil
}

func (k *kubeController) GetPod(namespace, name string) (*corev1.Pod, error) {
	pod, err := k.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Failed to get pod: %v", err)
		return &corev1.Pod{}, err
	}

	return pod, nil
}

func (k *kubeController) GetPodLogs(namespace, name string) (string, error) {
	podLogOptions := corev1.PodLogOptions{}
	req := k.clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOptions)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		log.Printf("Failed to get pod logs: %v", err)
		return "", err
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		log.Printf("Failed to copy pod logs: %v", err)
		return "", err
	}

	return buf.String(), nil
}

func getYamlKind(yamlString string) (string, error) {
	yamlDecoder := yaml.NewDecoder(strings.NewReader(yamlString))
	yamlContent := make(map[string]interface{})
	err := yamlDecoder.Decode(&yamlContent)
	if err != nil {
		log.Printf("Failed to decode yaml: %v", err)
		return "", err
	}

	return yamlContent["kind"].(string), nil
}

func yamlToJson(yamlString string) ([]byte, error) {
	yamlDecoder := yaml.NewDecoder(strings.NewReader(yamlString))
	yamlContent := make(map[string]interface{})
	err := yamlDecoder.Decode(&yamlContent)
	if err != nil {
		log.Printf("Failed to decode yaml: %v", err)
		return nil, err
	}

	jsonFile, err := json.Marshal(yamlContent)
	if err != nil {
		log.Printf("Failed to marshal yaml: %v", err)
		return nil, err
	}

	return jsonFile, nil
}

func (k *kubeController) ApplyYaml(yamlString string) (string, error) {
	kind, err := getYamlKind(yamlString)
	if err != nil {
		log.Printf("Failed to get yaml kind: %v", err)
		return "", err
	}

	jsonFile, err := yamlToJson(yamlString)
	if err != nil {
		log.Printf("Failed to convert yaml to json: %v", err)
		return "", err
	}

	switch kind {
	case "Deployment":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &appv1.Deployment{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.clientset.AppsV1().Deployments(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				_, err = k.clientset.AppsV1().Deployments(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
				if err != nil {
					log.Printf("Failed to update deployment: %v", err)
					return "", err
				}
			} else {
				log.Printf("Failed to create deployment: %v", err)
				return "", err
			}
		}

		return fmt.Sprintf("Successfully applied %s %s", kind, obj.Name), nil

	case "Service":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Service{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.clientset.CoreV1().Services(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				_, err = k.clientset.CoreV1().Services(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
				if err != nil {
					log.Printf("Failed to update service: %v", err)
					return "", err
				}
			} else {
				log.Printf("Failed to create service: %v", err)
				return "", err
			}
		}

		return fmt.Sprintf("Successfully applied %s %s", kind, obj.Name), nil

	case "Pod":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Pod{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.clientset.CoreV1().Pods(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				_, err = k.clientset.CoreV1().Pods(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
				if err != nil {
					log.Printf("Failed to update pod: %v", err)
					return "", err
				}
			} else {
				log.Printf("Failed to create pod: %v", err)
				return "", err
			}
		}

		return fmt.Sprintf("Successfully applied %s %s", kind, obj.Name), nil
	}

	return "", errors.New("unsupported kind")
}

func (k *kubeController) DeleteYaml(yamlString string) (string, error) {
	kind, err := getYamlKind(yamlString)
	if err != nil {
		log.Printf("Failed to get yaml kind: %v", err)
		return "", err
	}

	jsonFile, err := yamlToJson(yamlString)
	if err != nil {
		log.Printf("Failed to convert yaml to json: %v", err)
		return "", err
	}

	switch kind {
	case "Deployment":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &appv1.Deployment{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		log.Printf("Deleting deployment: %s, %s", obj.Name, obj.Namespace)

		err = k.clientset.AppsV1().Deployments(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete deployment: %v", err)
			return "", err
		}

		return fmt.Sprintf("Successfully deleted %s %s", kind, obj.Name), nil

	case "Service":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Service{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		err = k.clientset.CoreV1().Services(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete service: %v", err)
			return "", err
		}

		return fmt.Sprintf("Successfully deleted %s %s", kind, obj.Name), nil

	case "Pod":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Pod{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			log.Printf("Failed to decode yaml: %v", err)
			return "", err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		err = k.clientset.CoreV1().Pods(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete pod: %v", err)
			return "", err
		}

		return fmt.Sprintf("Successfully deleted %s %s", kind, obj.Name), nil
	}

	return "", errors.New("unsupported kind")
}
