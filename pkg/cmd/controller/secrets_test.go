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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	openshiftcontrolplanev1 "github.com/openshift/api/openshiftcontrolplane/v1"
)

type secretRequest struct {
	path  string
	query url.Values
}

type secretRequests struct {
	mutex sync.Mutex
	items []secretRequest
}

func (requests *secretRequests) snapshot() []secretRequest {
	requests.mutex.Lock()
	defer requests.mutex.Unlock()
	return append([]secretRequest(nil), requests.items...)
}

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

func writeSecretResponse(test *testing.T, writer http.ResponseWriter, object interface{}) {
	test.Helper()
	if status, ok := object.(*metav1.Status); ok {
		status = status.DeepCopy()
		status.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Status"}
		object = status
	}
	assert.NoError(test, json.NewEncoder(writer).Encode(object))
}

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

type cancellingSecretClient struct {
	coreclient.SecretInterface
	cancel context.CancelFunc
	calls  int
}

func (client *cancellingSecretClient) List(ctx context.Context, options metav1.ListOptions) (*corev1.SecretList, error) {
	client.calls++
	client.cancel()
	return secretPage(0, 1, "next"), nil
}

func TestListSecretsCancellation(test *testing.T) {
	test.Parallel()
	for _, alreadyCancelled := range []bool{false, true} {
		test.Run(fmt.Sprintf("already cancelled=%t", alreadyCancelled), func(test *testing.T) {
			test.Parallel()
			ctx, cancel := context.WithCancel(test.Context())
			defer cancel()
			client := &cancellingSecretClient{cancel: cancel}
			if alreadyCancelled {
				cancel()
			}
			result, err := listSecrets(ctx, client, metav1.ListOptions{})
			require.ErrorIs(test, err, context.Canceled)
			assert.Nil(test, result)
			if alreadyCancelled {
				assert.Zero(test, client.calls)
			} else {
				assert.Equal(test, 1, client.calls)
			}
		})
	}
}

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
