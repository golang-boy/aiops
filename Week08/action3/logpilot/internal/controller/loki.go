/*
 * @Author: 刘慧东
 * @Date: 2024-12-02 16:32:33
 * @LastEditors: 刘慧东
 * @LastEditTime: 2024-12-02 16:33:39
 */
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	logv1 "logpilot/api/v1"
	"net/http"
	"net/url"
	"strconv"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *LogPilotReconciler) queryLog(ctx context.Context, logPilot *logv1.LogPilot) (ctrl.Result, int64, string, error) {

	logger := log.FromContext(ctx)
	currentTime := time.Now().Unix()
	preTimeStamp := logPilot.Status.PreTimeStamp
	// print preTimeStamp for debugging
	fmt.Printf("preTimeStamp: %s\n", preTimeStamp)
	var preTime int64
	if preTimeStamp == "" {
		preTime = currentTime - 5 // 这里减了5秒
	} else {
		preTime, _ = strconv.ParseInt(preTimeStamp, 10, 64)
	}
	// 纳秒级时间戳
	endTime := currentTime * 1000000000 // 当前时间的纳秒级时间戳

	// 这里减了5秒, preTime没有时，起始时间为当前时间-10s, preTime有值时，起始时间为preTime-5s
	startTime := (preTime - 5) * 1000000000 // 上次的时间戳

	fmt.Printf("startTime: %d, endTime: %d\n", startTime, endTime)

	if startTime >= endTime {
		logger.Info("startTime is greater than or equal to endTime, skipping log query")
		// print startTime and endTime for debugging
		return ctrl.Result{RequeueAfter: 10 * time.Second}, 0, "", ErrReQueue
	}

	// Loki 查询
	lokiQuery := logPilot.Spec.LokiPromQL

	lokiURL := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&end=%d",
		logPilot.Spec.LokiUrl, url.QueryEscape(lokiQuery), startTime, endTime)
	fmt.Println("lokiURL: ", lokiURL)
	lokiLogs, err := r.queryLoki(lokiURL)
	fmt.Println(lokiLogs)
	if err != nil {
		logger.Error(err, "unable to query Loki")
		return ctrl.Result{}, 0, "", err
	}

	return ctrl.Result{}, currentTime, lokiLogs, nil
}

// queryLoki 从 Loki 获取日志
func (r *LogPilotReconciler) queryLoki(lokiURL string) (string, error) {
	resp, err := http.Get(lokiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println(string(body))

	var lokiResponse map[string]interface{}
	err = json.Unmarshal(body, &lokiResponse)
	if err != nil {
		return "", err
	}

	data, ok := lokiResponse["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid Loki response data format")
	}

	// 检查 result 是否为空
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		return "", nil
	}

	return string(body), nil
}
