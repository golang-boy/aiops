/*
 * @Author: 刘慧东
 * @Date: 2024-11-01 17:47:27
 * @LastEditors: 刘慧东
 * @LastEditTime: 2024-11-01 17:59:48
 */
package main

import (
	"context"
	"flag"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "/root/.kube/config", "location of kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}

	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	fmt.Printf("error %s", err.Error())
	// }

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	rest, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err.Error())
	}

	// 创建空的结构体，存储pod列表
	podList := &corev1.PodList{}

	rest.Get().
		// 指定命名空间
		Namespace("default").
		// 指定要获取的资源类型
		Resource("pods").
		// 设置请求参数，使用metav1.ListOptions结构体设置了Limit参数为500，并使用scheme.ParameterCodec进行参数编码。
		// 限制返回的响应将包含最多500个pod对象。
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
		// 发送请求并获取响应，使用context.TODO()作为上下文
		Do(context.TODO()).
		// 将响应解码为podList
		Into(podList)

	for _, v := range podList.Items {
		fmt.Printf("NameSpace: %v  Name: %v  Status: %v \n", v.Namespace, v.Name, v.Status.Phase)
	}

}
