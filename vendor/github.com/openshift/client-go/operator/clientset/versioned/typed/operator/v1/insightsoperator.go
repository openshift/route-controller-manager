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

// InsightsOperatorsGetter has a method to return a InsightsOperatorInterface.
// A group's client should implement this interface.
type InsightsOperatorsGetter interface {
	InsightsOperators() InsightsOperatorInterface
}

// InsightsOperatorInterface has methods to work with InsightsOperator resources.
type InsightsOperatorInterface interface {
	Create(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.CreateOptions) (*v1.InsightsOperator, error)
	Update(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.UpdateOptions) (*v1.InsightsOperator, error)
	UpdateStatus(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.UpdateOptions) (*v1.InsightsOperator, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.InsightsOperator, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.InsightsOperatorList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.InsightsOperator, err error)
	Apply(ctx context.Context, insightsOperator *operatorv1.InsightsOperatorApplyConfiguration, opts metav1.ApplyOptions) (result *v1.InsightsOperator, err error)
	ApplyStatus(ctx context.Context, insightsOperator *operatorv1.InsightsOperatorApplyConfiguration, opts metav1.ApplyOptions) (result *v1.InsightsOperator, err error)
	InsightsOperatorExpansion
}

// insightsOperators implements InsightsOperatorInterface
type insightsOperators struct {
	client rest.Interface
}

// newInsightsOperators returns a InsightsOperators
func newInsightsOperators(c *OperatorV1Client) *insightsOperators {
	return &insightsOperators{
		client: c.RESTClient(),
	}
}

// Get takes name of the insightsOperator, and returns the corresponding insightsOperator object, and an error if there is any.
func (c *insightsOperators) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.InsightsOperator, err error) {
	result = &v1.InsightsOperator{}
	err = c.client.Get().
		Resource("insightsoperators").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of InsightsOperators that match those selectors.
func (c *insightsOperators) List(ctx context.Context, opts metav1.ListOptions) (result *v1.InsightsOperatorList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.InsightsOperatorList{}
	err = c.client.Get().
		Resource("insightsoperators").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested insightsOperators.
func (c *insightsOperators) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("insightsoperators").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a insightsOperator and creates it.  Returns the server's representation of the insightsOperator, and an error, if there is any.
func (c *insightsOperators) Create(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.CreateOptions) (result *v1.InsightsOperator, err error) {
	result = &v1.InsightsOperator{}
	err = c.client.Post().
		Resource("insightsoperators").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(insightsOperator).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a insightsOperator and updates it. Returns the server's representation of the insightsOperator, and an error, if there is any.
func (c *insightsOperators) Update(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.UpdateOptions) (result *v1.InsightsOperator, err error) {
	result = &v1.InsightsOperator{}
	err = c.client.Put().
		Resource("insightsoperators").
		Name(insightsOperator.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(insightsOperator).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *insightsOperators) UpdateStatus(ctx context.Context, insightsOperator *v1.InsightsOperator, opts metav1.UpdateOptions) (result *v1.InsightsOperator, err error) {
	result = &v1.InsightsOperator{}
	err = c.client.Put().
		Resource("insightsoperators").
		Name(insightsOperator.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(insightsOperator).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the insightsOperator and deletes it. Returns an error if one occurs.
func (c *insightsOperators) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("insightsoperators").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *insightsOperators) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("insightsoperators").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched insightsOperator.
func (c *insightsOperators) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.InsightsOperator, err error) {
	result = &v1.InsightsOperator{}
	err = c.client.Patch(pt).
		Resource("insightsoperators").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied insightsOperator.
func (c *insightsOperators) Apply(ctx context.Context, insightsOperator *operatorv1.InsightsOperatorApplyConfiguration, opts metav1.ApplyOptions) (result *v1.InsightsOperator, err error) {
	if insightsOperator == nil {
		return nil, fmt.Errorf("insightsOperator provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(insightsOperator)
	if err != nil {
		return nil, err
	}
	name := insightsOperator.Name
	if name == nil {
		return nil, fmt.Errorf("insightsOperator.Name must be provided to Apply")
	}
	result = &v1.InsightsOperator{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("insightsoperators").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *insightsOperators) ApplyStatus(ctx context.Context, insightsOperator *operatorv1.InsightsOperatorApplyConfiguration, opts metav1.ApplyOptions) (result *v1.InsightsOperator, err error) {
	if insightsOperator == nil {
		return nil, fmt.Errorf("insightsOperator provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(insightsOperator)
	if err != nil {
		return nil, err
	}

	name := insightsOperator.Name
	if name == nil {
		return nil, fmt.Errorf("insightsOperator.Name must be provided to Apply")
	}

	result = &v1.InsightsOperator{}
	err = c.client.Patch(types.ApplyPatchType).
		Resource("insightsoperators").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
