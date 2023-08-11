package v1

const (
	ModeIngress  = "ingress"
	ModeNodePort = "nodeport"
)

const (
	ConditionTypeDeployment = "Deployment"
	ConditionTypeService    = "Service"
	ConditionTypeIngress    = "Ingress"

	ConditionMessageDeploymentOKFmt  = "Deployment %s is ready"
	ConditionMessageDeploymentNotFmt = "Deployment %s is not ready"
	ConditionMessageServiceOKFmt     = "Service %s is ready"
	ConditionMessageServiceNotFmt    = "Service %s is not ready"
	ConditionMessageIngressOKFmt     = "Ingress %s is ready"
	ConditionMessageIngressNotFmt    = "Ingress %s is not ready"

	ConditionReasonDeploymentReady    = "DeploymentReady"
	ConditionReasonDeploymentNotReady = "DeploymentNotReady"
	ConditionReasonServiceReady       = "ServiceReady"
	ConditionReasonServiceNotReady    = "ServiceNotReady"
	ConditionReasonIngressReady       = "IngressReady"
	ConditionReasonIngressNotReady    = "IngressNotReady"
	ConditionStatusTrue               = "True"
	ConditionStatusFalse              = "False"
)

const (
	StatusReasonSuccess  = "Success"
	StatusMessageSuccess = "Success"
	StatusPhaseComplete  = "Complete"
)
