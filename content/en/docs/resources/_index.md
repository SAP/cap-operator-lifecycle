---
title: "Resources"
linkTitle: "Resources"
weight: 20
type: "docs"
description: >
  How to configure the CAP Operator Manager resources
---

CAP Operator Manager uses resource `CAPOperator` to install the CAP Operator. The `CAPOperator` resource has the following attributes -

- `dnsTarget` _string_ - Public ingress URL for the cluster Load Balancer
- `subDomain` _string_ - Subdomain of the CAP Operator subscription Server
- `ingressGatewayLabels` - Labels used to identify the istio ingress-gateway component and its corresponding namespace. Usually {“app”:“istio-ingressgateway”,“istio”:“ingressgateway”}

The below example shows a fully configured `CAPOperator` resource:

```yaml
apiVersion: operator.sme.sap.com/v1alpha1
kind: CAPOperator
metadata:
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

Here, we will automatically determine the cluster shoot domain and install the CAP Operator by setting the subscription server domain and the DNS Target. The DNS target is derived using the `ingressGatewayLabels`. For the above example, if the determined the cluster shoot domain is `test.stage.kyma.ondemand.com`, then the domain will be set as `cap-op.test.stage.kyma.ondemand.com` by default. 

>Note: The length of the domain should be less than 64 characters. Depending up on your cluster shoot domain, please choose a length appropriate subdomain.

The user can also maintain the DNS Target manually. In such cases, we will take over the value as it is. The user can maintain the DNS Target as shown below:

```yaml
apiVersion: operator.sme.sap.com/v1alpha1
kind: CAPOperator
metadata:
  name: cap-operator
spec:
  subscriptionServer:
    subDomain: cap-op
  dnsTarget: public-ingress-custom.test.stage.kyma.ondemand.com
```
