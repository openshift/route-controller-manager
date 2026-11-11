package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metainternalversionscheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metainternalversionvalidation "k8s.io/apimachinery/pkg/apis/meta/internalversion/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
)

// secretRequest records a test client's request path and query options.
type secretRequest struct {
	path  string
	query url.Values
}

// secretRequests synchronizes request recording across HTTP handlers and tests.
type secretRequests struct {
	mutex sync.Mutex
	items []secretRequest
}

// snapshot copies the recorded requests so tests can inspect them without racing handlers.
func (requests *secretRequests) snapshot() []secretRequest {
	requests.mutex.Lock()
	defer requests.mutex.Unlock()
	return append([]secretRequest(nil), requests.items...)
}

// newSecretTestClient connects a real Kubernetes client to a recording HTTP server
// that applies API list-option decoding, defaults, and validation before the handler.
func newSecretTestClient(test *testing.T, handler http.HandlerFunc) (*kubernetes.Clientset, *rest.Config, *secretRequests) {
	test.Helper()
	requests := &secretRequests{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.mutex.Lock()
		requests.items = append(requests.items, secretRequest{path: request.URL.Path, query: request.URL.Query()})
		requests.mutex.Unlock()
		writer.Header().Set("Content-Type", "application/json")
		options := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(request.URL.Query(), metav1.SchemeGroupVersion, &options); err != nil {
			status := apierrors.NewBadRequest(err.Error()).ErrStatus
			writer.WriteHeader(int(status.Code))
			writeSecretResponse(test, writer, &status)
			return
		}
		metainternalversion.SetListOptionsDefaults(&options, true)
		if errors := metainternalversionvalidation.ValidateListOptions(&options, true); len(errors) > 0 {
			status := apierrors.NewInvalid(schema.GroupKind{Group: metav1.GroupName, Kind: "ListOptions"}, "", errors).ErrStatus
			writer.WriteHeader(int(status.Code))
			writeSecretResponse(test, writer, &status)
			return
		}
		handler(writer, request)
	}))
	test.Cleanup(server.Close)
	config := &rest.Config{Host: server.URL}
	client, err := kubernetes.NewForConfig(config)
	require.NoError(test, err)
	return client, config, requests
}

// writeSecretResponse encodes a test API response, including type metadata for errors.
func writeSecretResponse(test *testing.T, writer http.ResponseWriter, object interface{}) {
	test.Helper()
	if status, ok := object.(*metav1.Status); ok {
		status = status.DeepCopy()
		status.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Status"}
		object = status
	}
	assert.NoError(test, json.NewEncoder(writer).Encode(object))
}

// secretPage builds a deterministic Secret page with shared snapshot metadata
// and alternating namespaces for pagination and indexing assertions.
func secretPage(start, count int, continueToken string) *corev1.SecretList {
	page := &corev1.SecretList{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "SecretList"},
		ListMeta: metav1.ListMeta{
			ResourceVersion:    "42",
			Continue:           continueToken,
			RemainingItemCount: ptr.To(int64(100)),
		},
		Items: []corev1.Secret{},
	}
	for index := start; index < start+count; index++ {
		page.Items = append(page.Items, corev1.Secret{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       fmt.Sprintf("namespace-%d", index%2),
				Name:            fmt.Sprintf("secret-%d", index),
				ResourceVersion: "42",
			},
		})
	}
	return page
}

// TestListSecrets verifies complete pagination, list-option handling, and rejection
// of partial snapshots when any page request fails.
func TestListSecrets(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name                 string
		pageSizes            []int
		resourceVersion      string
		resourceVersionMatch metav1.ResourceVersionMatch
		failurePage          int
		failure              *apierrors.StatusError
	}{
		{name: "empty", pageSizes: []int{0}},
		{name: "single page", pageSizes: []int{7}},
		{name: "full final page", pageSizes: []int{100}},
		{name: "multiple pages", pageSizes: []int{100, 100, 5}},
		{name: "empty intermediate page", pageSizes: []int{100, 0, 5}},
		{name: "normalize zero version", pageSizes: []int{100, 5}, resourceVersion: "0", resourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan},
		{name: "normalize zero version without match", pageSizes: []int{100, 5}, resourceVersion: "0"},
		{name: "preserve implicit minimum version", pageSizes: []int{100, 5}, resourceVersion: "10"},
		{name: "preserve explicit version", pageSizes: []int{100, 5}, resourceVersion: "10", resourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan},
		{name: "preserve exact version", pageSizes: []int{100, 5}, resourceVersion: "10", resourceVersionMatch: metav1.ResourceVersionMatchExact},
		{name: "first page failure", pageSizes: []int{0}, failure: apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "", fmt.Errorf("denied"))},
		{name: "later page failure", pageSizes: []int{100, 5}, failurePage: 1, failure: apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "", fmt.Errorf("denied"))},
		{name: "expired continuation", pageSizes: []int{100, 5}, failurePage: 1, failure: apierrors.NewResourceExpired("expired continuation")},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			pageIndex, total := 0, 0
			client, _, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
				if testCase.failure != nil && pageIndex == testCase.failurePage {
					writer.WriteHeader(int(testCase.failure.ErrStatus.Code))
					writeSecretResponse(test, writer, &testCase.failure.ErrStatus)
					return
				}
				if pageIndex >= len(testCase.pageSizes) {
					http.Error(writer, "unexpected extra page", http.StatusBadRequest)
					return
				}
				continueToken := ""
				if pageIndex < len(testCase.pageSizes)-1 {
					continueToken = fmt.Sprintf("page-%d", pageIndex+1)
				}
				page := secretPage(total, testCase.pageSizes[pageIndex], continueToken)
				total += testCase.pageSizes[pageIndex]
				pageIndex++
				writeSecretResponse(test, writer, page)
			})
			options := metav1.ListOptions{
				Limit:                500,
				ResourceVersion:      testCase.resourceVersion,
				LabelSelector:        "example=selected",
				FieldSelector:        "metadata.namespace=namespace-0",
				TimeoutSeconds:       ptr.To(int64(30)),
				ResourceVersionMatch: testCase.resourceVersionMatch,
			}
			originalOptions := *options.DeepCopy()
			result, err := listSecrets(test.Context(), client.CoreV1().Secrets(metav1.NamespaceAll), options)
			require.Equal(test, originalOptions, options)
			if testCase.failure != nil {
				require.Error(test, err)
				var status apierrors.APIStatus
				require.ErrorAs(test, err, &status)
				assert.Equal(test, testCase.failure.ErrStatus.Reason, status.Status().Reason)
				assert.Nil(test, result)
			} else {
				require.NoError(test, err)
				expectedCount := 0
				for _, count := range testCase.pageSizes {
					expectedCount += count
				}
				require.Len(test, result.Items, expectedCount)
				for index, secret := range result.Items {
					assert.Equal(test, fmt.Sprintf("secret-%d", index), secret.Name)
					assert.Equal(test, fmt.Sprintf("namespace-%d", index%2), secret.Namespace)
				}
				assert.Equal(test, "42", result.ResourceVersion)
				assert.Empty(test, result.Continue)
				assert.Nil(test, result.RemainingItemCount)
			}
			recorded := requests.snapshot()
			expectedRequests := len(testCase.pageSizes)
			if testCase.failure != nil {
				expectedRequests = testCase.failurePage + 1
			}
			require.Len(test, recorded, expectedRequests)
			for index, request := range recorded {
				assert.Equal(test, "/api/v1/secrets", request.path)
				assert.Equal(test, "100", request.query.Get("limit"))
				assert.Equal(test, options.LabelSelector, request.query.Get("labelSelector"))
				assert.Equal(test, options.FieldSelector, request.query.Get("fieldSelector"))
				assert.Equal(test, "30", request.query.Get("timeoutSeconds"))
				if index == 0 {
					assert.Empty(test, request.query.Get("continue"))
					expectedVersion := testCase.resourceVersion
					expectedMatch := testCase.resourceVersionMatch
					if expectedVersion == "0" {
						expectedVersion = ""
						expectedMatch = ""
					} else if expectedVersion != "" && expectedMatch == "" {
						expectedMatch = metav1.ResourceVersionMatchNotOlderThan
					}
					assert.Equal(test, expectedVersion, request.query.Get("resourceVersion"))
					assert.Equal(test, string(expectedMatch), request.query.Get("resourceVersionMatch"))
				} else {
					assert.Equal(test, fmt.Sprintf("page-%d", index), request.query.Get("continue"))
					assert.Empty(test, request.query.Get("resourceVersion"))
					assert.Empty(test, request.query.Get("resourceVersionMatch"))
				}
			}
		})
	}
}

// secretRelistWatcher disables streaming initialization so tests exercise actual reflector relists.
type secretRelistWatcher struct {
	*cache.ListWatch
}

// IsWatchListSemanticsUnSupported forces the reflector's ordinary LIST/watch path.
func (*secretRelistWatcher) IsWatchListSemanticsUnSupported() bool { return true }

// TestSecretReflectorRelist preserves the baseline minimum-version interpretation
// using a real reflector, HTTP client, API validation, and storage revision selection.
func TestSecretReflectorRelist(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name        string
		baseline    bool
		match       metav1.ResourceVersionMatch
		wantVersion string
	}{
		{name: "baseline unlimited relist", baseline: true, wantVersion: "50"},
		{name: "paginated implicit minimum", wantVersion: "50"},
		{name: "explicit minimum", match: metav1.ResourceVersionMatchNotOlderThan, wantVersion: "50"},
		{name: "explicit exact", match: metav1.ResourceVersionMatchExact, wantVersion: "42"},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			var mutex sync.Mutex
			var revisions []int64
			var serverErrors []error
			client, _, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()
				options := metainternalversion.ListOptions{}
				if err := metainternalversionscheme.ParameterCodec.DecodeParameters(request.URL.Query(), metav1.SchemeGroupVersion, &options); err != nil {
					serverErrors = append(serverErrors, err)
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
				revision, _, err := storage.ValidateListOptions("/registry/secrets/", storage.APIObjectVersioner{}, storage.ListOptions{
					ResourceVersion: options.ResourceVersion, ResourceVersionMatch: options.ResourceVersionMatch,
					Recursive: true, Predicate: storage.SelectionPredicate{Limit: options.Limit, Continue: options.Continue},
				})
				if err != nil {
					serverErrors = append(serverErrors, err)
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
				revisions = append(revisions, revision)
				version := "42"
				continueToken := ""
				if len(revisions) > 1 {
					if revision == 0 || revision == 50 {
						version = "50"
					}
					if options.Limit > 0 && options.Continue == "" {
						snapshotRevision := int64(42)
						if version == "50" {
							snapshotRevision = 50
						}
						continueToken, err = storage.EncodeContinue("/registry/secrets/next", "/registry/secrets/", snapshotRevision)
						if err != nil {
							serverErrors = append(serverErrors, err)
							http.Error(writer, err.Error(), http.StatusInternalServerError)
							return
						}
					}
				}
				page := secretPage(0, 1, continueToken)
				page.ResourceVersion = version
				writeSecretResponse(test, writer, page)
			})
			var cancel context.CancelFunc
			lister := &secretRelistWatcher{ListWatch: &cache.ListWatch{
				ListWithContextFunc: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
					if testCase.baseline {
						return client.CoreV1().Secrets("").List(ctx, options)
					}
					if options.ResourceVersion != "" && options.ResourceVersion != "0" {
						options.ResourceVersionMatch = testCase.match
					}
					return listSecrets(ctx, client.CoreV1().Secrets(""), options)
				},
				WatchFuncWithContext: func(context.Context, metav1.ListOptions) (watch.Interface, error) {
					cancel()
					return watch.NewEmptyWatch(), nil
				},
			}}
			reflector := cache.NewReflector(lister, &corev1.Secret{}, cache.NewStore(cache.MetaNamespaceKeyFunc), 0)
			for pass := 0; pass < 2; pass++ {
				ctx, stop := context.WithTimeout(test.Context(), 5*time.Second)
				cancel = stop
				err := reflector.ListAndWatchWithContext(ctx)
				stop()
				require.NoError(test, err)
			}
			mutex.Lock()
			defer mutex.Unlock()
			require.Empty(test, serverErrors)
			recorded := requests.snapshot()
			if testCase.baseline {
				require.Equal(test, []int64{0, 0}, revisions)
				assert.Empty(test, recorded[1].query.Get("limit"))
			} else {
				wantRevision := int64(0)
				wantContinuationRevision := int64(50)
				if testCase.match == metav1.ResourceVersionMatchExact {
					wantRevision = 42
					wantContinuationRevision = 42
				}
				require.Equal(test, []int64{0, wantRevision, wantContinuationRevision}, revisions)
				assert.Equal(test, "100", recorded[1].query.Get("limit"))
				assert.NotEmpty(test, recorded[2].query.Get("continue"))
				assert.Empty(test, recorded[2].query.Get("resourceVersion"))
				assert.Empty(test, recorded[2].query.Get("resourceVersionMatch"))
			}
			assert.Equal(test, "42", recorded[1].query.Get("resourceVersion"))
			assert.Equal(test, testCase.wantVersion, reflector.LastSyncResourceVersion())
		})
	}
}

// cancellingSecretClient cancels the listing context while returning a configured page or error.
type cancellingSecretClient struct {
	coreclient.SecretInterface
	cancel        context.CancelFunc
	calls         int
	firstPage     *corev1.SecretList
	continueToken string
	failure       error
}

// List returns an optional initial page, then cancels before returning the configured final result.
func (client *cancellingSecretClient) List(ctx context.Context, options metav1.ListOptions) (*corev1.SecretList, error) {
	client.calls++
	if client.calls == 1 && client.firstPage != nil {
		return client.firstPage, nil
	}
	client.cancel()
	return secretPage(0, 1, client.continueToken), client.failure
}

// TestListSecretsCancellation verifies observed cancellation discards even a successful
// final page, while a concurrent request failure retains precedence.
func TestListSecretsCancellation(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		name             string
		alreadyCancelled bool
		firstPage        *corev1.SecretList
		continueToken    string
		failure          error
	}{
		{name: "before request", alreadyCancelled: true},
		{name: "between pages", continueToken: "next"},
		{name: "successful final page"},
		{name: "successful later final page", firstPage: secretPage(0, 100, "next")},
		{name: "request error wins", failure: apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "", fmt.Errorf("denied"))},
		{name: "later request error wins", firstPage: secretPage(0, 100, "next"), failure: apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "", fmt.Errorf("denied"))},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			ctx, cancel := context.WithCancel(test.Context())
			defer cancel()
			client := &cancellingSecretClient{cancel: cancel, firstPage: testCase.firstPage, continueToken: testCase.continueToken, failure: testCase.failure}
			if testCase.alreadyCancelled {
				cancel()
			}
			result, err := listSecrets(ctx, client, metav1.ListOptions{})
			if testCase.failure != nil {
				require.ErrorIs(test, err, testCase.failure)
			} else {
				require.ErrorIs(test, err, context.Canceled)
			}
			assert.Nil(test, result)
			if testCase.alreadyCancelled {
				assert.Zero(test, client.calls)
			} else if testCase.firstPage != nil {
				assert.Equal(test, 2, client.calls)
			} else {
				assert.Equal(test, 1, client.calls)
			}
		})
	}
}

// TestSecretFactoryDoesNotSyncCanceledSnapshot verifies that a successful final page
// discarded for observed cancellation cannot publish objects through the shared factory.
func TestSecretFactoryDoesNotSyncCanceledSnapshot(test *testing.T) {
	test.Parallel()
	factory := informers.NewSharedInformerFactory(nil, 0)
	informer := factory.InformerFor(&corev1.Secret{}, func(kubernetes.Interface, time.Duration) cache.SharedIndexInformer {
		return cache.NewSharedIndexInformer(&secretRelistWatcher{ListWatch: &cache.ListWatch{
			ListWithContextFunc: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
				listCtx, cancel := context.WithCancel(ctx)
				defer cancel()
				return listSecrets(listCtx, &cancellingSecretClient{cancel: cancel, firstPage: secretPage(0, 100, "next")}, options)
			},
			WatchFuncWithContext: func(context.Context, metav1.ListOptions) (watch.Interface, error) {
				return nil, fmt.Errorf("unexpected watch after canceled snapshot")
			},
		}}, &corev1.Secret{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	})
	listErrors := make(chan error, 1)
	require.NoError(test, informer.SetWatchErrorHandlerWithContext(func(ctx context.Context, reflector *cache.Reflector, err error) {
		select {
		case listErrors <- err:
		default:
		}
	}))
	ctx, cancel := context.WithTimeout(test.Context(), 5*time.Second)
	defer cancel()
	factory.Start(ctx.Done())
	test.Cleanup(func() {
		cancel()
		factory.Shutdown()
	})
	select {
	case err := <-listErrors:
		require.ErrorIs(test, err, context.Canceled)
	case <-ctx.Done():
		test.Fatal("informer did not report the canceled snapshot")
	}
	assert.False(test, informer.HasSynced())
	assert.Empty(test, informer.GetStore().List())
}

// TestSecretInformerFactoryPagination verifies that the shared factory reuses the
// paginated Secret informer while leaving Service lists and ordinary watches unchanged.
func TestSecretInformerFactoryPagination(test *testing.T) {
	test.Parallel()
	_, config, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("watch") == "true" {
			if request.URL.Query().Get("sendInitialEvents") == "true" {
				writer.WriteHeader(http.StatusMethodNotAllowed)
				writeSecretResponse(test, writer, &metav1.Status{Status: metav1.StatusFailure, Reason: metav1.StatusReasonMethodNotAllowed, Code: http.StatusMethodNotAllowed})
				return
			}
			writer.WriteHeader(http.StatusOK)
			writer.(http.Flusher).Flush()
			<-request.Context().Done()
			return
		}
		if request.URL.Path == "/api/v1/services" {
			writeSecretResponse(test, writer, &corev1.ServiceList{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceList"}, ListMeta: metav1.ListMeta{ResourceVersion: "42"}})
			return
		}
		switch request.URL.Query().Get("continue") {
		case "":
			writeSecretResponse(test, writer, secretPage(0, 100, "second"))
		case "second":
			writeSecretResponse(test, writer, secretPage(100, 100, "third"))
		case "third":
			writeSecretResponse(test, writer, secretPage(200, 5, ""))
		default:
			http.Error(writer, "unexpected continuation", http.StatusBadRequest)
		}
	})
	ctx, cancel := context.WithTimeout(test.Context(), 10*time.Second)
	defer cancel()
	controllerContext, err := NewControllerContext(ctx, &controllercmd.ControllerContext{KubeConfig: config}, openshiftcontrolplanev1.OpenShiftControllerManagerConfig{Controllers: []string{"*"}})
	require.NoError(test, err)
	factory := controllerContext.KubernetesInformers
	secretInformer := factory.Core().V1().Secrets()
	assert.Same(test, secretInformer.Informer(), factory.InformerFor(&corev1.Secret{}, nil))
	serviceInformer := factory.Core().V1().Services().Informer()
	controllerContext.StartInformers(ctx.Done())
	test.Cleanup(func() {
		cancel()
		factory.Shutdown()
	})
	require.True(test, cache.WaitForCacheSync(ctx.Done(), secretInformer.Informer().HasSynced, serviceInformer.HasSynced))
	secrets, err := secretInformer.Lister().List(labels.Everything())
	require.NoError(test, err)
	require.Len(test, secrets, 205)
	namespaceSecrets, err := secretInformer.Lister().Secrets("namespace-0").List(labels.Everything())
	require.NoError(test, err)
	assert.Len(test, namespaceSecrets, 103)
	require.Eventually(test, func() bool {
		for _, request := range requests.snapshot() {
			if request.path == "/api/v1/secrets" && request.query.Get("watch") == "true" && request.query.Get("sendInitialEvents") == "" {
				return true
			}
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)
	var secretLists, serviceLists []secretRequest
	for _, request := range requests.snapshot() {
		if request.query.Get("watch") == "true" {
			assert.Empty(test, request.query.Get("limit"))
			assert.Empty(test, request.query.Get("continue"))
			if request.query.Get("sendInitialEvents") == "" {
				assert.Equal(test, "42", request.query.Get("resourceVersion"))
			}
			continue
		}
		if request.path == "/api/v1/secrets" {
			secretLists = append(secretLists, request)
			assert.Equal(test, "100", request.query.Get("limit"))
		} else if request.path == "/api/v1/services" {
			serviceLists = append(serviceLists, request)
			assert.Equal(test, "500", request.query.Get("limit"))
			assert.Equal(test, "0", request.query.Get("resourceVersion"))
		}
	}
	require.Len(test, secretLists, 3)
	assert.Empty(test, secretLists[0].query.Get("resourceVersion"))
	assert.Empty(test, secretLists[0].query.Get("continue"))
	assert.Equal(test, "second", secretLists[1].query.Get("continue"))
	assert.Equal(test, "third", secretLists[2].query.Get("continue"))
	require.Len(test, serviceLists, 1)
}

// TestSecretInformerWatchList verifies streaming initialization and live updates
// populate the cache without falling back to paginated LIST requests.
func TestSecretInformerWatchList(test *testing.T) {
	test.Parallel()
	client, _, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("sendInitialEvents") != "true" {
			http.Error(writer, "expected streaming initialization", http.StatusBadRequest)
			return
		}
		page := secretPage(0, 2, "")
		for _, secret := range page.Items {
			writeSecretResponse(test, writer, map[string]interface{}{"type": watch.Added, "object": secret})
		}
		bookmark := &corev1.Secret{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}, ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "42",
			Annotations:     map[string]string{metav1.InitialEventsAnnotationKey: "true"},
		}}
		writeSecretResponse(test, writer, map[string]interface{}{"type": watch.Bookmark, "object": bookmark})
		liveSecret := secretPage(2, 1, "").Items[0]
		liveSecret.ResourceVersion = "43"
		writeSecretResponse(test, writer, map[string]interface{}{"type": watch.Added, "object": liveSecret})
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	})
	ctx, cancel := context.WithTimeout(test.Context(), 10*time.Second)
	defer cancel()
	informer := newPaginatedSecretInformer(client, 10*time.Minute)
	done := make(chan struct{})
	go func() {
		defer close(done)
		informer.RunWithContext(ctx)
	}()
	test.Cleanup(func() {
		cancel()
		<-done
	})
	require.True(test, cache.WaitForCacheSync(ctx.Done(), informer.HasSynced))
	require.Eventually(test, func() bool { return len(informer.GetStore().List()) == 3 }, 5*time.Second, 10*time.Millisecond)
	recorded := requests.snapshot()
	require.Len(test, recorded, 1)
	assert.Equal(test, "/api/v1/secrets", recorded[0].path)
	assert.Equal(test, "true", recorded[0].query.Get("watch"))
	assert.Equal(test, "true", recorded[0].query.Get("sendInitialEvents"))
	assert.Equal(test, "true", recorded[0].query.Get("allowWatchBookmarks"))
	assert.Equal(test, "NotOlderThan", recorded[0].query.Get("resourceVersionMatch"))
	assert.Empty(test, recorded[0].query.Get("limit"))
	assert.Empty(test, recorded[0].query.Get("continue"))
}

// TestSecretInformerDisabled verifies that disabling ingress-to-route prevents
// Secret informer registration and requests.
func TestSecretInformerDisabled(test *testing.T) {
	test.Parallel()
	for _, controllers := range [][]string{nil, {"*", "-openshift.io/ingress-to-route"}} {
		test.Run(fmt.Sprintf("controllers=%v", controllers), func(test *testing.T) {
			test.Parallel()
			_, config, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
				http.Error(writer, "no informer should start", http.StatusBadRequest)
			})
			ctx, cancel := context.WithTimeout(test.Context(), time.Second)
			defer cancel()
			controllerContext, err := NewControllerContext(ctx, &controllercmd.ControllerContext{KubeConfig: config}, openshiftcontrolplanev1.OpenShiftControllerManagerConfig{Controllers: controllers})
			require.NoError(test, err)
			controllerContext.StartInformers(ctx.Done())
			test.Cleanup(func() {
				cancel()
				controllerContext.KubernetesInformers.Shutdown()
			})
			assert.Empty(test, controllerContext.KubernetesInformers.WaitForCacheSync(ctx.Done()))
			assert.Empty(test, requests.snapshot())
		})
	}
}

// TestListSecretsInFlightCancellation verifies that cancellation interrupts an
// active HTTP request and returns no Secret snapshot.
func TestListSecretsInFlightCancellation(test *testing.T) {
	test.Parallel()
	started := make(chan struct{})
	requestCancelled := make(chan struct{})
	client, _, _ := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
		close(started)
		<-request.Context().Done()
		close(requestCancelled)
	})
	ctx, cancel := context.WithTimeout(test.Context(), 5*time.Second)
	defer cancel()
	type outcome struct {
		result *corev1.SecretList
		err    error
	}
	finished := make(chan outcome, 1)
	go func() {
		result, err := listSecrets(ctx, client.CoreV1().Secrets(metav1.NamespaceAll), metav1.ListOptions{})
		finished <- outcome{result: result, err: err}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		test.Fatal("Secret LIST request did not start")
	}
	cancel()
	select {
	case actual := <-finished:
		require.ErrorIs(test, actual.err, context.Canceled)
		assert.Nil(test, actual.result)
	case <-time.After(5 * time.Second):
		test.Fatal("Secret LIST request did not stop after cancellation")
	}
	select {
	case <-requestCancelled:
	case <-time.After(5 * time.Second):
		test.Fatal("server did not observe request cancellation")
	}
}

// TestSecretInformerDoesNotSyncPartialList verifies that a later-page failure
// leaves the informer unsynced and its cache empty.
func TestSecretInformerDoesNotSyncPartialList(test *testing.T) {
	test.Parallel()
	client, _, requests := newSecretTestClient(test, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("sendInitialEvents") == "true" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			writeSecretResponse(test, writer, &metav1.Status{Status: metav1.StatusFailure, Reason: metav1.StatusReasonMethodNotAllowed, Code: http.StatusMethodNotAllowed})
			return
		}
		if request.URL.Query().Get("continue") == "" {
			writeSecretResponse(test, writer, secretPage(0, 100, "second"))
			return
		}
		writer.WriteHeader(http.StatusForbidden)
		writeSecretResponse(test, writer, &metav1.Status{Status: metav1.StatusFailure, Reason: metav1.StatusReasonForbidden, Code: http.StatusForbidden})
	})
	ctx, cancel := context.WithTimeout(test.Context(), 5*time.Second)
	defer cancel()
	informer := newPaginatedSecretInformer(client, 0)
	listErrors := make(chan error, 1)
	require.NoError(test, informer.SetWatchErrorHandlerWithContext(func(ctx context.Context, reflector *cache.Reflector, err error) {
		select {
		case listErrors <- err:
		default:
		}
	}))
	done := make(chan struct{})
	go func() {
		defer close(done)
		informer.RunWithContext(ctx)
	}()
	test.Cleanup(func() {
		cancel()
		<-done
	})
	select {
	case err := <-listErrors:
		require.True(test, apierrors.IsForbidden(err), "expected later-page error, got %v", err)
	case <-ctx.Done():
		test.Fatal("informer did not report the later-page failure")
	}
	assert.False(test, informer.HasSynced())
	assert.Empty(test, informer.GetStore().List())
	for _, request := range requests.snapshot() {
		if request.query.Get("watch") != "true" {
			assert.Equal(test, "100", request.query.Get("limit"))
		}
	}
}
