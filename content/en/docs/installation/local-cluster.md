---
title: "Local Cluster"
linkTitle: "Local Cluster"
weight: 20
type: "docs"
tags: ["setup"]
description: >
  How to install CAP Operator using CAP Operator Manager in a local cluster
---

## Install CAP Operator Manager
To install the latest version of CAP Operator Manager, please execute the following command:

```bash
kubectl apply -f https://github.com/SAP/cap-operator-lifecycle/releases/download/manager%2Fv0.0.1/manager_manifest.yaml
```

This would create namespace `cap-operator-system` with CAP Operator Manager installed. 

![cap-op-man-install](/cap-operator-lifecycle/img/cap-op-man-install.png)
## Install CAP Operator using CAP Operator Manager
Once the CAP Operator Manager is running, you can install CAP operator by executing the following command:

```bash
kubectl apply -n cap-operator-system -f https://github.com/SAP/cap-operator-lifecycle/releases/download/manager%2Fv0.0.1/manager_default_CR.yaml
```
**This would work only if the `ingressGatewayLabels` in your clusters matches the following values:**

```bash
ingressGatewayLabels:
  istio: ingressgateway
  app: istio-ingressgateway
```

If not, you will have to manually create the `CAPOperator` resource by applying below yaml to `cap-operator-system` namespace after filling the `ingressGatewayLabels` values from your cluster.

```yaml
apiVersion: operator.sme.sap.com/v1alpha1
kind: CAPOperator
metadata:
  name: cap-operator
spec:
  subscriptionServer:
    subDomain: cap-op
  ingressGatewayLabels:
    istio: <<--istio-->>
    app: <<--app-->>
```
Once the `CAPOperator` resource is created, the CAP Operator Manager will start installing the CAP Operator in the namespace. Once the resource is ready, you will be able to see the CAP Operator Pods running in the namespace.

![cap-op-man-cr-ready](/cap-operator-lifecycle/img/cap-op-man-cr-ready.png)

CAP Operator Pods:

![cap-op-pods](/cap-operator-lifecycle/img/cap-op-pods.png)