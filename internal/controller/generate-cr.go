package controller

import (
	"bytes"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"

	myAppsv1 "zwh.com/pkg/zwh-deployment/api/v1"
)

func parseTemplate(md *myAppsv1.ZwhDeployment, templateName string) ([]byte, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("internal/controller/templates/%s", templateName))
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

// NewIssuer 实现创建issuer资源对象
func NewIssuer(md *myAppsv1.ZwhDeployment) (*unstructured.Unstructured, error) {
	if md.Spec.Expose.Mode != myAppsv1.ModeIngress ||
		!md.Spec.Expose.Tls {
		return nil, nil
	}
	// Sample
	//apiVersion: cert-manager.io/v1
	//kind: Issuer
	//metadata:
	//  name: selfsigned-issuer
	//spec:
	//  selfSigned: {}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
			"metadata": map[string]interface{}{
				"name":      md.Name,
				"namespace": md.Namespace,
			},
			"spec": map[string]interface{}{
				"selfSigned": map[string]interface{}{},
			},
		},
	}, nil
}

// NewCert 实现创建certificate资源
func NewCert(md *myAppsv1.ZwhDeployment) (*unstructured.Unstructured, error) {
	if md.Spec.Expose.Mode != myAppsv1.ModeIngress ||
		!md.Spec.Expose.Tls {
		return nil, nil
	}
	// Sample
	//apiVersion: cert-manager.io/v1
	//kind: Certificate
	//metadata:
	//  name: serving-cert  # this name should match the one appeared in kustomizeconfig.yaml
	//  namespace: system
	//spec:
	//  dnsNames:
	//  - <spec.expose.ingressDomain>
	//  issuerRef:
	//    kind: Issuer
	//    name: selfsigned-issuer
	//  secretName: webhook-server-cert
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      md.Name,
				"namespace": md.Namespace,
			},
			"spec": map[string]interface{}{
				"dnsNames": []interface{}{
					md.Spec.Expose.IngressDomain,
				},
				"issuerRef": map[string]interface{}{
					"kind": "Issuer",
					"name": md.Name,
				},
				"secretName": md.Name,
			},
		},
	}, nil
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
