# 第七周

operator 实践

operator = controller + CRD


## 概念解释

operator 是什么?
 - operator 是一种扩展 kubernetes API 的方法，它使用自定义资源定义（CRD）来扩展 API，并使用控制器来确保自定义资源的状态始终与用户的期望相匹配。

有什么好处?

 - 利用controller控制循环, 不用自己写那一坨
 - 利用k8s的资源管理能力, 与k8s原生集成
 - 简化有状态应用的开发流程



## 实践一

    使用kubebuilder创建一个operator


 流程：

 1. 安装kubebuilder
```
# download kubebuilder and install locally.
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```


 2. 初始化一个项目

```
(robot3) root@localhost:application(main %=) $ go mod init app
go: creating new go.mod: module app
(robot3) root@localhost:application(main %=) $ kubebuilder init --domain=aiops.org
(robot3) root@localhost:application(main %=) $ CGO_ENABLED=0 kubebuilder create api --group app --version v1 --kind Application

```

3. 编辑api目录下的application_types.go文件，定义crd，执行make manifests生成crd文件

4. 在internal/controllers/application_controller.go文件中，实现Reconcile方法，实现自定义资源的状态管理

5. 资源创建完毕后，需要更新资源状态，因此需要定义状态字段，失败时，可以再次放入队列进行重试，成功时，可以删除队列中的资源

   每一项资源启动后，需要设置引用关系```SetControllerReference```，这样删除自定义资源时，可以都清理掉。通过这种方式就不用写删除逻辑了

6. 编写完毕后，执行make install，安装crd，执行make run，启动operator

7. 在k8s中创建资源，执行kubectl get application，可以看到资源已经创建成功

```
(robot3) root@localhost:application(main %=) $ kubectl get application
NAME                 AGE
application-sample   62s
(robot3) root@localhost:application(main %=) $ kubectl get deploy     
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
application-sample   0/1     1            0           71s
demo-1               0/1     1            0           3d23h
nginx-deployment     1/1     1            1           3d21h
test-deployment      0/1     1            0           21h
(robot3) root@localhost:application(main %=) $ kubectl get service
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
application-sample   ClusterIP   10.96.165.239   <none>        8080/TCP       87s
kubernetes           ClusterIP   10.96.0.1       <none>        443/TCP        3d23h
nginx-service        NodePort    10.96.110.235   <none>        80:30080/TCP   3d23h
(robot3) root@localhost:application(main %=) $ kubectl get ingress
NAME                 CLASS   HOSTS             ADDRESS   PORTS   AGE
application-sample   nginx   example.foo.com             80      119s
(robot3) root@localhost:application(main %=) $ kubectl get pods   
NAME                                  READY   STATUS             RESTARTS         AGE
application-sample-67776846b4-vjvkb   0/1     ImagePullBackOff   0                2m14s
demo-1-99979696-54kvp                 0/1     CrashLoopBackOff   26 (2m29s ago)   110m
error-logging-pod                     0/1     CrashLoopBackOff   22 (5m2s ago)    92m
nginx-deployment-7d9745ffbd-96pcm     1/1     Running            0                3d21h
redis                                 0/1     ImagePullBackOff   0                111m
test-deployment-67b84cd4c6-7nvpp      0/1     ErrImagePull       0                21h
```

### 目录结构解释
(robot3) root@localhost:application(main=) $ tree -L 2
.
|-- Dockerfile
|-- Makefile
|-- PROJECT
|-- README.md
|-- api                                   // api资源定义
|   `-- v1
|-- bin
|   |-- controller-gen -> /home/gogo/aiops/Week07/application/bin/controller-gen-v0.16.4
|   |-- controller-gen-v0.16.4
|   |-- kustomize -> /home/gogo/aiops/Week07/application/bin/kustomize-v5.4.3
|   `-- kustomize-v5.4.3
|-- cmd                                   // 应用程序入口
|   `-- main.go
|-- config                                // crd对象，生成的样例在这里
|   |-- crd
|   |-- default
|   |-- manager
|   |-- network-policy
|   |-- prometheus
|   |-- rbac
|   `-- samples
|-- go.mod
|-- go.sum
|-- hack
|   `-- boilerplate.go.txt
|-- internal
|   `-- controller                      // 控制器
`-- test
    |-- e2e
    `-- utils


## 实践二

    试试定时弹性伸缩, 把nginx的工作负载通过hpa在某个时间点进行扩缩容

    定时弹性伸缩是什么？

    根据预定的时间表自动增加或减少资源的使用量。例如，在业务高峰期自动增加资源，而在业务低谷期减少资源，以优化资源利用率和降低成本。这种功能特别适用于那些业务量有明显时间周期性变化的应用场景‌


    ```
    spec:
      scaleTarget:
        apiVersion: apps/v1
        kind: Deployment
        name: nginx
      jobs:
        - name: "scale-up"
          schedule: "*/1 * * * *"   // 每分钟扩一下
          size: 3             // 扩到3个
    ```




1. 建个operator项目
```
 go mod init hpa
 kubebuilder init --domain=aiops.org
 CGO_ENABLED=0 kubebuilder create api --group autoscal --version v1 --kind Hpa
```


2. 编辑api/v1/hpa_types.go




