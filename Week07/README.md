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
```
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
```


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


2. 编辑api/v1/hpa_types.go, 增加spec字段和status相关字段
```
type HpaSpec struct {
	ScaleTarget ScaleTarget `json:"scaleTarget"`
	JobSpec     []JobSpec   `json:"jobs"`
}

type JobSpec struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Size     int    `json:"size"`
}

type ScaleTarget struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// 另外添加一些注解，方便命令行查看

// +kubebuilder:printcolumn:name="Target",type="string",JSONPath=".spec.scaleTarget.name",description="目标工作负载"
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.jobs[*].schedule",description="Cron 表达式"
// +kubebuilder:printcolumn:name="Target Size",type="integer",JSONPath=".spec.jobs[*].size",description="目标副本数"

type Hpa struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HpaSpec   `json:"spec,omitempty"`
	Status HpaStatus `json:"status,omitempty"`
}


make manifests  // 生成crd文件
```

增加注解后，命令行执行效果如下：
```
(robot3) root@localhost:hpa(main *%=) $ kubectl get hpa
NAME         TARGET             SCHEDULE      TARGET SIZE
hpa-sample   nginx-deployment   */1 * * * *   3
```

3. 编辑controller逻辑

   逻辑大体流程为：
     1. 从命名空间中获取hpa的crd资源对象
     2. 循环处理hpa对象中定义的jobs，检测调度时间是否到达
     3. 如果到达，则根据hpa对象中定义的scaleTarget和size字段, 以及状态中的信息, 对目标工作负载进行扩缩容操作
     4. 更新hpa对象的状态，记录扩缩容操作的结果

4. 启动

   1. make install 安装crd的base下的资源文件
   2. 运行operator, make run
   3. 修改config/samples/autoscal_v1_hpa.yaml, 添加配置项

```
(robot3) root@localhost:hpa(main *%=) $ kubectl get deploy
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
application-sample   0/1     1            0           9d
demo-1               0/1     1            0           13d
nginx-deployment     1/1     1            1           13d
test-deployment      0/1     1            0           10d
```

5. kubectl apply -f config/samples/autoscal_v1_hpa.yaml 部署

环境中已有nginx-deployment, 部署hpa后，每分钟扩容一次，副本数将变为3
```shell
(robot3) root@localhost:hpa(main *%=) $ kubectl apply -f config/samples/autoscal_v1_hpa.yaml
hpa.autoscal.aiops.org/hpa-sample created
(robot3) root@localhost:hpa(main *%=) $ 
(robot3) root@localhost:hpa(main *%=) $ kubectl get deploy 
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
application-sample   0/1     1            0           9d
demo-1               0/1     1            0           13d
nginx-deployment     3/3     3            3           13d
```


 6. 手动缩容后，查看是否还会加上来
```
(robot3) root@localhost:hpa(main *%=) $ kubectl scale deployment nginx-deployment --replicas=1
deployment.apps/nginx-deployment scaled
(robot3) root@localhost:hpa(main *%=) $ 
(robot3) root@localhost:hpa(main *%=) $ kubectl get deploy
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
application-sample   0/1     1            0           9d
demo-1               0/1     1            0           13d
nginx-deployment     1/1     1            1           13d
test-deployment      0/1     1            0           10d
(robot3) root@localhost:hpa(main *%=) $ kubectl get deploy --watch
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
application-sample   0/1     1            0           9d
demo-1               0/1     1            0           13d
nginx-deployment     3/3     3            3           13d
test-deployment      0/1     1            0           10d
```

  
## 实践三
  
    怎么打包这个operator呢？

  以实践二为例

1. 构建镜像

```
      export IMG=<some-registry>/<project-name>:tag  // 如果需要推送，则设置该环境变量

      make docker-build 
```

2. 构建安装资源文件, 在dist目录下生成该文件，可以使用其在任何k8s集群中安装自己的operator

```
   make build-installer
```




## 实践四


    试试operator-sdk

  可以通过helm chart, 生成operator， 会自动引用helm chart中的values.yaml的内容，作为crd的spec
  当crd变化时，自动变更，生成新的重新安装


1. 安装operator-sdk

https://sdk.operatorframework.io/docs/installation/

```
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.37.0
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

上述略过了包校验

2. 做个部署nginx的operator

```shell
(robot3) root@localhost:Week07(main *=) $ mkdir nginx
(robot3) root@localhost:Week07(main *=) $ cd nginx/
(robot3) root@localhost:nginx(main *=) $ go mod init nginx-operator
go: creating new go.mod: module nginx-operator
(robot3) root@localhost:nginx(main *%=) $ operator-sdk init --domain aiops.org --plugins=helm 
INFO[0000] Writing kustomize manifests for you to edit... 
Next: define a resource with:
$ operator-sdk create api
(robot3) root@localhost:nginx(main *%=) $ ls
Dockerfile  Makefile  PROJECT  config  go.mod  helm-charts  watches.yaml
```

3. 创建一个CRD

指定helm chart的repo地址和chart名称
```
(robot3) root@localhost:nginx(main *%=) $  operator-sdk create api --group web --version v1 --kind Nginx --helm-chart-repo https://charts.bitnami.com/bitnami --helm-chart nginx 

INFO[0011] Writing kustomize manifests for you to edit... 
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x38 pc=0x20f6d8b]

goroutine 1 [running]:
helm.sh/helm/v3/pkg/registry.(*Client).Tags(0x0, {0xc005efa570?, 0xc0000a9210?})
        /home/runner/go/pkg/mod/helm.sh/helm/v3@v3.14.3/pkg/registry/client.go:671 +0x12b
helm.sh/helm/v3/internal/resolver.(*Resolver).Resolve(0xc0000a9458, {0xc004a7acc0, 0x1, 0x1}, 0xc005d1af30?)
        /home/runner/go/pkg/mod/helm.sh/helm/v3@v3.14.3/internal/resolver/resolver.go:153 +0x588
helm.sh/helm/v3/pkg/downloader.(*Manager).resolve(0xc005d1aa20?, {0xc004a7acc0?, 0xc004a7acc0?, 0x1?}, 0x0?)
        /home/runner/go/pkg/mod/helm.sh/helm/v3@v3.14.3/pkg/downloader/manager.go:235 +0x5f
helm.sh/helm/v3/pkg/downloader.(*Manager).Update(0xc005d1aa20)
        /home/runner/go/pkg/mod/helm.sh/helm/v3@v3.14.3/pkg/downloader/manager.go:194 +0xdc
helm.sh/helm/v3/pkg/downloader.(*Manager).Build(0xc005d1aa20)
        /home/runner/go/pkg/mod/helm.sh/helm/v3@v3.14.3/pkg/downloader/manager.go:95 +0x3d2
```

报了个错误, 这应该是helm的依赖问题

```
(robot3) root@localhost:nginx(main *%=) $ helm repo add bitnami https://charts.bitnami.com/bitnami
"bitnami" has been added to your repositories
(robot3) root@localhost:nginx(main *%=) $ helm dependencies build 
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "stable" chart repository
...Successfully got an update from the "bitnami" chart repository
Update Complete. ⎈Happy Helming!⎈
Saving 1 charts
Downloading common from repo https://charts.bitnami.com/bitnami
Deleting outdated charts
```

依赖下载完毕后，删除掉之前创建的CRD，重新来一次
```
FATA[0000] failed to create API: unable to scaffold with "kustomize.common.kubebuilder.io/v2": error scaffolding kustomize API manifests: failed to create config/samples/web_v1_nginx.yaml: file already exists 

(robot3) root@localhost:nginx(main *%=) $ operator-sdk create api --group web --version v1 --kind Nginx  --helm-chart ./helm-charts/nginx
INFO[0000] Writing kustomize manifests for you to edit... 
Created helm-charts/nginx
Generating RBAC rules
WARN[0020] Skipping rule generation for manifest-1. Failed to determine resource scope for policy/v1beta1, Kind=PodDisruptionBudget. 
WARN[0020] The RBAC rules generated in config/rbac/role.yaml are based on the chart's default manifest. Some rules may be missing for resources that are only enabled with custom values, and some existing rules may be overly broad. Double check the rules generated in config/rbac/role.yaml to ensure they meet the operator's permission requirements. 
```


```
(robot3) root@localhost:nginx(main *%=) $ make docker-build         // 此处可以通过IMG定义镜像名称
...
Successfully built e500a19a9337
Successfully tagged controller:latest
```

部署operator

```
kind load docker-image controller:latest   // 加入kind
make deploy IMG="controller:latest"
```

4. 打包 operator

安装 olm(在集群中安装一下负载)

```
  operator-sdk olm install
  operator-sdk run bundle  $BUNDLE_IMG
```


## 总结operator

 * 开发时最好设计成一个幂等的operator ，即重复执行多次，结果一致
 * 只需关注期望状态和当前状态的差异，执行业务逻辑