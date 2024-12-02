package k8s

import (
	"context"
	"copilot/pkgs"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
)

type K8SHelper struct {
	c     *pkgs.OpenAI
	cGo   *pkgs.ClientGo
	tools []openai.Tool
}

func NewK8SHelper(kubeconfig string) (*K8SHelper, error) {
	client, err := pkgs.NewOpenAIClient()
	if err != nil {
		return nil, err
	}

	clientGo, err := pkgs.NewClientGo(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &K8SHelper{
		c:   client,
		cGo: clientGo,
	}, nil
}

func (k *K8SHelper) Hello() string {

	res, err := k.c.SendMessage("Hello", "你好，我是k8s助手")
	if err != nil {
		return err.Error()
	}

	return res
}

func (k *K8SHelper) SetTools() {

	// 用来生成 K8s YAML 并部署资源
	f1 := openai.FunctionDefinition{
		Name:        "generateAndDeployResource",
		Description: "生成 K8s YAML 并部署资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"user_input": {
					Type:        jsonschema.String,
					Description: "用户输出的文本内容，要求包含资源类型和镜像",
				},
			},
			Required: []string{"location"},
		},
	}
	t1 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f1,
	}

	// 用来查询 K8s 资源
	f2 := openai.FunctionDefinition{
		Name:        "queryResource",
		Description: "查询 K8s 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8s 资源标准类型，例如 Pod、Deployment、Service 等",
				},
			},
		},
	}

	t2 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f2,
	}

	// 用来删除 K8s 资源
	f3 := openai.FunctionDefinition{
		Name:        "deleteResource",
		Description: "删除 K8s 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8s 资源标准类型，例如 Pod、Deployment、Service 等",
				},
				"resource_name": {
					Type:        jsonschema.String,
					Description: "资源名称",
				},
			},
		},
	}

	t3 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f3,
	}

	k.tools = []openai.Tool{t1, t2, t3}
}

func (k *K8SHelper) FuncCalling(input string) string {

	dialogue := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: input},
	}

	resp, err := k.c.Client.CreateChatCompletion(context.TODO(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: dialogue,
			Tools:    k.tools,
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		return fmt.Sprintf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))

	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		return fmt.Sprintf("Completion error: len(toolcalls): %v\n", len(msg.ToolCalls))
	}

	// 查看 OpenAI 返回的函数调用信息
	// dialogue = append(dialogue, msg)
	m := fmt.Sprintf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
		msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

	fmt.Println("debug: ", m)

	// 4. 解析 OpenAI 返回的消息，手动调用对应的函数
	result, err := k.callFunction(msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
	if err != nil {
		return fmt.Sprintf("Error calling function: %v\n", err)
	}
	return result
}

func (k *K8SHelper) callFunction(name string, arguments string) (string, error) {
	if name == "generateAndDeployResource" {
		params := struct {
			UserInput string `json:"user_input"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return k.generateAndDeployResource(params.UserInput)
	}
	if name == "queryResource" {
		params := struct {
			Namespace    string `json:"namespace"`
			ResourceType string `json:"resource_type"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return k.queryResource(params.Namespace, params.ResourceType)
	}
	if name == "deleteResource" {
		params := struct {
			Namespace    string `json:"namespace"`
			ResourceType string `json:"resource_type"`
			ResourceName string `json:"resource_name"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		err := k.deleteResource(params.Namespace, params.ResourceType, params.ResourceName)
		if err != nil {
			return "", fmt.Errorf("failed to delete resource: %v", err)
		}
		return "Resource deleted successfully", nil
	}
	return "", fmt.Errorf("unknown function: %s", name)
}

// 4. 生成 K8s YAML 并部署资源
func (k *K8SHelper) generateAndDeployResource(userInput string) (string, error) {
	yamlContent, err := k.c.SendMessage("你现在是一个 K8s 资源生成器，根据用户输入生成 K8s YAML，注意除了 YAML 内容以外不要输出任何内容，此外不要把 YAML 放在 ``` 代码快里", userInput)
	if err != nil {
		return "", fmt.Errorf("ChatGPT error: %v", err)
	}
	// 调用 dynamic client 部署资源，封装到 pkgs/clien_go.go 中
	resources, err := restmapper.GetAPIGroupResources(k.cGo.DiscoveryClient)
	if err != nil {
		return "", err
	}

	// 将 YAML 转成 unstructured 对象
	unstructuredObj := &unstructured.Unstructured{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlContent), nil, unstructuredObj)
	if err != nil {
		return "", err
	}
	// 创建 mapper
	mapper := restmapper.NewDiscoveryRESTMapper(resources)
	// 从 unstructuredObj 中提取 GVK
	gvk := unstructuredObj.GroupVersionKind()
	// 用 GVK 转 GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", err
	}

	namespace := unstructuredObj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	_, err = k.cGo.DynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("YAML content:\n%s\n\nDeployment successful.", yamlContent), nil
}

// 5. 查询 K8s 资源
func (k *K8SHelper) queryResource(namespace, resourceType string) (string, error) {
	resourceType = strings.ToLower(resourceType)
	var gvr schema.GroupVersionResource
	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Query the resources using the dynamic client
	resourceList, err := k.cGo.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	// Iterate over the resources and print their names (or handle them as needed)
	result := ""
	for _, item := range resourceList.Items {
		result += fmt.Sprintf("Found %s: %s\n", resourceType, item.GetName())
	}

	return result, nil
}

func (k *K8SHelper) deleteResource(namespace, resourceType, resourceName string) error {
	resourceType = strings.ToLower(resourceType)
	var gvr schema.GroupVersionResource
	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	fmt.Printf("Deleting resource: %s  %s\n", resourceName, resourceType)
	err := k.cGo.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}
