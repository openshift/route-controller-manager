package route

var ControllerManagerInitialization = map[string]InitFunc{
	"openshift.io/ingress-ip":       RunIngressIPController,
	"openshift.io/ingress-to-route": RunIngressToRouteController,
	"openshift.io/secret-injector":  RunRouteSecretInjectorController,
}

const (
	infraServiceIngressIPControllerServiceAccountName    = "service-ingress-ip-controller"
	InfraIngressToRouteControllerServiceAccountName      = "ingress-to-route-controller"
	infraRouteSecretInjectorControllerServiceAccountName = "route-secret-injector-controller"

	defaultOpenShiftInfraNamespace = "openshift-infra"
)
