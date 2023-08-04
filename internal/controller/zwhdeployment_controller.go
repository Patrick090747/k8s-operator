/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myAppsv1 "zwh.com/pkg/zwh-deployment/api/v1"
)

// ZwhDeploymentReconciler reconciles a ZwhDeployment object
type ZwhDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.zwh.com,resources=zwhdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.zwh.com,resources=zwhdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.zwh.com,resources=zwhdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ZwhDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ZwhDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "ZwhDployment", req.NamespacedName)
	logger.Info("Reconcile is started")

	// TODO(user): your logic here
	//1.获取资源对象
	md := new(myAppsv1.ZwhDeployment)

	if err := r.Client.Get(ctx, req.NamespacedName, md); err != nil {

		return ctrl.Result{}, client.IgnoreNotFound(err)

	}
	//防止污染缓存
	mdCopy := md.DeepCopy()

	//========处理deployment===========
	deploy := new(appsv1.Deployment)
	if err := r.Client.Get(ctx, req.NamespacedName, deploy); err != nil {
		if errors.IsNotFound(err) {
			//2.1不存在对象
			//2.1.1创建deployment
			r.createDeployment(mdCopy)

		} else {

			return ctrl.Result{}, err
		}
	} else {
		//2.2存在对象
		//2.2.1更新deployment对象
		r.updateDeployment(mdCopy)
	}

	//========处理pod=========

	//========处理service=========
	//3.获取service资源对象
	//3.1不存在对象
	//3.1.1mode为ingres
	//3.1.1.1创建普通service
	//3.1.2mode为nodeport
	//3.1.2.1创建nodeport模式的service

	//3.2存在对象
	//3.2.1mode为ingress
	//3.2.1.1创建普通service
	//3.2.2mode为nodeport
	//3.2.2.1创建nodeport模式的service

	//========处理ingress=========
	//4.获取ingress资源对象
	//4.1不存在对象
	//4.1.1mode为ingress
	//4.1.1.1创建ingress
	//4.1.2mode为nodeport
	//4.1.2.1退出
	//4.2存在对象
	//4.2.1mode为ingress
	//4.2.1.1更新ingress
	//4.2.2mode为nodeport
	//4.2.2.1删除ingress
	logger.Info("Reconcile is ended")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZwhDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myAppsv1.ZwhDeployment{}).
		Complete(r)
}

func (r *ZwhDeploymentReconciler) createDeployment(_ *myAppsv1.ZwhDeployment) {

}

func (r *ZwhDeploymentReconciler) updateDeployment(_ *myAppsv1.ZwhDeployment) {

}
