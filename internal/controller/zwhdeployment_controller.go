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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"

	myAppsv1 "zwh.com/pkg/zwh-deployment/api/v1"
)

var WaitRequeue = 10 * time.Second

// ZwhDeploymentReconciler reconciles a ZwhDeployment object
type ZwhDeploymentReconciler struct {
	client.Client
	DynamicClient dynamic.Interface // 用来访问 issuer和certificate资源
	Scheme        *runtime.Scheme
}

// 创建GVR, 共动态客户端使用
var (
	// issuer
	issuerGVR = schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "issuers",
	}
	// certificate
	certGVR = schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}
)

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

	// 状态更新策略
	// 创建的时候
	//    更新为创建
	// 更新的时候
	//    根据获取的状态来判断时候更新status
	// 删除的时候
	//    只有在操作 ingress 的时候，并且 mode 为 nodeport 的时候

	logger := log.FromContext(ctx, "ZwhDployment", req.NamespacedName)

	logger.Info("Reconcile is started.")

	// 1. 获取资源对象
	md := new(myAppsv1.ZwhDeployment)
	if err := r.Client.Get(ctx, req.NamespacedName, md); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 防止污染缓存
	mdCopy := md.DeepCopy()

	// 处理最终的返回
	defer func() {
		if r.Ready(mdCopy) {
			_ = r.Client.Status().Update(ctx, mdCopy)
			return
		}
		if mdCopy.Status.ObservedGeneration != md.Status.ObservedGeneration {
			_ = r.Client.Status().Update(ctx, mdCopy)
		}
	}()

	// ======= 处理 deployment ======
	// 2. 获取deployment资源对象
	deploy := new(appsv1.Deployment)
	if err := r.Client.Get(ctx, req.NamespacedName, deploy); err != nil {
		if errors.IsNotFound(err) {
			// 2.1 不存在对象
			// 2.1.1 创建 deployment
			if errCreate := r.createDeployment(ctx, mdCopy); err != nil {
				return ctrl.Result{}, errCreate
			}
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeDeployment,
				fmt.Sprintf("Deployment %s,err:%s", req.Name, err.Error()),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonDeploymentNotReady,
			); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else {
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeDeployment,
				fmt.Sprintf("Deployment %s,err:%s", req.Name, err.Error()),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonDeploymentNotReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
	} else {
		//2.2存在对象
		//2.2.1更新deployment
		if err := r.updateDeployment(ctx, mdCopy, deploy); err != nil {
			return ctrl.Result{}, err
		}
		if deploy.Status.AvailableReplicas == mdCopy.Spec.Replicas {
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeDeployment,
				fmt.Sprintf(myAppsv1.ConditionMessageDeploymentOKFmt, req.Name),
				myAppsv1.ConditionStatusTrue,
				myAppsv1.ConditionReasonDeploymentReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else {
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeDeployment,
				fmt.Sprintf(myAppsv1.ConditionMessageDeploymentNotFmt, req.Name),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonDeploymentNotReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		}
	}

	// ======= 处理 service =========
	// 3. 获取 service 资源对象
	svc := new(corev1.Service)
	if err := r.Client.Get(ctx, req.NamespacedName, svc); err != nil {
		if errors.IsNotFound(err) {
			// 3.1 不存在 创建 service
			//3.1.1mode为ingress
			if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeIngress {
				//3.1.1.1创建普通service
				if err := r.createService(ctx, mdCopy); err != nil {
					return ctrl.Result{}, err
				}
			} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeNodePort {
				//mode为nodeport
				//3.1.2.1创建 nodeport模式的 service
				if err := r.createNPService(ctx, mdCopy); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, myAppsv1.ErrorNotSupportMode
			}
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeService,
				fmt.Sprintf(myAppsv1.ConditionMessageServiceNotFmt, req.Name),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonServiceNotReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else {
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeService,
				fmt.Sprintf("Service %s,err:%s", req.Name, err.Error()),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonServiceNotReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
	} else {
		//3.2存在
		if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeIngress {
			//3.2.1 mode为ingress
			//3.2.1.1更新普通的service
			if err := r.updateService(ctx, mdCopy, svc); err != nil {
				return ctrl.Result{}, err
			}
		} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeNodePort {
			//3.2.2 mode为nodeport
			//3.2.2.1更新nodeport模式的service
			if err := r.updateNPSerive(ctx, mdCopy, svc); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, myAppsv1.ErrorNotSupportMode
		}
		if _, errStatus := r.updateStatus(ctx,
			mdCopy,
			myAppsv1.ConditionTypeService,
			fmt.Sprintf(myAppsv1.ConditionMessageServiceOKFmt, req.Name),
			myAppsv1.ConditionStatusTrue,
			myAppsv1.ConditionReasonServiceReady); errStatus != nil {
			return ctrl.Result{}, errStatus
		}
	}
	//if err := r.createService(ctx, mdCopy); err != nil {
	//			return ctrl.Result{}, err
	//		}
	//	}

	//======处理ingress ==========
	//4获取ingress资源
	ig := new(networkv1.Ingress)
	if err := r.Client.Get(ctx, req.NamespacedName, ig); err != nil {
		if errors.IsNotFound(err) {
			// 4.1 不存在
			if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeIngress {
				// 4.1.1 mode 为 ingress
				// 4.1.1.1 创建 ingress
				if err := r.createIngress(ctx, mdCopy); err != nil {
					return ctrl.Result{}, err
				}
				if _, errStatus := r.updateStatus(ctx,
					mdCopy,
					myAppsv1.ConditionTypeIngress,
					fmt.Sprintf(myAppsv1.ConditionMessageIngressNotFmt, req.Name),
					myAppsv1.ConditionStatusFalse,
					myAppsv1.ConditionReasonIngressNotReady); errStatus != nil {
					return ctrl.Result{}, errStatus
				}
			} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeNodePort {
				//4.1.2mode为nodeport
				//4.1.2.1退出
				return ctrl.Result{}, nil
			}
		} else {
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeIngress,
				fmt.Sprintf("Ingress %s,err: %s", req.Name, err.Error()),
				myAppsv1.ConditionStatusFalse,
				myAppsv1.ConditionReasonIngressNotReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}

	} else {
		//4,2存在
		if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeIngress {
			//4.2.1 mode为ingress
			//4,2,1,1 更新ingress
			if err := r.updateIngress(ctx, mdCopy, ig); err != nil {
				return ctrl.Result{}, err
			}
			if _, errStatus := r.updateStatus(ctx,
				mdCopy,
				myAppsv1.ConditionTypeIngress,
				fmt.Sprintf(myAppsv1.ConditionMessageIngressOKFmt, req.Name),
				myAppsv1.ConditionStatusTrue,
				myAppsv1.ConditionReasonIngressReady); errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == myAppsv1.ModeNodePort {
			//4,2,2 mode 为nodeport
			// 4.2.2.1删除ingress
			if err := r.deleteIngress(ctx, mdCopy); err != nil {
				return ctrl.Result{}, err
			}
			r.deleteStatus(mdCopy, myAppsv1.ConditionTypeIngress)
		}
	}
	//最后检查状态时候最终完成
	if sus, errStatus := r.updateStatus(ctx,
		mdCopy,
		"",
		"",
		"",
		""); errStatus != nil {
		return ctrl.Result{}, errStatus
	} else if !sus {
		logger.Info("reconcile is ended")
		return ctrl.Result{RequeueAfter: WaitRequeue}, nil
	}
	logger.Info("reconcile is ended")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZwhDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myAppsv1.ZwhDeployment{}).
		Owns(&appsv1.Deployment{}). //监控deployment类型，变更就触发reconciler
		Owns(&corev1.Service{}).    //监控service类型，变更就触发reconciler
		Owns(&networkv1.Ingress{}). //监控ingress类型，变更就触发reconciler
		Complete(r)
}

func (r *ZwhDeploymentReconciler) createDeployment(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	deploy, err := NewDeployment(md)
	if err != nil {
		return err
	}

	//设置deployment所属于md
	if err := controllerutil.SetControllerReference(md, deploy, r.Scheme); err != nil {
		return err
	}
	return r.Client.Create(ctx, deploy, client.DryRunAll)

}

func (r *ZwhDeploymentReconciler) updateDeployment(ctx context.Context, md *myAppsv1.ZwhDeployment, dp *appsv1.Deployment) error {
	deploy, err := NewDeployment(md)

	if err != nil {
		return err
	}
	// 设置 deployment 所属于 md
	if err := controllerutil.SetControllerReference(md, deploy, r.Scheme); err != nil {
		return err
	}

	//预更新deployment。得到更新后的数据
	if err := r.Update(ctx, deploy, client.DryRunAll); err != nil {
		return err
	}
	if reflect.DeepEqual(dp.Spec, deploy.Spec) {
		return nil
	}
	return r.Client.Update(ctx, deploy)

}

func (r *ZwhDeploymentReconciler) createService(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	svc, err := NewService(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, svc, r.Scheme); err != nil {
		return err
	}
	return r.Client.Create(ctx, svc)
}

func (r *ZwhDeploymentReconciler) createNPService(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	svc, err := NewServiceNP(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, svc, r.Scheme); err != nil {
		return err
	}
	return r.Client.Create(ctx, svc)
}

func (r *ZwhDeploymentReconciler) updateService(ctx context.Context, md *myAppsv1.ZwhDeployment, service *corev1.Service) error {
	svc, err := NewService(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, svc, r.Scheme); err != nil {
		return err
	}
	//预更新service。得到更新后的数据
	if err := r.Update(ctx, svc, client.DryRunAll); err != nil {
		return err
	}
	if reflect.DeepEqual(service.Spec, svc.Spec) {
		return nil
	}
	return r.Client.Update(ctx, svc)

}

func (r *ZwhDeploymentReconciler) updateNPSerive(ctx context.Context, md *myAppsv1.ZwhDeployment, service *corev1.Service) error {
	svc, err := NewServiceNP(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, svc, r.Scheme); err != nil {
		return err
	}
	//预更新service。得到更新后的数据
	if err := r.Update(ctx, svc, client.DryRunAll); err != nil {
		return err
	}
	if reflect.DeepEqual(service.Spec, svc.Spec) {
		return nil
	}
	return r.Client.Update(ctx, svc)
}

func (r *ZwhDeploymentReconciler) createIngress(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	ig, err := NewIngress(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, ig, r.Scheme); err != nil {
		return err
	}
	return r.Client.Create(ctx, ig)
}

func (r *ZwhDeploymentReconciler) updateIngress(ctx context.Context, md *myAppsv1.ZwhDeployment, ingress *networkv1.Ingress) error {
	ig, err := NewIngress(md)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(md, ig, r.Scheme); err != nil {
		return err
	}
	//预更新service。得到更新后的数据
	if err := r.Update(ctx, ingress, client.DryRunAll); err != nil {
		return err
	}
	if reflect.DeepEqual(ingress.Spec, ig.Spec) {
		return nil
	}
	return r.Client.Update(ctx, ig)
}

func (r *ZwhDeploymentReconciler) deleteIngress(ctx context.Context, md *myAppsv1.ZwhDeployment) error {

	ig, err := NewIngress(md)
	if err != nil {
		return err
	}
	return r.Client.Delete(ctx, ig)
}

// 更新Condition，并变更版本
func (r *ZwhDeploymentReconciler) updateConditions(md *myAppsv1.ZwhDeployment, conditionType, message, status, reason string) {
	// 1. 获取 status
	//status := md.Status
	// 2. 获取 conditions 字段
	//conditions := status.Conditions
	// 3. 根据当前的需求，获取指定的 condition
	var condition *myAppsv1.Condition
	for i := range md.Status.Conditions {
		// 4. 是否获取到
		if md.Status.Conditions[i].Type == conditionType {
			// 4.1 获取到了
			condition = &md.Status.Conditions[i]
		}
	}
	if condition != nil {
		// 4.1.1 获取当前线上的 conditon 状态，与存储的condition进行比较，如果相同，跳过。不同，替换
		if condition.Status != status ||
			condition.Message != message ||
			condition.Reason != reason {
			condition.Status = status
			condition.Message = message
			condition.Reason = reason
			md.Status.ObservedGeneration += 1
		}
	} else {
		// 4.2 没获取到，创建这个conditon，更新到conditons中
		md.Status.Conditions = append(md.Status.Conditions,
			createCondition(conditionType, message, status, reason))
		md.Status.ObservedGeneration += 1
	}
}

// Ready 判断本次 reconcile 是否达到预期
func (r *ZwhDeploymentReconciler) Ready(md *myAppsv1.ZwhDeployment) bool {
	// 5. 继续处理其他的conditions
	m, re, p, sus := isSuccess(md.Status.Conditions)
	if sus {
		// 6.1 如果所有conditions的状态都为成功，则更新总的 status 为成功。
		md.Status.Message = myAppsv1.StatusMessageSuccess
		md.Status.Reason = myAppsv1.StatusReasonSuccess
		md.Status.Phase = myAppsv1.StatusPhaseComplete
	} else {
		// 6.2 遍历所有的conditons 状态，如果有任意一个condition不是完成的状态，则将这个状态更新到总的 status 中。更待一定时间再次入队。
		if md.Status.Message != m ||
			md.Status.Reason != re ||
			md.Status.Phase != p {
			md.Status.Message = m
			md.Status.Reason = re
			md.Status.Phase = p
			md.Status.ObservedGeneration += 1
		}
	}
	return sus
}

func (r *ZwhDeploymentReconciler) updateStatus(ctx context.Context, md *myAppsv1.ZwhDeployment, conditionType, message, status, reason string) (bool, error) {
	if conditionType != "" {
		//1.获取status
		//status := md.Status
		//2.获取conditions字段
		//conditions := status.Conditions

		//3.根据当前的需求获取指定的condition
		var condition *myAppsv1.Condition
		for i := range md.Status.Conditions {
			//4.判断是否获取到
			if md.Status.Conditions[i].Type == conditionType {
				//4.1获取到了
				condition = &md.Status.Conditions[i]
			}
		}

		if condition != nil {
			//4.1.1获取当前线上condition状态，与存储的condition进行比较，如果相同那么就跳过。不同，替换
			if condition.Status != status ||
				condition.Message != message ||
				condition.Reason != reason {
				condition.Status = status
				condition.Message = message
				condition.Reason = reason
			}
		} else {
			//4.2没获取到，创建这个condition，更新到conditions中
			md.Status.Conditions = append(md.Status.Conditions,
				createCondition(conditionType, message, status, reason))
		}
	}

	//5.继续处理其他的condition

	m, re, p, sus := isSuccess(md.Status.Conditions)
	if sus {
		//6.1如果所有conditions的状态都为成功，则更新总的status为成功
		md.Status.Message = myAppsv1.StatusMessageSuccess
		md.Status.Reason = myAppsv1.StatusReasonSuccess
		md.Status.Phase = myAppsv1.StatusPhaseComplete
	} else {
		//6.2遍历所有的conditions状态，如果有任意一个condition不是完成的状态，则将这个状态更新到总的status中。等待一段时间后入队
		md.Status.Message = m
		md.Status.Reason = re
		md.Status.Phase = p
	}
	//执行更新
	return sus, r.Client.Status().Update(ctx, md)
}

// 需要是幂等的，可以多次执行，不管是否存在。如果存在就删除，不存在就什么也不做
// 只是删除对应的Condition不做更多的操作
func (r *ZwhDeploymentReconciler) deleteStatus(md *myAppsv1.ZwhDeployment, conditionType string) {
	// 1. 遍历conditions
	var tmp []myAppsv1.Condition
	copy(tmp, md.Status.Conditions)
	for i := range tmp {
		// 2. 找到要删除的对象
		if tmp[i].Type == conditionType {
			// 3. 执行删除
			md.Status.Conditions = deleteCondition(tmp, i)
		}
	}
}

func (r *ZwhDeploymentReconciler) createIssuer(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	// 1. 创建 issuer 资源
	i, err := NewIssuer(md)
	if err != nil {
		return err
	}

	// 设置 issuer 所属于 md
	if err := controllerutil.SetControllerReference(md, i, r.Scheme); err != nil {
		return err
	}

	// 在k8s中创建issuer资源
	if _, err := r.DynamicClient.Resource(issuerGVR).
		Namespace(md.Namespace).
		Create(ctx, i, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			// 这是一个折中的考虑，在没有比较完整的处理证书更新的方案前，
			// 这是一个简单并且不会出现意外错误的处理方式
			return nil
		}
		return err
	}
	return nil
}

func (r *ZwhDeploymentReconciler) createCert(ctx context.Context, md *myAppsv1.ZwhDeployment) error {
	// 1. 创建 issuer 资源
	c, err := NewCert(md)
	if err != nil {
		return err
	}

	// 设置 issuer 所属于 md
	if err := controllerutil.SetControllerReference(md, c, r.Scheme); err != nil {
		return err
	}

	// 在k8s中创建certificate资源
	if _, err := r.DynamicClient.Resource(certGVR).
		Namespace(md.Namespace).
		Create(ctx, c, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			// 这是一个折中的考虑，在没有比较完整的处理证书更新的方案前，
			// 这是一个简单并且不会出现意外错误的处理方式
			return nil
		}
		return err
	}
	return nil
}

func isSuccess(conditions []myAppsv1.Condition) (message, reason, phase string, sus bool) {
	if len(conditions) == 0 {
		return "", "", "", false
	}
	for i := range conditions {
		if conditions[i].Status == myAppsv1.ConditionStatusFalse {
			return conditions[i].Message, conditions[i].Reason, conditions[i].Type, false
		}
	}
	return "", "", "", true
}

func createCondition(conditionType, message, status, reason string) myAppsv1.Condition {
	return myAppsv1.Condition{
		Type:               conditionType,
		Message:            message,
		Status:             status,
		Reason:             reason,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}

func deleteCondition(conditions []myAppsv1.Condition, i int) []myAppsv1.Condition {
	// 前提：切片中的元素顺序不敏感
	// 1. 要删除的元素的索引值不能大于切片长度
	if i >= len(conditions) {
		return []myAppsv1.Condition{}
	}

	// 2. 如果切片长度为1，且索引值为0，直接清空
	if len(conditions) == 1 && i == 0 {
		return conditions[:0]
	}

	// 3. 如果长度-1等于索引值，删除最后一个元素
	if len(conditions)-1 == i {
		return conditions[:len(conditions)-1]
	}

	// 4. 交换索引位置的元素和最后一个元素，删除最后一个元素
	conditions[i], conditions[len(conditions)-1] = conditions[len(conditions)-1], conditions[i]
	return conditions[:len(conditions)-1]
}
