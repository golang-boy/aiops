/*
 * @Author: 刘慧东
 * @Date: 2024-11-04 18:29:57
 * @LastEditors: 刘慧东
 * @LastEditTime: 2024-11-04 19:11:47
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
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

	// 创建速率限制队列
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())
	// 对 Deployment 监听
	deployInformer := informerFactory.Apps().V1().Deployments()
	informer := deployInformer.Informer()
	c := NewController(queue, deployInformer.Informer().GetIndexer(), informer)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.onAddDeployment(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.onUpdateDeployment(new)
		},
		DeleteFunc: func(obj interface{}) {
			c.onDeleteDeployment(obj)
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informerFactory.Start(stopper)
	informerFactory.WaitForCacheSync(stopper)

	// 处理队列中的事件
	go func() {
		for {
			if !c.process() {
				break
			}
		}
	}()

	<-stopper
}

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.TypedRateLimitingInterface[string]
	informer cache.Controller
}

func NewController(queue workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) process() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.print(key)
	c.handleErr(err, key)
	return true
}

// 输出日志
func (c *Controller) print(key string) error {
	// 通过 key 从 indexer 中获取完整的对象
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		fmt.Printf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		fmt.Printf("Deployment %s does not exist anymore\n", key)
	} else {
		deployment := obj.(*v1.Deployment)
		fmt.Printf("Sync/Add/Update for Deployment %s, Replicas: %d\n", deployment.Name, *deployment.Spec.Replicas)
		if deployment.Name == "test-deployment" {
			time.Sleep(2 * time.Second)
			return fmt.Errorf("simulated error for deployment %s", deployment.Name)
		}
	}
	return nil
}

func (c *Controller) onAddDeployment(obj interface{}) {
	// 生成 key
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *Controller) onUpdateDeployment(new interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err == nil {
		c.queue.Add(key)
	}

}

func (c *Controller) onDeleteDeployment(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *Controller) handleErr(err error, key string) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		log.Printf("Retry %d for key %s\n", c.queue.NumRequeues(key), key)
		// 重新加入队列，并且进行速率限制，这会让他过一段时间才会被处理，避免过度重试
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	log.Printf("Dropping deployment %q out of the queue: %v\n", key, err)
}
