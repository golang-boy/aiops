package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// LLMAnalysisResult 用于存储 LLM 分析的结果
type LLMAnalysisResult struct {
	HasErrors bool   // 是否有错误日志
	Analysis  string // LLM 返回的分析结果
}

// analyzeLogsWithLLM 调用 LLM 接口分析日志
func (r *LogPilotReconciler) analyzeLogsWithLLM(endpoint, token, model, logs string) (*LLMAnalysisResult, error) {
	config := openai.DefaultConfig(token)
	config.BaseURL = endpoint
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("你现在是一名运维专家，以下日志是从日志系统里获取的日志，请分析日志的错误等级，如果遇到严重的问题，例如请求外部系统失败、外部系统故障、致命故障、数据库连接错误等严重问题时，请给出简短的建议，对于你认为严重需要通知运营人员的，请在返回内容里增加[feishu]标识:\n%s", logs),
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	return parseLLMResponse(&resp), nil
}

// parseLLMResponse 解析 LLM API 的响应
func parseLLMResponse(resp *openai.ChatCompletionResponse) *LLMAnalysisResult {
	result := &LLMAnalysisResult{
		Analysis: resp.Choices[0].Message.Content, // 从 LLM 返回的文本中获取分析结果
	}

	// 简单判断分析结果是否包含错误的标识符
	if strings.Contains(strings.ToLower(result.Analysis), "feishu") {
		result.HasErrors = true
	} else {
		result.HasErrors = false
	}

	return result
}

// sendFeishuAlert 发送飞书告警
func (r *LogPilotReconciler) sendFeishuAlert(webhook, analysis string) error {
	// 飞书消息内容
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": analysis,
		},
	}

	// 将消息内容序列化为 JSON
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(messageBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发出请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send Feishu alert, status code: %d", resp.StatusCode)
	}

	return nil
}
