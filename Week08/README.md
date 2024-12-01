# 第八周

operator实践


## 实践一

    创建一个operator,实现云上竞价实例数量的动态维护,由此构成一个竞价实例池,供给应用


流程：

1. 创建一个operator
    ```
    mkdir -p spotpool && cd spotpool
    go mod init spotpool
    kuberbuilder init --domain aiops.org
    kuberbuilder create api --group demo --version v1 --kind SpotPool
    ```
   
2. 编辑api/v1/spotpool_types.go
    ```go
    type SpotPoolSpec struct {
        // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
        // Important: Run "make" to regenerate code after modifying this file

        // 访问云厂商的key
        SecretId string `json:"secretId,omitempty"`
        SecretKey string `json:"secretKey,omitempty"`

        Region string `json:"region"`
        AvaliableZone string `json:"availableZone"`

        InstanceType string `json:"instanceType"`
        SubnetId string `json:"subnetId"`
        SecurityGroupId string `json:"securityGroupId"`
        VpcId string `json:"vpcId"`
        ImageId string `json:"imageId"`
        InstanceChargeType string   `json:"instanceChargeType,omitempty"`

        // 动态实例数量范围
        MinSize int32 `json:"minSize"`
        MaxSize int32 `json:"maxSize"`
    }

    type SpotpoolStatus struct {
        // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
        // Important: Run "make" to regenerate code after modifying this file
        Size       int32              `json:"size,omitempty"`
        Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
        Instances  []Instances        `json:"instances,omitempty"`
    }

    type Instances struct {
        InstanceId string `json:"instanceId,omitempty"`
        PublicIp   string `json:"publicIp,omitempty"`
    }
    ```


3. 生成crd
   ```
   make manifests
   ```

4. 在controller中编写reconcile业务逻辑
 
    主体逻辑如下：
        1. 通过云厂商的api获取运行中的竞价实例数量(这里利用腾讯云开源的sdk进行操作https://cloud.tencent.com/document/sdk/Go)

           如果数量小于minSize,则创建竞价实例,如果大于maxSize,则销毁竞价实例

        2. 上述逻辑执行完毕后，重新入队，隔一段时间后，再次执行上述逻辑

5. 安装crd
```
make install
```

6. 运行operator
```make run```

7. 部署crd
```
kubectl apply -f config/crd/bases/demo.aiops.org_spotpools.yaml
```

##　实践二

1. 在实践一的基础上(有了竞价实例池)，增加动态更新网关配置的能力，同时，部署kong网关, 这样通过网关，可以动态的将流量分发到不同的竞价实例上。

2. 竞价实例里预先部署好ollama，并且下载好大模型,将该实例制作为镜像，修改crd，增加网关配置，调整实例的镜像id, 启动后通过网关访问模型。

3. 当竞价实例被释放时，需要更新网关配置，将实例从网关中移除。


## 实践三

  日志流检测的operator, 当有错误日志时，将日志发送到大模型获取建议，将严重的问题发送到飞书


流程：

1. 通过iac代码创建loki实例
``` shell
setup_cli() {
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
}

setup_helm_repo() {
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
}

setup_kube_prometheus_grafana_loki() {
    helm upgrade -i loki -n monitoring --create-namespace grafana/loki-stack -f /tmp/values.yaml
    helm upgrade -i kube-prometheus-stack -n monitoring --create-namespace prometheus-community/kube-prometheus-stack --version "54.0.1" --set grafana.adminPassword=loki123
}

set_loki_nodeport() {
    export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
    kubectl patch service loki -n monitoring -p '{"spec":{"type":"NodePort","ports":[{"port":3100,"nodePort":31000}]}}'
    kubectl patch service kube-prometheus-stack-grafana -n monitoring -p '{"spec":{"type":"NodePort","ports":[{"port":80,"nodePort":31001}]}}'
}
```
2. 创建operator

```
mkdir logpilot && cd logpilot
go mod init logpilot
kubebuilder init  --domain =aiops.org
kubebuilder create api --group log --version v1 --kind LogPilot
```

3. 修改crd

```go
    type LogPilotSpec struct {
        // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
        // Important: Run "make" to regenerate code after modifying this file
        // Loki URL
	    LokiURL string `json:"lokiURL"`
	    // Loki PromQL
	    LokiPromQL string `json:"lokiPromQL"`
	    // LLM Endpoint
	    LLMEndpoint string `json:"llmEndpoint"`
	    // LLM Token
	    LLMToken string `json:"llmToken"`
	    // LLM Model
	    LLMModel string `json:"llmModel"`
	    // Feishu Webhook
	    FeishuWebhook string `json:"feishuWebhook"`
    }
    type LogPilotStatus struct {
	    // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	    // Important: Run "make" to regenerate code after modifying this file
	    // PreTimeStamp
	    PreTimeStamp string `json:"preTimeStamp"`
    }
```

4. reconcile的逻辑

    1. 获取crd对象
    2. 每隔一段时间(10s)，执行一次reconcile
        1. 设置好查询的起止时间，通过loki的api获取日志
        2. 将日志发送到大模型，获取建议
        3. 如果包含严重问题，则发送到飞书


## 实践四

   从pod中获取日志，将日志送到ragflow, ragflow将日志送到大模型，获取建议，将建议发送到飞书

 

