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

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1 "k8s.io/api/networking/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "app/api/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.aiops.org,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.aiops.org,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.aiops.org,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var app appv1.Application
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		logger.Error(err, "unable to fetch Application")
		return ctrl.Result{}, nil
	}

	logger.Info("Reconciling Application", "Name", app.Name)

	// 创建deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}
	labels := map[string]string{
		"app": app.Name,
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		replicas := int32(1)
		if app.Spec.Deployment.Replicas != 0 {
			replicas = app.Spec.Deployment.Replicas
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Name,
							Image: app.Spec.Deployment.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: app.Spec.Deployment.Port, // 配置中端口，作为容器端口
								},
							},
						},
					},
				},
			},
		}
		// 标记该资源谁管理
		if err := controllerutil.SetControllerReference(&app, deployment, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		app.Status.AvailableReplicas = 0
		return ctrl.Result{}, err
	}

	logger.Info("Deployment created or updated", "Name", deployment.Name)

	// 创建服务
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports:    app.Spec.Service.Ports,
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(&app, service, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "unable to create or update Service")
		return ctrl.Result{}, err
	}

	logger.Info("Service created or updated", "Name", service.Name)

	// 创建ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		ingress.Spec = networkingv1.IngressSpec{
			IngressClassName: app.Spec.Ingress.IngressClassName,
			Rules:            app.Spec.Ingress.Rules,
		}
		// Set owner reference
		if err := controllerutil.SetControllerReference(&app, ingress, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "unable to create or update Ingress")
		return ctrl.Result{}, err
	}

	logger.Info("Ingress created or updated", "Name", ingress.Name)

	// 都成功后更新状态
	app.Status.AvailableReplicas = *deployment.Spec.Replicas
	if err := r.Status().Update(ctx, &app); err != nil {
		logger.Error(err, "unable to update Application status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.Application{}).
		Named("application").
		Complete(r)
}
