package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	demov1 "spotpool/api/v1"
	"strings"
)

func (r *SpotPoolReconciler) checkKongAiProxy(ctx context.Context, spotpool *demov1.SpotPool) error {
	// 获取 KongGatewayIP
	kongGatewayIP := spotpool.Spec.KongGatewayIP
	kongURL := fmt.Sprintf("http://%s:8001", kongGatewayIP)

	// 使用 http.Client 发起 GET 请求
	client := &http.Client{}
	resp, err := client.Get(kongURL + "/services")
	if err != nil {
		return fmt.Errorf("failed to fetch services from Kong: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	// 解析响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	services, ok := result["data"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid response format, 'data' field not found")
	}

	// 检查是否存在 name=ai-proxy 的服务
	for _, service := range services {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			if name, exists := serviceMap["name"].(string); exists && name == "ai-proxy" {
				fmt.Println("ai-proxy service found.")
				return nil
			}
		}
	}

	// 创建新的 ai-proxy 服务
	fmt.Println("ai-proxy service not found, creating a new one.")
	if err := r.createKongServiceAndRouteAndUpstream(spotpool); err != nil {
		return err
	}

	fmt.Println("ai-proxy service and route created successfully.")
	return nil
}

func (r *SpotPoolReconciler) createKongServiceAndRouteAndUpstream(spotpool *demov1.SpotPool) error {
	kongGatewayIP := spotpool.Spec.KongGatewayIP
	kongURL := fmt.Sprintf("http://%s:8001", kongGatewayIP)

	client := &http.Client{}

	// 创建 Upstream
	upstreamURL := fmt.Sprintf("%s/upstreams", kongURL)
	upstreamData := map[string]string{
		"name": "llama2-upstream",
	}
	upstreamBody, _ := json.Marshal(upstreamData)
	// 请求
	upstreamResp, err := client.Post(upstreamURL, "application/json", bytes.NewReader(upstreamBody))
	if err != nil {
		return fmt.Errorf("failed to create upstream: %v", err)
	}
	defer upstreamResp.Body.Close()

	// 构造创建服务的 POST 请求数据
	serviceData := map[string]string{
		"name":     "ai-proxy",
		"host":     "llama2-upstream",
		"protocol": "http",
		"path":     "/api/chat",
	}

	serviceBody, _ := json.Marshal(serviceData)

	// 发起创建服务的请求
	serviceResp, err := client.Post(kongURL+"/services", "application/json", bytes.NewReader(serviceBody))
	if err != nil {
		return fmt.Errorf("failed to create ai-proxy service: %v", err)
	}
	defer serviceResp.Body.Close()

	if serviceResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create ai-proxy service, status: %d", serviceResp.StatusCode)
	}

	// 创建 route
	routeURL := fmt.Sprintf("%s/services/ai-proxy/routes", kongURL)
	routeData := map[string]interface{}{
		"name":  "ollama-chat",
		"paths": []string{"~/ollama-chat$"},
	}
	routeBody, _ := json.Marshal(routeData)

	// 发起创建 route 的请求
	routeResp, err := client.Post(routeURL, "application/json", bytes.NewReader(routeBody))
	if err != nil {
		return fmt.Errorf("failed to create ai-proxy route: %v", err)
	}
	defer routeResp.Body.Close()

	if routeResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create ai-proxy route, status: %d", routeResp.StatusCode)
	}

	return nil
}

func (r *SpotPoolReconciler) syncKongUpstream(ctx context.Context, spotpool *demov1.SpotPool) error {
	kongURL := fmt.Sprintf("http://%s:8001/upstreams/llama2-upstream/targets", spotpool.Spec.KongGatewayIP)

	// 获取当前 targets
	resp, err := http.Get(kongURL)
	if err != nil {
		return fmt.Errorf("failed to fetch targets from Kong: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status when fetching targets: %d", resp.StatusCode)
	}

	// 读取并解析响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// 使用 map 解析 JSON 数据
	var targetResponse map[string]interface{}
	err = json.Unmarshal(body, &targetResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal targets response: %v", err)
	}

	// 获取当前 targets 数据
	currentTargets, ok := targetResponse["data"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid response format, 'data' field not found")
	}

	// 从 spotpool 中获取实例列表
	instances := spotpool.Status.Instances

	// 创建一个目标 IP 的集合，以便于后续比较
	instanceIPs := make(map[string]bool)
	for _, instance := range instances {
		instanceIPs[instance.PublicIp] = true
	}

	// 如果当前 targets 为空，直接添加所有实例为新的 targets
	if len(currentTargets) == 0 {
		for _, instance := range instances {
			instanceIP := fmt.Sprintf("%s:8080", instance.PublicIp)
			postData := map[string]string{
				"target": instanceIP,
			}
			postBody, err := json.Marshal(postData)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON for new target: %v", err)
			}

			postResp, err := http.Post(
				kongURL,
				"application/json",
				strings.NewReader(string(postBody)),
			)
			if err != nil {
				return fmt.Errorf("failed to create new target for %s: %v", instanceIP, err)
			}
			defer postResp.Body.Close()

			if postResp.StatusCode != http.StatusCreated {
				return fmt.Errorf("failed to create target for %s, response status: %d", instanceIP, postResp.StatusCode)
			}

			fmt.Printf("Created target for %s\n", instanceIP)
		}
		return nil // 处理完所有实例后返回
	}

	// 遍历当前 targets，检查与实例的匹配
	for _, target := range currentTargets {
		targetMap, ok := target.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取目标 IP
		targetIP := targetMap["target"].(string)      // 格式: "43.156.25.202:8080"
		targetHost := strings.Split(targetIP, ":")[0] // 获取 IP 部分

		// 如果在实例列表中找到匹配的 IP
		if _, exists := instanceIPs[targetHost]; !exists {
			// 如果目标 IP 不在实例列表中，则删除该 target
			targetID := targetMap["target"].(string)
			deleteURL := fmt.Sprintf("http://%s:8001/upstreams/llama2-upstream/targets/%s", spotpool.Spec.KongGatewayIP, targetID)
			req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create delete request: %v", err)
			}

			delResp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to delete target %s: %v", targetID, err)
			}
			delResp.Body.Close()

			if delResp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("failed to delete target %s, response status: %d", targetID, delResp.StatusCode)
			}

			fmt.Printf("Deleted target %s\n", targetID)
		}
	}

	// 检查实例列表中是否有目标不存在的情况
	for _, instance := range instances {
		instanceIP := fmt.Sprintf("%s:8080", instance.PublicIp) // 使用格式化的 target 字符串
		found := false

		// 检查当前目标是否存在
		for _, target := range currentTargets {
			targetMap, ok := target.(map[string]interface{})
			if !ok {
				continue
			}
			if targetMap["target"] == instanceIP {
				found = true
				break
			}
		}

		// 如果实例在目标列表中不存在，则添加新的 target
		if !found {
			postData := map[string]string{
				"target": instanceIP,
			}
			postBody, err := json.Marshal(postData)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON for new target: %v", err)
			}

			postResp, err := http.Post(
				kongURL,
				"application/json",
				strings.NewReader(string(postBody)),
			)
			if err != nil {
				return fmt.Errorf("failed to create new target for %s: %v", instanceIP, err)
			}
			defer postResp.Body.Close()

			if postResp.StatusCode != http.StatusCreated {
				return fmt.Errorf("failed to create target for %s, response status: %d", instanceIP, postResp.StatusCode)
			}

			fmt.Printf("Created target for %s\n", instanceIP)
		}
	}

	return nil
}
