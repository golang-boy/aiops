package k8s

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8SHelper) GetPodEventsAndLogs() (map[string][]string, error) {

	result := make(map[string][]string)
	// 获取 Warning 级别的事件
	events, err := k.cGo.Clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "type=Warning,reason=Failed",
	})
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}

	for _, event := range events.Items {
		podName := event.InvolvedObject.Name
		namespace := event.InvolvedObject.Namespace
		message := event.Message

		// fmt.Println("Event Message:", message)

		// 获取 Pod 的日志
		if event.InvolvedObject.Kind == "Pod" {
			logOptions := &corev1.PodLogOptions{}
			req := k.cGo.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
			podLogs, err := req.Stream(context.TODO())
			if err != nil {
				continue
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(podLogs)
			if err != nil {
				continue
			}
			// 只存有日志的 Pod，否则单纯靠 event 信息无法给出建议
			// 将事件信息存入 map
			result[podName] = append(result[podName], fmt.Sprintf("Event Message: %s", message))
			result[podName] = append(result[podName], fmt.Sprintf("Namespace: %s", namespace))
			// 将日志信息存入 map
			result[podName] = append(result[podName], fmt.Sprintf("Logs:\n%s", buf.String()))
		}
	}

	return result, nil
}

func (k *K8SHelper) AskGpt(infos map[string][]string) (string, error) {
	// 拼接所有 Pod 的事件和日志信息
	combinedInfo := "找到以下 Pod Waring 事件及其日志：\n\n"
	for podName, info := range infos {
		combinedInfo += fmt.Sprintf("Pod 名称: %s\n", podName)
		// 每个 Pod 的 event 和日志拼接成一个字符串
		for _, line := range info {
			combinedInfo += line + "\n"
		}
		combinedInfo += "\n"
	}

	fmt.Println(combinedInfo)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "您是一位k8s专家，你要帮助用户诊断多个Pod问题。",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("以下是多个 Pod Event 事件和对应的日志:\n%s\n请主要针对 Pod Log 给出实质性、可操作的建议, 优先给出纯命令操作方法", combinedInfo),
		},
	}

	// 请求 ChatGPT 获取建议
	resp, err := k.c.Client.CreateChatCompletion(
		context.TODO(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: messages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("error calling OpenAI API: %v", err)
	}

	responseText := resp.Choices[0].Message.Content
	return responseText, nil
}
