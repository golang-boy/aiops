package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	// kubeconfig := flag.String("kubeconfig", "/root/.kube/config", "location of kubeconfig file")
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	fmt.Printf("error %s", err.Error())
	// }

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 可以看一下 Pods 里面有什么操作
	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
	for _, pod := range pods.Items {
		fmt.Printf("Pod name %s\n", pod.Name)
	}

	fmt.Println("List deployments")

	deployment, err := clientset.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	for _, d := range deployment.Items {
		fmt.Printf("deployment name %s\n", d.Name)
	}

	fmt.Println("List services")

	services, err := clientset.CoreV1().Services("default").List(context.Background(), metav1.ListOptions{})
	for _, s := range services.Items {
		fmt.Printf("service name %s\n", s.Name)
	}
}
