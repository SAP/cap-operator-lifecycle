---
title: "Kyma Cluster"
linkTitle: "Kyma Cluster"
weight: 10
type: "docs"
tags: ["setup"]
description: >
  How to install CAP Operator using CAP Operator Manager in a Kyma cluster
---
## Install CAP Operator

- To install CAP Operator, open your Kyma busola dashboard and navigate to `kyma-system` namespace.

  ![kyma-namespace](/cap-operator-lifecycle/img/kyma-namespace.png)

- On the left tab, expand section `Kyma` and open `Module Templates`. Here, you will be able to see two entries - `cap-operator-fast` and `cap-operator-regular`. Currently, the version of CAP Operator is same on both the channels. So you can install anyone of them.

  ![kyma-module-template](/cap-operator-lifecycle/img/kyma-module-template.png)

- Open `Kyma` section from the left tab. 

  ![kyma-default](/cap-operator-lifecycle/img/kyma-default.png)

- Open resource `default` and click on Edit.

  ![kyma-default-edit](/cap-operator-lifecycle/img/kyma-default-edit.png)

- Under the `Modules` section, check `cap-operator` and click on update.

  ![kyma-default-edit-select-cap-op](/cap-operator-lifecycle/img/kyma-default-edit-select-cap-op.png)

- Now the Kyma lifecycle manager will first install the CAP Operator Manager and then the CAP Operator Manager will install the CAP Operator using the below default resource.

```yaml
apiVersion: operator.sme.sap.com/v1alpha1
kind: CAPOperator
metadata:
  labels:
    app.kubernetes.io/name: cap-operator
    app.kubernetes.io/instance: cap-operator
    app.kubernetes.io/part-of: cap-operator-manager
    app.kuberentes.io/managed-by: kustomize
  name: cap-operator
spec:
  subscriptionServer:
    subDomain: cap-op
  ingressGatewayLabels:
    - name: istio
      value: ingressgateway
    - name: app
      value: istio-ingressgateway
```
  ![kyma-kyma-cap-op-installing](/cap-operator-lifecycle/img/kyma-cap-op-installing.png)

- Once the CAP Operator is installed, the module state as well the kyma resource state will be `Ready`.

  ![kyma-kyma-cap-op-ready](/cap-operator-lifecycle/img/kyma-cap-op-ready.png)

- After the CAP Operator is installed, you will be able to see a new section `CAP Operator` on the left tab. Under this tab, there will be a section for each of the CAP Operator resources. Using this, you can navigate through the CAP Operator resources once the CAP application is installed.

  ![kyma-kyma-cap-op-resources](/cap-operator-lifecycle/img/kyma-cap-op-resources.png)