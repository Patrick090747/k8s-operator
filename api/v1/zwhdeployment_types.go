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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ZwhDeploymentSpec defines the desired state of ZwhDeployment
type ZwhDeploymentSpec struct {
	//Image 存储镜像地址
	Image string `json:"image"`
	//Port 存储服务提供的端口
	Port int32 `json:"port"`
	//Replicas 存储要部署多少个副本
	//+optional
	Replicas int32 `json:"replicas,omitempty"`
	//StartCmd 存储启动命令
	//+optional
	StartCmd string `json:"startCmd,omitempty"`
	//Args 存储启动命令参数
	//+optional
	Args []string `json:"args,omitempty"`
	//Volumes	存储存储卷，直接使用pod中的定义方式
	//Volumes []corev1.Volume
	////VolumeMounts 存储存储卷挂载，直接使用pod中的定义方式
	//VolumeMounts []corev1.VolumeMount
	//Environments 存储环境变量，直接使用pod中的定义方式
	Environments []corev1.EnvVar `json:"environments,omitempty"`
	//Expose service要暴露的端口
	Expose *Expose `json:"expose"`
}

// Expose 存储服务暴露的端口
type Expose struct {
	//Mode 模式 nodeport or ingress
	Mode string `json:"mode"`
	//NodePort 节点端口	,在mode 为nodeport时，需要填写
	NodePort int32 `json:"nodePort,omitempty"`
	//IngressDomain 域名.在mode 为ingress时，需要填写
	//+optional
	IngressDomain string `json:"ingressDomain,omitempty"`
	//ServicePort service 端口,一般是随机生成,为了防止冲突，使用同上面ZwhDeploymentSpec的port值
	//+optional
	ServicePort int32 `json:"servicePort,omitempty"`
}

// ZwhDeploymentStatus defines the observed state of ZwhDeployment
type ZwhDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// 处于什么阶段
	Phase string `json:"phase,omitempty"`
	// 这个阶段的信息
	Message string `json:"message,omitempty"`
	// 处于这个阶段的原因
	Reason string `json:"reason,omitempty"`
	// 这个阶段的子资源的状态
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition 子资源的状态
type Condition struct {
	//子资源类型
	Type string `json:"type,omitempty"`
	//这个子资源状态的信息
	Status string `json:"status,omitempty"`
	//处于这个状态的原因
	Reason string `json:"reason,omitempty"`
	//这个子资源状态对的信息
	Message string `json:"message,omitempty"`
	//最后创建、更新时间
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"` // 最后创建/更新的时间
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ZwhDeployment is the Schema for the zwhdeployments API
type ZwhDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZwhDeploymentSpec   `json:"spec,omitempty"`
	Status ZwhDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZwhDeploymentList contains a list of ZwhDeployment
type ZwhDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZwhDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZwhDeployment{}, &ZwhDeploymentList{})
}
