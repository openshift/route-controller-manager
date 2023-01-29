package secretinjector

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	kv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	routelisters "github.com/openshift/client-go/route/listers/route/v1"
)

const secretReferenceAnnotation = ""

type Controller struct {
	eventRecorder record.EventRecorder

	routeClient routeclient.RoutesGetter

	secretLister corelisters.SecretLister
	routeLister  routelisters.RouteLister

	// syncs are the items that must return true before the queue can be processed
	syncs []cache.InformerSynced

	// queue is the list of namespace keys that must be synced.
	queue workqueue.RateLimitingInterface

	// expectations track upcoming route creations that we have not yet observed
	expectations *expectations
	// expectationDelay controls how long the controller waits to observe its
	// own creates. Exposed only for testing.
	expectationDelay time.Duration

	// Prometheus metrics
	metricsCreated    bool
	metricsCreateOnce sync.Once
	metricsCreateLock sync.RWMutex
}

// TODO: remove expectations
// expectations track an upcoming change to a named resource related
// to an ingress. This is a thread safe object but callers assume
// responsibility for ensuring expectations do not leak.
type expectations struct {
	lock   sync.Mutex
	expect map[queueKey]sets.String
}

// newExpectations returns a tracking object for upcoming events
// that the controller may expect to happen.
func newExpectations() *expectations {
	return &expectations{
		expect: make(map[queueKey]sets.String),
	}
}

// Expect that an event will happen in the future for the given route
// and a named resource related to that route.
func (e *expectations) Expect(namespace, routeName, name string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: routeName}
	set, ok := e.expect[key]
	if !ok {
		set = sets.NewString()
		e.expect[key] = set
	}
	set.Insert(name)
}

// Satisfied clears the expectation for the given resource name on an
// route.
func (e *expectations) Satisfied(namespace, routeName, name string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: routeName}
	set := e.expect[key]
	set.Delete(name)
	if set.Len() == 0 {
		delete(e.expect, key)
	}
}

// Expecting returns true if the provided route is still waiting to
// see changes.
func (e *expectations) Expecting(namespace, routeName string) bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: routeName}
	return e.expect[key].Len() > 0
}

// Clear indicates that all expectations for the given route should
// be cleared.
func (e *expectations) Clear(namespace, routeName string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	key := queueKey{namespace: namespace, name: routeName}
	delete(e.expect, key)
}

type queueKey struct {
	namespace string
	name      string
}

// NewController instantiates a Controller
func NewController(eventsClient kv1core.EventsGetter, routeClient routeclient.RoutesGetter, secrets coreinformers.SecretInformer, routes routeinformers.RouteInformer) *Controller {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	// TODO: remove the wrapper when every clients have moved to use the clientset.
	broadcaster.StartRecordingToSink(&kv1core.EventSinkImpl{Interface: eventsClient.Events("")})
	recorder := broadcaster.NewRecorder(legacyscheme.Scheme, corev1.EventSource{Component: "ingress-to-route-controller"})

	c := &Controller{
		eventRecorder: recorder,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-to-route"),

		expectations:     newExpectations(),
		expectationDelay: 2 * time.Second,

		routeClient: routeClient,

		secretLister: secrets.Lister(),
		routeLister:  routes.Lister(),

		syncs: []cache.InformerSynced{
			secrets.Informer().HasSynced,
			routes.Informer().HasSynced,
		},
	}

	// any change to a secret of type TLS in the namespace
	secrets.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *corev1.Secret:
				return t.Type == corev1.SecretTypeTLS || t.Type == corev1.SecretTypeOpaque
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.processNamespace,
			DeleteFunc: c.processNamespace,
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.processNamespace(newObj)
			},
		},
	})

	// skip reconciliation if the route is owned by an Ingress
	routes.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *routev1.Route:
				return !hasIngressOwnerRef(t.GetOwnerReferences())
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.processRoute,
			DeleteFunc: c.processRoute,
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.processRoute(newObj)
			},
		},
	})

	klog.Info("ingress-to-route metrics registered with prometheus")

	return c
}

func (c *Controller) processNamespace(obj interface{}) {
	switch t := obj.(type) {
	case metav1.Object:
		ns := t.GetNamespace()
		if len(ns) == 0 {
			utilruntime.HandleError(fmt.Errorf("object %T has no namespace", obj))
			return
		}
		c.queue.Add(queueKey{namespace: ns})
	default:
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
	}
}

func (c *Controller) processRoute(obj interface{}) {
	switch t := obj.(type) {
	case *routev1.Route:
		if hasIngressOwnerRef(t.OwnerReferences) {
			return
		}
		c.expectations.Satisfied(t.Namespace, t.Name, t.Name)
		c.queue.Add(queueKey{namespace: t.Namespace, name: t.Name})
	default:
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
	}
}

// Run begins watching and syncing.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting controller")

	if !cache.WaitForCacheSync(stopCh, c.syncs...) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	klog.Infof("Shutting down controller")
}

func (c *Controller) worker() {
	for c.processNext() {
	}
	klog.V(4).Infof("Worker stopped")
}

func (c *Controller) processNext() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	klog.V(5).Infof("processing %v begin", key)
	err := c.sync(key.(queueKey))
	c.handleNamespaceErr(err, key)
	klog.V(5).Infof("processing %v end", key)

	return true
}

func (c *Controller) handleNamespaceErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	klog.V(4).Infof("Error syncing %v: %v", key, err)
	c.queue.AddRateLimited(key)
}

func (c *Controller) syncRouteStatus(route *routev1.Route, secret *corev1.Secret) {
	if hasSecretReferenceAnnotation(route.GetAnnotations()) &&
		secretValid(secret) &&
		route.Spec.TLS != nil && len(route.Spec.TLS.Certificate) > 0 && len(route.Spec.TLS.Key) > 0 {

		// update status to Managed
	} else if hasSecretReferenceAnnotation(route.GetAnnotations()) &&
		!secretValid(secret) &&
		route.Spec.TLS != nil && len(route.Spec.TLS.Certificate) == 0 && len(route.Spec.TLS.Key) == 0 {

		// TODO: check if route.TLS can be nil

		// update status to Default
		// update message as defaulting due to secret not found
	} else if !hasSecretReferenceAnnotation(route.GetAnnotations()) && route.Spec.TLS != nil && len(route.Spec.TLS.Certificate) > 0 && len(route.Spec.TLS.Key) > 0 {
		// update status to Custom
	}
}

func (c *Controller) sync(key queueKey) error {
	// sync all ingresses in the namespace
	if len(key.name) == 0 {
		routes, err := c.routeLister.Routes(key.namespace).List(labels.Everything())
		if err != nil {
			return err
		}
		for _, route := range routes {
			c.queue.Add(queueKey{namespace: route.Namespace, name: route.Name})
		}
		return nil
	}

	route, err := c.routeClient.Routes(key.namespace).Get(context.TODO(), key.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// validate if route needs to be processed and if it has all pre-conditions
	// satisfied

	// check if route is owned by ingress-to-route Controller
	if hasIngressOwnerRef(route.GetOwnerReferences()) {
		klog.V(5).Infof("Route %s/%s is owned by ingress-to-route controller", key.namespace, key.name)
		return nil
	}

	// check if route is either Edge or Re-encrypt
	if route.Spec.TLS != nil && route.Spec.TLS.Termination == routev1.TLSTerminationPassthrough {
		c.expectations.Satisfied(key.namespace, key.name, key.name)
		klog.V(5).Infof("Route %s/%s has TLS termination of type %q, not supported by secret-injector controller", key.namespace, key.name, route.Spec.TLS.Termination)
		return nil
	}

	// check if route needs to be reconciled or not based on status
	// check have vs want

	// if have != want

	updatedRoute := route.DeepCopy()
	secretRef *corev1.Secret = nil
	if _, ok := updatedRoute.GetAnnotations()[secretReferenceAnnotation]; ok {
		// if route has secret reference annotation
		// need to process secret and update route

	} else {

	}

	c.syncRouteStatus(updatedRoute)

	c.routeClient.Routes(updatedRoute.Namespace).Update(context.TODO(), updatedRoute, metav1.UpdateOptions{})

	return nil
}

// validOwnerRefAPIVersions is a set of recognized API versions for the ingress
// owner ref.
var validOwnerRefAPIVersions = sets.NewString(
	"networking.k8s.io/v1",
	"networking.k8s.io/v1beta1",
	"extensions.k8s.io/v1beta1",
)

func hasIngressOwnerRef(owners []metav1.OwnerReference) bool {
	for _, ref := range owners {
		if ref.Kind != "Ingress" || !validOwnerRefAPIVersions.Has(ref.APIVersion) || ref.Controller == nil || !*ref.Controller {
			continue
		}
		return true
	}
	return false
}

func hasSecretReferenceAnnotation(annotations map[string]string) bool {
	_, ok := annotations[secretReferenceAnnotation]
	return ok
}

func secretValid(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}

	return secret.Type == corev1.SecretTypeTLS || secret.Type == corev1.SecretTypeOpaque
}
