/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	logv1 "logpilot/api/v1"
)

// LogPilotReconciler reconciles a LogPilot object
type LogPilotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	ErrReQueue = fmt.Errorf("requeue")
)

// +kubebuilder:rbac:groups=log.aiops.org,resources=logpilots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=log.aiops.org,resources=logpilots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=log.aiops.org,resources=logpilots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LogPilot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *LogPilotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var logPilot logv1.LogPilot

	if err := r.Get(ctx, req.NamespacedName, &logPilot); err != nil {
		logger.Error(err, "unable to fetch LogPilot")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	res, nextStartime, logs, err := r.queryLog(ctx, &logPilot)
	if err != nil {
		if errors.Is(err, ErrReQueue) {
			return res, nil
		}
		logger.Error(err, "unable to query log")
		return ctrl.Result{}, err
	}

	// 如果有日志结果，调用 LLM 进行分析
	if logs != "" {
		fmt.Println("send logs to LLM")
		analysisResult, err := r.analyzeLogsWithLLM(logPilot.Spec.LLMEndpoint, logPilot.Spec.LLMToken, logPilot.Spec.LLMModel, logs)
		if err != nil {
			logger.Error(err, "unable to analyze logs with LLM")
			return ctrl.Result{}, err
		}

		// 如果 LLM 返回的结果表明日志有问题，发送飞书告警
		if analysisResult.HasErrors {
			err := r.sendFeishuAlert(logPilot.Spec.FeishuWebhook, analysisResult.Analysis)
			if err != nil {
				logger.Error(err, "unable to send alert to Feishu")
				return ctrl.Result{}, err
			}
		}
	}

	// 更新状态中的 PreTimeStamp
	logPilot.Status.PreTimeStamp = fmt.Sprintf("%d", nextStartime)
	if err := r.Status().Update(ctx, &logPilot); err != nil {
		logger.Error(err, "unable to update LogPilot status")
		return ctrl.Result{}, err
	}

	// 10 秒后再次 Reconcile
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPilotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logv1.LogPilot{}).
		Named("logpilot").
		Complete(r)
}
