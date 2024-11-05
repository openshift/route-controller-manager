# **Route Controller Manager Hacking**

## Building and Deploying

This document explains how to modify and test changes to the route-controller-manager in an OpenShift cluster.

### Prerequisites

* An OpenShift cluster
* An admin-scoped KUBECONFIG for the cluster

### Preparing the Cluster

To modify the route-controller-manager image, you need to scale down the operators managing it. This will allow you to make changes without the operators automatically reverting them.

### Scaling Down Operators

First, scale down the cluster-version-operator (CVO) and the openshift-controller-manager-operator. These operators will try to manage and reconcile changes, so stopping them temporarily will let you make changes to the route-controller-manager image.
```
$ oc -n openshift-cluster-version scale --replicas=0 deployments/cluster-version-operator
$ oc -n openshift-controller-manager-operator scale --replicas=0 deployments/openshift-controller-manager-operator
```
Confirm that the deployments have been scaled down by checking their status:

```
$ oc get deployments -n openshift-cluster-version cluster-version-operator
$ oc get deployments -n openshift-controller-manager-operator openshift-controller-manager-operator
```

### Modifying the Route Controller Manager Image

Once the relevant operators are scaled down, you can proceed to update the route-controller-manager deployment:
Edit the route-controller-manager deployment to specify a new image:
```
$ oc -n openshift-route-controller-manager edit deployment route-controller-manager
```

### Restoring the Operators

After you've tested your changes, scale the operators back up to restore normal management:
```
$ oc -n openshift-cluster-version scale --replicas=1 deployments/cluster-version-operator
$ oc -n openshift-controller-manager-operator scale --replicas=1 deployments/openshift-controller-manager-operator
```

#### Additional Notes

Scaling Down: Scaling down the operators temporarily prevents automatic reconciliation, allowing you to modify the deployment configuration directly.