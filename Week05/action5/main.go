/*
 * @Author: 刘慧东
 * @Date: 2024-11-04 18:29:57
 * @LastEditors: 刘慧东
 * @LastEditTime: 2024-11-04 18:42:04
 */
package main

import (
	"flag"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "/root/.kube/config", "location of kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	// 初始化 rest.Config 对象
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 初始化 informer
	// 多个资源共享这个连接
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour*12)

	// 对 Deployment 监听
	deployInformer := informerFactory.Apps().V1().Deployments()
	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAddDeployment,
		UpdateFunc: onUpdateDeployment,
		DeleteFunc: onDeleteDeployment,
	})

	// 对 Service 监听
	serviceInformer := informerFactory.Core().V1().Services()
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAddService,
		UpdateFunc: onUpdateService,
		DeleteFunc: onDeleteService,
	})

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informerFactory.Start(stopper)
	// 等待所有启动的 Informer 的缓存被同步
	informerFactory.WaitForCacheSync(stopper)

	// Lister 实际上就是本地缓存，他从 Indexer 里取数据
	deployLister := deployInformer.Lister()
	// Lister，从本地缓存中获取 default 中的所有 deployment 列表，最终从 Indexer 取数据
	deployments, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for idx, deploy := range deployments {
		fmt.Printf("%d -> %s\n", idx+1, deploy.Name)
	}

	fmt.Println("========================================")
	serviceLister := serviceInformer.Lister()
	services, err := serviceLister.Services("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for idx, service := range services {
		fmt.Printf("%d -> %s\n", idx+1, service.Name)
	}

	// 阻塞主 goroutine
	<-stopper

}

func onAddDeployment(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("add a deployment:", deploy.Name)
}

func onUpdateDeployment(old, new interface{}) {
	oldDeploy := old.(*v1.Deployment)
	newDeploy := new.(*v1.Deployment)
	fmt.Println("update deployment:", oldDeploy.Name, newDeploy.Name)
}

func onDeleteDeployment(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("delete a deployment:", deploy.Name)
}

func onAddService(obj interface{}) {
	service := obj.(*v1core.Service)
	fmt.Println("add a service:", service.Name)
}

func onUpdateService(old, new interface{}) {
	oldService := old.(*v1core.Service)
	newService := new.(*v1core.Service)
	fmt.Println("update service:", oldService.Name, newService.Name)
}

func onDeleteService(obj interface{}) {
	service := obj.(*v1core.Service)
	fmt.Println("delete a service:", service.Name)
}
