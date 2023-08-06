package controller

import (
	"bytes"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"

	myAppsv1 "zwh.com/pkg/zwh-deployment/api/v1"
)

func parseTemplate(md *myAppsv1.ZwhDeployment, templateName string) ([]byte, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("controller/templates/%s", templateName))
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	if err := tmpl.Execute(b, md); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func NewDeployment(md *myAppsv1.ZwhDeployment) (*appsv1.Deployment, error) {
	content, err := parseTemplate(md, "deployment.yaml")
	if err != nil {
		return nil, err
	}
	deploy := new(appsv1.Deployment)
	if err := yaml.Unmarshal(content, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

func NewIngress(md *myAppsv1.ZwhDeployment) (*networkv1.Ingress, error) {
	content, err := parseTemplate(md, "ingress.yaml")
	if err != nil {
		return nil, err
	}
	ig := new(networkv1.Ingress)
	if err := yaml.Unmarshal(content, ig); err != nil {
		return nil, err
	}
	return ig, nil
}

func NewService(md *myAppsv1.ZwhDeployment) (*corev1.Service, error) {
	content, err := parseTemplate(md, "service.yaml")
	if err != nil {
		return nil, err
	}
	svc := new(corev1.Service)
	if err := yaml.Unmarshal(content, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func NewServiceNP(md *myAppsv1.ZwhDeployment) (*corev1.Service, error) {
	content, err := parseTemplate(md, "service.yaml")
	if err != nil {
		return nil, err
	}
	svc := new(corev1.Service)
	if err := yaml.Unmarshal(content, svc); err != nil {
		return nil, err
	}
	return svc, nil
}
