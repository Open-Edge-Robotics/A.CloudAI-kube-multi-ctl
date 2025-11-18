package controller

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubeyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"gopkg.in/yaml.v3"
)

type KubeController struct {
	Clientset *kubernetes.Clientset
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
			slog.Error("Failed to get k8s config: " + err.Error())
			return nil, err
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			slog.Error("Failed to create k8s client: %v" + err.Error())
			return nil, err
		}
	}
	return clientset, nil
}

func NewKubeController(path *string) (*KubeController, error) {
	clientset, err := GetKubeClient(path)
	if err != nil {
		slog.Error("Failed to create k8s client: %v" + err.Error())
		return nil, err
	}

	return &KubeController{
		Clientset: clientset,
	}, nil
}

func (k *KubeController) GetNodes() (*corev1.NodeList, error) {
	nodes, err := k.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		slog.Error("Failed to list nodes: %v" + err.Error())
		return nil, err
	}

	return nodes, nil
}

func (k *KubeController) GetNode(name string) (*corev1.Node, error) {
	node, err := k.Clientset.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to get node: %v" + err.Error())
		return &corev1.Node{}, err
	}

	return node, nil
}

func (k *KubeController) GetPods(namespace *string) (*corev1.PodList, error) {
	pods, err := k.Clientset.CoreV1().Pods(*namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		slog.Error("Failed to list pods: %v" + err.Error())
		return nil, err
	}

	return pods, nil
}

func (k *KubeController) GetPod(namespace, name string) (*corev1.Pod, error) {
	pod, err := k.Clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to get pod: %v" + err.Error())
		return &corev1.Pod{}, err
	}

	return pod, nil
}

func (k *KubeController) GetPodLogs(namespace, name string) (*string, error) {
	podLogOptions := corev1.PodLogOptions{}
	req := k.Clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOptions)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		slog.Error("Failed to get pod logs: %v" + err.Error())
		return nil, err
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		slog.Error("Failed to copy pod logs: %v" + err.Error())
		return nil, err
	}

	result := buf.String()

	return &result, nil
}

func getYamlKind(yamlString string) (*string, error) {
	yamlDecoder := yaml.NewDecoder(strings.NewReader(yamlString))
	yamlContent := make(map[string]interface{})
	err := yamlDecoder.Decode(&yamlContent)
	if err != nil {
		slog.Error("Failed to decode yaml: %v" + err.Error())
		return nil, err
	}

	result := yamlContent["kind"].(string)

	return &result, nil
}

func yamlToJson(yamlString string) ([]byte, error) {
	yamlDecoder := yaml.NewDecoder(strings.NewReader(yamlString))
	yamlContent := make(map[string]interface{})
	err := yamlDecoder.Decode(&yamlContent)
	if err != nil {
		slog.Error("Failed to decode yaml: %v" + err.Error())
		return nil, err
	}

	jsonFile, err := json.Marshal(yamlContent)
	if err != nil {
		slog.Error("Failed to marshal yaml: %v" + err.Error())
		return nil, err
	}

	return jsonFile, nil
}

func (k *KubeController) ApplyYaml(yamlString string) (*string, error) {
	kind, err := getYamlKind(yamlString)
	if err != nil {
		slog.Error("Failed to get yaml kind: %v" + err.Error())
		return nil, err
	}

	jsonFile, err := yamlToJson(yamlString)
	if err != nil {
		slog.Error("Failed to convert yaml to json: %v" + err.Error())
		return nil, err
	}

	switch *kind {
	case "Deployment":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &appv1.Deployment{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.Clientset.AppsV1().Deployments(obj.Namespace).Get(context.TODO(), obj.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				_, err = k.Clientset.AppsV1().Deployments(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
				if err != nil {
					slog.Error("Failed to create deployment: %v" + err.Error())
					return nil, err
				}
				createResult := fmt.Sprintf("Successfully deployment applied %s %s", *kind, obj.Name)
				return &createResult, nil
			}
			return nil, err
		}

		_, err = k.Clientset.AppsV1().Deployments(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("Failed to update deployment: %v" + err.Error())
			return nil, err
		}
		updateResult := fmt.Sprintf("Successfully deployment updated %s %s", *kind, obj.Name)
		return &updateResult, nil

	case "Service":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Service{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.Clientset.CoreV1().Services(obj.Namespace).Get(context.TODO(), obj.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				_, err = k.Clientset.CoreV1().Services(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
				if err != nil {
					slog.Error("Failed to create service: %v" + err.Error())
					return nil, err
				}
				createResult := fmt.Sprintf("Successfully service applied %s %s", *kind, obj.Name)
				return &createResult, nil
			}
			return nil, err
		}

		_, err = k.Clientset.CoreV1().Services(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("Failed to update service: %v" + err.Error())
			return nil, err
		}
		updateResult := fmt.Sprintf("Successfully service applied %s %s", *kind, obj.Name)
		return &updateResult, nil

	case "Pod":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Pod{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		_, err = k.Clientset.CoreV1().Pods(obj.Namespace).Get(context.TODO(), obj.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				_, err = k.Clientset.CoreV1().Pods(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
				if err != nil {
					slog.Error("Failed to create pod: %v" + err.Error())
					return nil, err
				}
				createResult := fmt.Sprintf("Successfully pod applied %s %s", *kind, obj.Name)
				return &createResult, nil
			}
			return nil, err
		}

		_, err = k.Clientset.CoreV1().Pods(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("Failed to update pod: %v" + err.Error())
			return nil, err
		}
		updateResult := fmt.Sprintf("Successfully pod applied %s %s", *kind, obj.Name)
		return &updateResult, nil
	}

	return nil, err
}

func (k *KubeController) DeleteYaml(yamlString string) (*string, error) {
	kind, err := getYamlKind(yamlString)
	if err != nil {
		slog.Error("Failed to get yaml kind: %v" + err.Error())
		return nil, err
	}

	jsonFile, err := yamlToJson(yamlString)
	if err != nil {
		slog.Error("Failed to convert yaml to json: %v" + err.Error())
		return nil, err
	}

	switch *kind {
	case "Deployment":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &appv1.Deployment{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		err = k.Clientset.AppsV1().Deployments(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("Failed to delete deployment: %v" + err.Error())
			return nil, err
		}

		result := fmt.Sprintf("Successfully deployment deleted %s %s", *kind, obj.Name)
		return &result, nil

	case "Service":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Service{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		err = k.Clientset.CoreV1().Services(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("Failed to delete service: %v" + err.Error())
			return nil, err
		}

		result := fmt.Sprintf("Successfully service deleted %s %s", *kind, obj.Name)

		return &result, nil

	case "Pod":
		decode := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &corev1.Pod{}
		_, _, err := decode.Decode(jsonFile, nil, obj)
		if err != nil {
			slog.Error("Failed to decode yaml: %v" + err.Error())
			return nil, err
		}

		if obj.Namespace == "" {
			obj.Namespace = "default"
		}

		err = k.Clientset.CoreV1().Pods(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("Failed to delete pod: %v" + err.Error())
			return nil, err
		}

		result := fmt.Sprintf("Successfully pod deleted %s %s", *kind, obj.Name)

		return &result, nil
	}

	return nil, err
}
