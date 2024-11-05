/*
 * @Author: 刘慧东
 * @Date: 2024-11-04 17:42:05
 * @LastEditors: 刘慧东
 * @LastEditTime: 2024-11-04 17:53:46
 */

package main

import (
	"context"
	"flag"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/root/.kube/config", "location of kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	// 根据配置创建dynamicClient
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	gvr := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "pods",
	}
	unStructObj, err := dynamicClient.Resource(gvr).Namespace("kube-system").List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		panic(err)
	}

	podList := &corev1.PodList{}
	// 将unstructured对象转换为podList
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(unStructObj.UnstructuredContent(), podList); err != nil {
		panic(err)
	}

	fmt.Println(unStructObj)

	apiVersion, found, err := unstructured.NestedString(unStructObj.Object, "apiVersion")

	if err != nil || !found {
		panic(err)
	}

	fmt.Println(apiVersion)

	// 根据api

	for _, v := range podList.Items {
		fmt.Printf("namespace: %s, name: %s, status: %s\n", v.Namespace, v.Name, v.Status.Phase)
	}
}
