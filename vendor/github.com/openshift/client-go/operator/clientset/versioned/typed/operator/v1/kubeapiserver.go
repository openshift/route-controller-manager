// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1 "github.com/openshift/api/operator/v1"
	operatorv1 "github.com/openshift/client-go/operator/applyconfigurations/operator/v1"
	scheme "github.com/openshift/client-go/operator/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KubeAPIServersGetter has a method to return a KubeAPIServerInterface.
// A group's client should implement this interface.
type KubeAPIServersGetter interface {
	KubeAPIServers() KubeAPIServerInterface
}

// KubeAPIServerInterface has methods to work with KubeAPIServer resources.
type KubeAPIServerInterface interface {
	Create(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.CreateOptions) (*v1.KubeAPIServer, error)
	Update(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.UpdateOptions) (*v1.KubeAPIServer, error)
	UpdateStatus(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.UpdateOptions) (*v1.KubeAPIServer, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.KubeAPIServer, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.KubeAPIServerList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeAPIServer, err error)
	Apply(ctx context.Context, kubeAPIServer *operatorv1.KubeAPIServerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeAPIServer, err error)
	ApplyStatus(ctx context.Context, kubeAPIServer *operatorv1.KubeAPIServerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeAPIServer, err error)
	KubeAPIServerExpansion
}

// kubeAPIServers implements KubeAPIServerInterface
type kubeAPIServers struct {
	client rest.Interface
}

// newKubeAPIServers returns a KubeAPIServers
func newKubeAPIServers(c *OperatorV1Client) *kubeAPIServers {
	return &kubeAPIServers{
		client: c.RESTClient(),
	}
}

// Get takes name of the kubeAPIServer, and returns the corresponding kubeAPIServer object, and an error if there is any.
func (c *kubeAPIServers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.KubeAPIServer, err error) {
	result = &v1.KubeAPIServer{}
	err = c.client.Get().
		Resource("kubeapiservers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KubeAPIServers that match those selectors.
func (c *kubeAPIServers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KubeAPIServerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.KubeAPIServerList{}
	err = c.client.Get().
		Resource("kubeapiservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kubeAPIServers.
func (c *kubeAPIServers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("kubeapiservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kubeAPIServer and creates it.  Returns the server's representation of the kubeAPIServer, and an error, if there is any.
func (c *kubeAPIServers) Create(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.CreateOptions) (result *v1.KubeAPIServer, err error) {
	result = &v1.KubeAPIServer{}
	err = c.client.Post().
		Resource("kubeapiservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeAPIServer).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kubeAPIServer and updates it. Returns the server's representation of the kubeAPIServer, and an error, if there is any.
func (c *kubeAPIServers) Update(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.UpdateOptions) (result *v1.KubeAPIServer, err error) {
	result = &v1.KubeAPIServer{}
	err = c.client.Put().
		Resource("kubeapiservers").
		Name(kubeAPIServer.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeAPIServer).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kubeAPIServers) UpdateStatus(ctx context.Context, kubeAPIServer *v1.KubeAPIServer, opts metav1.UpdateOptions) (result *v1.KubeAPIServer, err error) {
	result = &v1.KubeAPIServer{}
	err = c.client.Put().
		Resource("kubeapiservers").
		Name(kubeAPIServer.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeAPIServer).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kubeAPIServer and deletes it. Returns an error if one occurs.
func (c *kubeAPIServers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("kubeapiservers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kubeAPIServers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("kubeapiservers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kubeAPIServer.
func (c *kubeAPIServers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeAPIServer, err error) {
	result = &v1.KubeAPIServer{}
	err = c.client.Patch(pt).
		Resource("kubeapiservers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied kubeAPIServer.
func (c *kubeAPIServers) Apply(ctx context.Context, kubeAPIServer *operatorv1.KubeAPIServerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeAPIServer, err error) {
	if kubeAPIServer == nil {
		return nil, fmt.Errorf("kubeAPIServer provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(kubeAPIServer)
	if err != nil {
		return nil, err
	}
	name := kubeAPIServer.Name
	if name == nil {
		return nil, fmt.Errorf("kubeAPIServer.Name must be provided to Apply")
	}
	result = &v1.KubeAPIServer{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("kubeapiservers").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *kubeAPIServers) ApplyStatus(ctx context.Context, kubeAPIServer *operatorv1.KubeAPIServerApplyConfiguration, opts metav1.ApplyOptions) (result *v1.KubeAPIServer, err error) {
	if kubeAPIServer == nil {
		return nil, fmt.Errorf("kubeAPIServer provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(kubeAPIServer)
	if err != nil {
		return nil, err
	}

	name := kubeAPIServer.Name
	if name == nil {
		return nil, fmt.Errorf("kubeAPIServer.Name must be provided to Apply")
	}

	result = &v1.KubeAPIServer{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("kubeapiservers").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
