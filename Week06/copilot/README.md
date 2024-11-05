# 第六周

1. 使用go调用大模型
2. 通过go结合k8s和大模型


流程：

   通过命令行实现上述功能

1. 创建代码框架
   ```
   mkdir -p copilot
   go mod init copilot
   cobra init
   cobra add ask
   cobra add gpt -p "askCmd"
   ```

2. 编写代码，具体流程为

  1. 通过openai库，设置好functioncalling中的tools定义
  2. 创建openai的对话上下文, 大模型理解对话后，会调用对应的tool
  3. tool对应的执行函数，通过client-go调用k8s的api，执行对应的操作

```
(robot3) root@localhost:copilot(main %=) $ OPENAI_API_KEY="xxxx" ./copilot ask gpt
有什么可以帮助你：
>>>> hello
Completion error: len(toolcalls): 0

>>>> 删除资源redis，类型为pod
OpenAI called us back wanting to invoke our function 'deleteResource' with params '{"namespace":"default","resource_name":"redis","resource_type":"Pod"}'

>>>> 帮我部署deployment，镜像是nginx
Error calling function: deployments.apps "nginx-deployment" already exists

>>>> 帮我生成一个资源，镜像是redis
YAML content:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:latest
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  ports:
  - port: 6379
    protocol: TCP
  selector:
    app: redis
  type: ClusterIP

Deployment successful.


>>>> 帮我查询有哪些pods
Found pod: demo-1-99979696-xlvtw
Found pod: nginx-deployment-7d9745ffbd-96pcm
Found pod: redis-7968b47c9-zkgj9
Found pod: test-deployment-67b84cd4c6-7nvpp
Found pod: coredns-5d78c9869d-8hkxg
Found pod: coredns-5d78c9869d-qq8dz
Found pod: etcd-kind-control-plane
Found pod: kindnet-2bdlt
Found pod: kube-apiserver-kind-control-plane
Found pod: kube-controller-manager-kind-control-plane
Found pod: kube-proxy-25tfh
Found pod: kube-scheduler-kind-control-plane
Found pod: local-path-provisioner-6bc4bddd6b-5mp6h
```
 
```
>>>> 有哪些pod
debug:  OpenAI called us back wanting to invoke our function 'queryResource' with params '{"resource_type":"Pod"}'

Found pod: demo-1-99979696-xlvtw
Found pod: nginx-deployment-7d9745ffbd-96pcm
Found pod: redis-7968b47c9-8wxl5
Found pod: redis-7968b47c9-r5gkv
Found pod: test-deployment-67b84cd4c6-7nvpp
Found pod: coredns-5d78c9869d-8hkxg
Found pod: coredns-5d78c9869d-qq8dz
Found pod: etcd-kind-control-plane
Found pod: kindnet-2bdlt
Found pod: kube-apiserver-kind-control-plane
Found pod: kube-controller-manager-kind-control-plane
Found pod: kube-proxy-25tfh
Found pod: kube-scheduler-kind-control-plane
Found pod: local-path-provisioner-6bc4bddd6b-5mp6h

>>>> 删除redis的deployment
debug:  OpenAI called us back wanting to invoke our function 'deleteResource' with params '{"namespace":"default","resource_name":"redis","resource_type":"Deployment"}'

Deleting resource: redis  deployment
Resource deleted successfully
>>>> 有哪些pod
debug:  OpenAI called us back wanting to invoke our function 'queryResource' with params '{"namespace":"default"}'

Error calling function: unsupported resource type: 

>>>> 有哪些deployment
debug:  OpenAI called us back wanting to invoke our function 'queryResource' with params '{"resource_type":"Deployment"}'

Found deployment: demo-1
Found deployment: nginx-deployment
Found deployment: test-deployment
Found deployment: coredns
Found deployment: local-path-provisioner

>>>> 有哪些pods
debug:  OpenAI called us back wanting to invoke our function 'queryResource' with params '{"resource_type":"Pod"}'

Found pod: demo-1-99979696-xlvtw
Found pod: nginx-deployment-7d9745ffbd-96pcm
Found pod: test-deployment-67b84cd4c6-7nvpp
Found pod: coredns-5d78c9869d-8hkxg
Found pod: coredns-5d78c9869d-qq8dz
Found pod: etcd-kind-control-plane
Found pod: kindnet-2bdlt
Found pod: kube-apiserver-kind-control-plane
Found pod: kube-controller-manager-kind-control-plane
Found pod: kube-proxy-25tfh
Found pod: kube-scheduler-kind-control-plane
Found pod: local-path-provisioner-6bc4bddd6b-5mp6h
 ```


总结：
  
    - 指令需要明确，否则大模型会理解错误，导致指令执行失败
    - 优化点为将用户输入的进行指令调优,确保指令的准确性
    - functionCalling需要注意返回，否则会导致指令执行失败
