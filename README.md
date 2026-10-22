# Route Controller Manager

The Route Controller Manager consists of additional controllers that enhance Openshift Routes, Ingresses, and Services.

## Ingress to Route Controller

Controller ensures that zero or more routes exist to match any supported ingress. The
controller creates a controller owner reference from the route to the parent ingress,
allowing users to orphan their ingress. All owned routes have specific spec fields
managed (those attributes present on the ingress), while any other fields may be
modified by the user.

### Insecure traffic policy for generated TLS Routes

Set the `route.openshift.io/insecure-edge-termination-policy` annotation on an
Ingress to control `spec.tls.insecureEdgeTerminationPolicy` on its generated TLS
Routes. Values are case-sensitive:

| Value | Behavior |
|-------|----------|
| `Redirect` | Redirect insecure HTTP connections to HTTPS. |
| `Allow` | Allow insecure HTTP connections with edge or reencrypt termination. |
| `None` | Disable insecure HTTP connections. |
| `""` | Disable insecure HTTP connections, like `None`, using an empty Route policy. |

For example, explicitly request an empty policy with a quoted empty string:

```yaml
metadata:
  annotations:
    route.openshift.io/insecure-edge-termination-policy: ""
```

A supported value controls every generated TLS Route and overrides direct edits
to that Route's insecure traffic policy. Removing the annotation relinquishes
control and preserves existing Route policies. Without the annotation, a newly
generated TLS configuration defaults to `Redirect` unless an existing TLS Route
policy is preserved. This controller default differs from the Route API default:
an omitted or empty Route policy means `None`, not `Redirect`.

Invalid values and `Allow` with passthrough termination are ignored: new TLS
configurations use the default, and existing TLS Route policies are preserved,
except that an existing `Allow` policy is converted to `Redirect` whenever the
desired termination is passthrough. This includes transitions from edge or
reencrypt to passthrough, because the Route API forbids `Allow` with passthrough.
The same exception applies when the annotation is absent or removed; compatible
existing policies remain unchanged.
Passthrough supports `Redirect`, `None`, and the empty value. The annotation does
not enable TLS and has no effect on the TLS configuration of non-TLS Routes,
although the annotation itself is copied to generated Routes.


## Service Ingress IP Controller

Controller is responsible for allocating ingress ip addresses to Service objects of type LoadBalancer. 
It allocates adresses from `spec.observedConfig.ingress.ingressIPNetworkCIDR` range in
`openshiftcontrollermanagers.operator.openshift.io cluster` config and can be used to assign a unique external IP addresses.

