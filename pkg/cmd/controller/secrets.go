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

// secretListPageSize bounds the number of Secrets requested in each LIST page.
const secretListPageSize int64 = 100

// listSecrets collects every continuation page into a complete Secret snapshot,
// returning no partial result if a request fails or cancellation is observed.
// Cancellation checks are best-effort: cancellation after the final check can
// still race successful completion.
func listSecrets(ctx context.Context, client coreclient.SecretInterface, options metav1.ListOptions) (*corev1.SecretList, error) {
	options.Limit = secretListPageSize
	// ResourceVersion "0" allows the watch cache to ignore the page limit. Use
	// an empty version to request pagination, also clearing its match constraint
	// because a match requires a version. Preserve explicit nonzero versions and matches.
	if options.ResourceVersion == "0" {
		options.ResourceVersion = ""
		options.ResourceVersionMatch = ""
	} else if options.Continue == "" && options.ResourceVersion != "" && options.ResourceVersionMatch == "" {
		// Preserve the reflector's unlimited relist minimum-version semantics:
		// adding a limit without a match would select an exact historical snapshot.
		options.ResourceVersionMatch = metav1.ResourceVersionMatchNotOlderThan
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
		if err := ctx.Err(); err != nil {
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
		// The continuation token carries the first page's snapshot version. As in
		// the client-go pager, clear version and match constraints to avoid invalid
		// combinations of continuation and resource-version options.
		options.ResourceVersion = ""
		options.ResourceVersionMatch = ""
	}
}

// newPaginatedSecretInformer uses paginated LIST snapshots while retaining
// watch-list initialization, ordinary watches, and namespace indexing.
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
