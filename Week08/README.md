# 第八周

继续深入学习 operator


## 实践一

    创建一个operator,实现云上竞价实例数量的动态维护,由此构成一个竞价实例池,供给应用

   [代码在这里](./action2/)


流程：

1. 创建一个operator
    ```
    mkdir -p spotpool && cd spotpool
    go mod init spotpool
    kubebuilder init --domain aiops.org
    kubebuilder create api --group demo --version v1 --kind SpotPool
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

5. 安装crd, 运行operator
```
make install  && make run
```

6. 部署crd资源
```
kubectl apply -f config/crd/bases/demo.aiops.org_spotpools.yaml
```

执行结果

**要么余额不足，要么就是售罄，晚上回去再试。一小时多，网关和gpu服务器竞价实例大约花费2.5元**
```
INFO[0134] Instance ins-mqldieay State RUNNING          
INFO[0134] Found 1 running instances                    
INFO[0134] Starting 1 CVM instances                     
An API error has returned: [TencentCloudSDKError] Code=LimitExceeded.SpotQuota, Message=指定可用区竞价实例售罄。, RequestId=5319d197-3a51-4601-8d29-ae033be8da642024-12-02T15:26:19+08:00     DPANIC  odd number of arguments passed as key-value pairs for logging   {"controller": "spotpool", "controllerGroup": "demo.aiops.org", "controllerKind": "SpotPool", "SpotPool": {"name":"spotpool-sample","namespace":"default"}, "namespace": "default", "name": "spotpool-sample", "reconcileID": "4eac451f-12a0-490c-a583-cb985e3e7701", "ignored key": "failed to star
```

##　实践二

  [代码与实践一在一起](./action2/)

1. 在实践一的基础上(有了竞价实例池)，增加动态更新网关配置的能力，同时，部署kong网关, 这样通过网关，可以动态的将流量分发到不同的竞价实例上。

2. 竞价实例里预先部署好ollama，并且下载好大模型,将该实例制作为镜像，修改crd，增加网关配置，调整实例的镜像id, 启动后通过网关访问模型。

3. 当竞价实例被释放时，需要更新网关配置，将实例从网关中移除。


两份iac代码：[kong实例的创建](./action2/kong)，[llm模型预加载腾讯云镜像实例的创建](./action2/llm_image/)

operator的crd资源文件中配置申请好的信息

terraform llm_image/apply
```
Apply complete! Resources: 3 added, 0 changed, 0 destroyed.

Outputs:

availability_zone = "ap-singapore-2"
cvm_public_ip = "43.134.9.15"
image_id = "img-5gn3j31i"
region = "ap-singapore"
security_group_id = "sg-aglabemo"
ssh_password = "password123"
subnet_id = "subnet-gfrv79s2"
vpc_id = "vpc-86f66zht"
```
crd config
```
apiVersion: demo.aiops.org/v1
kind: SpotPool
metadata:
  labels:
    app.kubernetes.io/name: spotpool2
    app.kubernetes.io/managed-by: kustomize
  name: spotpool-sample
spec:
  secretId: xxx
  secretKey: xxx
  region: ap-singapore
  availabilityZone: ap-singapore-2
  instanceType: "GN7.2XLARGE32"
  minimum: 2
  maximum: 2
  subnetId: subnet-gfrv79s2
  vpcId: vpc-86f66zht
  securityGroupIds:
    - sg-aglabemo
  imageId: img-5gn3j31i
  # Ubuntu Server 24.04 LTS 公共镜像：img-mmytdhbn
  instanceChargeType: SPOTPAID
  kongGatewayIP: "119.28.76.99"
```

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


5. loki信息
```
Apply complete! Resources: 24 added, 0 changed, 0 destroyed.

Outputs:

grafana_url = "http://119.28.139.114:31001"
kube_config = "./config.yaml"
loki_password = "loki123"
loki_url = "http://119.28.139.114:31000"
public_ip = "119.28.139.114"
vm_password = "password123"
```

6. 配置crd资源



## 实践四

   从pod中获取日志，将日志送到ragflow, ragflow将日志送到大模型，获取建议，将建议发送到飞书

 

