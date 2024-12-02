# 第五周
---

这周学习client-go


## 实践一

###  1.通过本地配置文件获取k8s信息
```
	kubeconfig := flag.String("kubeconfig", "/root/.kube/config", "location of kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
```

    使用client-go获取一下k8s中默认的pods（官方示例实践）

    [代码在这里](action1/main.go)

    先部署一个nginx服务，client-go中通过get pods拿到信息


    ```shell
    (robot3) root@localhost:action1(main *%=) $ ./action1 
    Pod name nginx-deployment-57d84f57dc-6khmf
    Pod name nginx-deployment-57d84f57dc-t2nsj
    List deployments
    deployment name nginx-deployment
    List services
    service name kubernetes
    service name nginx-service
    ```

    ```
    (robot3) root@localhost:action1(main *%=) $ kubectl get pods
    NAME                                READY   STATUS             RESTARTS   AGE
    nginx-deployment-57d84f57dc-6khmf   0/1     ImagePullBackOff   0          9m28s
    nginx-deployment-57d84f57dc-t2nsj   0/1     ImagePullBackOff   0          9m28s
    (robot3) root@localhost:action1(main *%=) $ 
    (robot3) root@localhost:action1(main *%=) $ kubectl get deployments
    NAME               READY   UP-TO-DATE   AVAILABLE   AGE
    nginx-deployment   0/2     2            0           10m
    (robot3) root@localhost:action1(main *%=) $ kubectl get services   
    NAME            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
    kubernetes      ClusterIP   10.96.0.1     <none>        443/TCP        38d
    nginx-service   NodePort    10.96.87.96   <none>        80:30080/TCP   10m
    ```


名字都能对得上



### 2.通过InclusterConfig函数获取k8s api的访问token,


client-go的代码是需要放到k8s中运行的，因此需要配置role，和rolebinding，绑定到默认的serviceaccount上，serviceaccount中包含token，通过token访问api，因此需要通过InclusterConfig函数获取k8s api的访问token

```
make images
kind load docker-image demo1  #加载到kind集群
kubectl apply -f deployment.yaml
```

```
(robot3) root@localhost:action1(main *%=) $ kubectl get pods 
NAME                                READY   STATUS             RESTARTS      AGE
demo-1-55b9456855-2qs9c             0/1     CrashLoopBackOff   5 (36s ago)   3m33s
nginx-deployment-57d84f57dc-245jv   0/1     ImagePullBackOff   0             14m
nginx-deployment-57d84f57dc-hpwgc   0/1     ErrImagePull       0             14m
(robot3) root@localhost:action1(main *%=) $ kubectl  logs    demo-1-55b9456855-2qs9c
error pods is forbidden: User "system:serviceaccount:default:default" cannot list resource "pods" in API group "" in the namespace "default"List deployments
List services
```

没权限，需要给权限

```
kubectl create role demo --resource pods,deployment --verb list

kubectl create  rolebinding  demo  --role demo --serviceaccount default:default
```

* kubectl create role: 创建一个新的 Role。
* demo: 这是要创建的 Role 的名称。
* --resource pods,deployment: 指定该 Role 授权的资源类型。在这里，它允许访问 pods 和 deployments。
* --verb list: 指定允许的操作。在这个例子中，允许执行 list 操作，即列出指定资源的权限。


* kubectl create rolebinding: 创建一个新的 RoleBinding。
* demo: 这是要创建的 RoleBinding 的名称。
* --role demo: 指定要绑定的 Role 的名称，在这里是 demo。
* --serviceaccount default:default: 指定要绑定 Role 的服务账户。default:default 指的是在 default 命名空间中名为 default 的服务账户。


```
(robot3) root@localhost:action1(main *%=) $ kubectl get pods
NAME                                READY   STATUS             RESTARTS      AGE
demo-1-55b9456855-f9zfr             0/1     Completed          2 (21s ago)   25s
nginx-deployment-57d84f57dc-245jv   0/1     ImagePullBackOff   0             27m
nginx-deployment-57d84f57dc-hpwgc   0/1     ImagePullBackOff   0             27m
(robot3) root@localhost:action1(main *%=) $ kubectl logs demo-1-55b9456855-f9zfr
Pod name demo-1-55b9456855-f9zfr
Pod name nginx-deployment-57d84f57dc-245jv
Pod name nginx-deployment-57d84f57dc-hpwgc
List deployments
deployment name demo-1
deployment name nginx-deployment
List services
```

yeah, 拿到自己想要的结果,但是service是没显示出来，角色中新增services资源


```shell
(robot3) root@localhost:action1(main *%=) $ kubectl delete role demo
role.rbac.authorization.k8s.io "demo" deleted
(robot3) root@localhost:action1(main *%=) $ kubectl create role demo --resource pods,deployment,services --verb list
role.rbac.authorization.k8s.io/demo created
(robot3) root@localhost:action1(main *%=) $ kubectl rollout restart deployment demo-1        
deployment.apps/demo-1 restarted
(robot3) root@localhost:action1(main *%=) $ kubectl logs demo-1-99979696-xlvtw  
Pod name demo-1-99979696-xlvtw
Pod name nginx-deployment-57d84f57dc-245jv
Pod name nginx-deployment-57d84f57dc-hpwgc
List deployments
deployment name demo-1
deployment name nginx-deployment
List services
service name kubernetes
service name nginx-service
```

### 3.service account是什么? 

允许 Pod 在访问 Kubernetes API 时以特定的身份进行身份验证和授权, 是一种资源。
创建namespace时，会创建一个默认的service account，并且会自动创建一个对应的secret，这个secret中包含了访问api的token。这些会默认给所有pod注入service account配置文件

通过rolebinding，将role绑定到service account，从而给service account赋予访问k8s api的权限

role定义能访问的资源

```
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
```


**clientset是最常用的client,实现了k8s标准对象接口, restClient是底层接口，可以自定义请求(不方便,代码在action2目录)**


### 4. 试试创建一个deployment

删除旧的，执行程序
```
(robot3) root@localhost:action3(main *%=) $ kubectl delete  deployments nginx-deployment
deployment.apps "nginx-deployment" deleted
(robot3) root@localhost:action3(main *%=) $ ./action3 
Created deployment "nginx-deployment".
(robot3) root@localhost:action3(main *%=) $ kubectl get deployments
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
demo-1             0/1     1            0           72m
nginx-deployment   0/1     1            0           3s
```

这个需要镜像

```
kind load docker-image nginx:latest
(robot3) root@localhost:action3(main *%=) $ kubectl delete deployment nginx-deployment
deployment.apps "nginx-deployment" deleted
(robot3) root@localhost:action3(main *%=) $ ./action3 
Created deployment "nginx-deployment".
(robot3) root@localhost:action3(main *%=) $ kubectl get pods --watch
NAME                                READY   STATUS             RESTARTS         AGE
demo-1-99979696-xlvtw               0/1     CrashLoopBackOff   13 (2m50s ago)   44m
nginx-deployment-7d9745ffbd-96pcm   1/1     Running            0                4s
```

### 5. dynamicClient 是什么?

 [代码在这里](./action4/main.go)

  通过定义gvr,获取k8s资源, gvr是group,version,resource的缩写。然后将返回的非结构化数据转换为对应的结构化数据 (里面使用的是反射)

  写好的资源配置文件，代码中转为unstructured.Unstructured对象，根据它提供的一些函数进行元数据处理，调用动态客户端，生成资源

## 实践二

    使用 Informer + RateLimitingQueue 监听 Pod 事件； 
    [代码在这里](./action6/main.go)

    生产中不推荐使用watch, 原因：
    1. watch直连api server, api server压力很大
    2. 不同资源的watch,api server需要维护多个连接，资源多，连接多
    3. 断开重连，会丢失事件
    4. 没有控制，事件会源源不断来, 需要缓存一下
    5. watch到的事件，处理失败后，没法再来一次

    因此，需要使用informer机制来处理系统事件


```
(robot3) root@localhost:action6(main *%>) $ kubectl create deployment test-deployment --image nginx:latest
deployment.apps/test-deployment created
(robot3) root@localhost:action6(main *%>) $ 
```

controller重启后，收到deployment为test-deployment的相关事件，触发失败，失败重新入队重试，重试5次后，删除队列中的deployment，不再重试
```
(robot3) root@localhost:action6(main *%>) $ ./action 
Sync/Add/Update for Deployment demo-1, Replicas: 1
Sync/Add/Update for Deployment nginx-deployment, Replicas: 1
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:16 Retry 0 for key default/test-deployment
Sync/Add/Update for Deployment coredns, Replicas: 2
Sync/Add/Update for Deployment local-path-provisioner, Replicas: 1
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:18 Retry 1 for key default/test-deployment
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:20 Retry 2 for key default/test-deployment
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:22 Retry 3 for key default/test-deployment
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:24 Retry 4 for key default/test-deployment
Sync/Add/Update for Deployment test-deployment, Replicas: 1
2024/11/04 19:13:26 Dropping deployment "default/test-deployment" out of the queue: simulated error for deployment test-deployment
```


总结: 
1. list watch后，informer会收到变化的对象，然后放入indexer中缓存，
2. 通过cache中的函数计算得出key，然后放入用户队列
3. 控制循环controller从队列中取出key，根据key从cache中取出对象，进行业务处理
4. 处理失败，重新放入队列，这是个延迟队列，隔一段时间后，会重试


## 实践三

    创建一个新的自定义 CRD（Group：aiops.com, Version: v1alpha1, Kind: AIOps），并使用 dynamicClient 获取该资源。

流程：
1. 定义crd
```
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresources.aiops.org
spec:
  group: aiops.org
  versions:
    - name: v1alpha1           # 定义版本
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:         # 定义资源属性
                field1:
                  type: string
                  description: First example field
                field2:
                  type: string
                  description: Second example field
            status:
              type: object
  scope: Namespaced
  names:
    plural: myresources     # 资源名字, 命令行获取时输入的, 复数的
    singular: myresource    # 资源名字, 命令行获取时输入的, 单数的
    kind: AIOps             # 定义kind
    shortNames:
      - myres
```
2. 创建crd
```
(robot3) root@localhost:action7(main *%=) $ kubectl apply -f crd.yaml 
customresourcedefinition.apiextensions.k8s.io/myresources.aiops.org created
```

3. 定义并创建自己的crd对象

```
apiVersion: aiops.org/v1alpha1
kind: AIOps
metadata:
  name: my-first-crd
  namespace: default
spec:
  field1: "ExampleValue1"
  field2: "ExampleValue2"
```
```
(robot3) root@localhost:action7(main *%=) $ kubectl apply -f resource.yml 
aiops.aiops.org/my-first-crd created
```

```
(robot3) root@localhost:action7(main *%=) $ kubectl get myresources
NAME           AGE
my-first-crd   48s
(robot3) root@localhost:action7(main *%=) $ kubectl get myresource 
NAME           AGE
my-first-crd   52s
```


4. 通过dynamicclient获取自己的crd对象,(根据kind去获取)

```
(robot3) root@localhost:action7(main *%=) $ ./action get AIOps      
Name: my-first-crd, Namespace: default, UID: 5b33348c-3d5b-483f-a731-cc197eb5d587
(robot3) root@localhost:action7(main *%=) $ ./action get aiops
Name: my-first-crd, Namespace: default, UID: 5b33348c-3d5b-483f-a731-cc197eb5d587
```

总结：
 1. 定义并创建crd后，通过mapper对象，将gvk转换为gvr
 2. 有了gvr后，通过```dynamicClient.Resource(mapping.Resource).Namespace("default")```这种方式，获取到对应的resource对象,进而list拿到信息


### 概念解释

1.  gvk是什么？

    gvk是Group Version Kind的缩写，是Kubernetes API中的一种标识符，用于唯一标识一个资源类型。它由三个部分组成：
   
    - Group：资源的API组，例如apps、batch、core等。
    - Version：资源的API版本，例如v1、v1beta1等。
    - Kind：资源的类型，例如Pod、Service、Deployment等。

例如，在Kubernetes中，Pod资源的gvk为：

    - Group：core
    - Version：v1
    - Kind：Pod

2. gvr是什么？

    gvr是Group Version Resource的缩写，是Kubernetes API中的一种标识符，用于唯一标识一个资源实例。它由三个部分组成：

    - Group：资源的API组，例如apps、batch、core等。
    - Version：资源的API版本，例如v1、v1beta1等。
    - Resource：资源的名称，例如pods、services、deployments等。

例如，在Kubernetes中，一个名为my-pod的Pod资源的gvr为：

    - Group：core
    - Version：v1
    - Resource：pods
    - Name：my-pod

3. gvk和gvr的区别

  一个是资源的类型，一个是资源的实例,前者是抽象的，后者是具体的。

  
