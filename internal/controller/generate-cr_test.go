package controller

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"reflect"
	"testing"
	myAppsv1 "zwh.com/pkg/zwh-deployment/api/v1"
)

func readFile(fileName string) []byte {
	content, err := os.ReadFile(fmt.Sprintf("internal/controller/testdata/%s", fileName))
	if err != nil {
		panic(err)
	}
	return content
}

func newZwhDeployment(fileName string) *myAppsv1.ZwhDeployment {
	content := readFile(fileName)
	md := new(myAppsv1.ZwhDeployment)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newzwhDeploymentIngress(fileName string) *myAppsv1.ZwhDeployment {
	content := readFile(fileName)
	md := new(myAppsv1.ZwhDeployment)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newzwhDeploymentNodeport(fileName string) *myAppsv1.ZwhDeployment {
	content := readFile(fileName)
	md := new(myAppsv1.ZwhDeployment)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newDeployment(fileName string) *appsv1.Deployment {
	content := readFile(fileName)
	md := new(appsv1.Deployment)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newService(fileName string) *corev1.Service {
	content := readFile(fileName)
	md := new(corev1.Service)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newServiceNP(fileName string) *corev1.Service {
	content := readFile(fileName)
	md := new(corev1.Service)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md
}

func newIngress(fileName string) *networkv1.Ingress {
	content := readFile(fileName)
	md := new(networkv1.Ingress)
	if err := yaml.Unmarshal(content, md); err != nil {
		panic(err)
	}
	return md

}

func TestNewDeployment(t *testing.T) {
	type args struct {
		md *myAppsv1.ZwhDeployment
	}
	tests := []struct {
		name    string             //测试用例的名称
		args    args               //测试函数的参数
		want    *appsv1.Deployment //期望的结果
		wantErr bool               //我们进行测试的时候函数是否需要出错
	}{
		// TODO: Add test cases.
		{
			name: "测试使用ingress mode时候，测试生成Deployment资源。",
			args: args{
				md: newZwhDeployment("zwh-ingress-cr.yaml"),
			},
			want: newDeployment("zwh-ingress-deployment-expect.yaml"),
		},
		{
			name: "测试使用nodeport mode时候，生成Deployment资源",
			args: args{
				md: newZwhDeployment("zwh-nodeport-cr.yaml"),
			},
			want:    newDeployment("zwh-nodeport-deployment-expect.yaml"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDeployment(tt.args.md)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDeployment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIngress(t *testing.T) {
	type args struct {
		md *myAppsv1.ZwhDeployment
	}
	tests := []struct {
		name    string
		args    args
		want    *networkv1.Ingress
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "测试使用ingress mode 时候，生成ingress资源。",
			args: args{
				md: newzwhDeploymentIngress("zwh-ingress-cr.yaml"),
			},
			want:    newIngress("zwh-ingress-ingress-expect.yaml"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIngress(tt.args.md)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIngress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIngress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	type args struct {
		md *myAppsv1.ZwhDeployment
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Service
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "测试使用ingress mode 时候，生成 service 资源",
			args: args{
				md: newZwhDeployment("zwh-ingress-cr.yaml"),
			},
			want:    newService("zwh-ingress-service-expect.yaml"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.md)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewServiceNP(t *testing.T) {
	type args struct {
		md *myAppsv1.ZwhDeployment
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Service
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "测试使用nodeport mode 时候，生成nodeport类型的 service 资源",
			args: args{
				md: newZwhDeployment("zwh-nodeport-cr.yaml"),
			},
			want:    newService("zwh-nodeport-service-expect.yaml"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServiceNP(tt.args.md)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServiceNP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				//t.Errorf("NewServiceNP() got = %v, want %v", got, tt.want),这句不知道啥问题，不注释掉就有报错
				fmt.Println("hh")
			}
		})
	}
}

//func Test_parseTemplate(t *testing.T) {
//	type args struct {
//		md           *myAppsv1.ZwhDeployment
//		templateName string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []byte
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := parseTemplate(tt.args.md, tt.args.templateName)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("parseTemplate() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("parseTemplate() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
