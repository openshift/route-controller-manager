package route

import (
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	secretinjector "github.com/openshift/route-controller-manager/pkg/route/secret-injector"
)

func RunRouteSecretInjectorController(ctx *ControllerContext) (bool, error) {
	clientConfig := ctx.ClientBuilder.ConfigOrDie(infraRouteSecretInjectorControllerServiceAccountName)
	coreClient, err := coreclient.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}
	routeClient, err := routeclient.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}

	controller := secretinjector.NewController(
		coreClient,
		routeClient,
		ctx.KubernetesInformers.Core().V1().Secrets(),
		ctx.RouteInformers.Route().V1().Routes(),
	)

	go controller.Run(5, ctx.Stop)

	return true, nil
}
