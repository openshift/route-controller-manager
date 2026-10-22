package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const secretListPageSize int64 = 100

func listSecrets(ctx context.Context, client coreclient.SecretInterface, options metav1.ListOptions) (*corev1.SecretList, error) {
	options.Limit = secretListPageSize
	if options.ResourceVersion == "0" {
		options.ResourceVersion = ""
		options.ResourceVersionMatch = ""
	}

	var result *corev1.SecretList
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		page, err := client.List(ctx, options)
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = page
		} else {
			result.Items = append(result.Items, page.Items...)
		}
		if page.Continue == "" {
			result.Continue = ""
			result.RemainingItemCount = nil
			return result, nil
		}
		options.Continue = page.Continue
		options.ResourceVersion = ""
		options.ResourceVersionMatch = ""
	}
}

func newPaginatedSecretInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	secrets := client.CoreV1().Secrets(metav1.NamespaceAll)
	return cache.NewSharedIndexInformer(
		cache.ToListWatcherWithWatchListSemantics(&cache.ListWatch{
			ListWithContextFunc: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
				return listSecrets(ctx, secrets, options)
			},
			WatchFuncWithContext: secrets.Watch,
		}, client),
		&corev1.Secret{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
}
