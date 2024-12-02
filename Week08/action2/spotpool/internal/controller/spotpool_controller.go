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
	"time"

	logger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "spotpool/api/v1"
)

// SpotPoolReconciler reconciles a SpotPool object
type SpotPoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=demo.aiops.org,resources=spotpools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=demo.aiops.org,resources=spotpools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=demo.aiops.org,resources=spotpools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SpotPool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SpotPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	spotpool := &demov1.SpotPool{}
	err := r.Get(ctx, req.NamespacedName, spotpool)
	if err != nil {
		log.Error(err, "unable to fetch SpotPool")
		return ctrl.Result{}, err
	}

	err = r.checkKongAiProxy(ctx, spotpool)
	if err != nil {
		log.Error(err, "unable to check KongAiProxy")
		return ctrl.Result{}, err
	}

	// res, err := r.checkRunningInstances(ctx, spotpool)
	// if err != nil {
	// 	if errors.Is(err, errInstanceRetry) {
	// 		return res, nil
	// 	}
	// 	log.Error(err, "unable to check RunningInstances")
	// 	return ctrl.Result{}, err
	// }

	// 获取 Running 实例数量，这里需要一直检查，因为 IP 地址是异步分配的
	_, err = r.getRunningInstanceIDs(spotpool)
	if err != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Get the current number of running CVM instances.
	runningCount, err := r.countRunningInstances(spotpool)
	if err != nil {
		log.Error(err, "failed to get running CVM instance count")
		return ctrl.Result{}, err
	}

	// Take actions based on the desired state (Minimum and Maximum) and current state (runningCount).
	switch {
	case runningCount < spotpool.Spec.Minimum:
		// Need to start more CVMs.
		delta := spotpool.Spec.Minimum - runningCount
		logger.Info("Starting ", delta, " CVM instances")
		err = r.runInstances(spotpool, delta)
		if err != nil {
			log.Info(err.Error(), "failed to start CVM instances")
			// 40s 后重试
			return ctrl.Result{RequeueAfter: 40 * time.Second}, nil
		}
	case runningCount > spotpool.Spec.Maximum:
		// Need to terminate CVMs.
		delta := runningCount - spotpool.Spec.Maximum
		logger.Info("Terminating ", delta, " CVM instances")
		err = r.terminateInstances(spotpool, delta)
		if err != nil {
			log.Info(err.Error(), "failed to terminate CVM instances")
			return ctrl.Result{RequeueAfter: 40 * time.Second}, nil
		}
	}

	// Update the Spotpool status with the current number of running instances.
	spotpool.Status.Size = runningCount

	log.Info("SpotPool %s has %d running instances", spotpool.Name, runningCount)

	err = r.Status().Update(ctx, spotpool)
	if err != nil {
		log.Error(err, "failed to update Spotpool status")
		return ctrl.Result{}, err
	}

	err = r.syncKongUpstream(ctx, spotpool)
	if err != nil {
		log.Error(err, "unable to sync KongUpstream")
		return ctrl.Result{RequeueAfter: 40 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: 40 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpotPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.SpotPool{}).
		Named("spotpool").
		Complete(r)
}
