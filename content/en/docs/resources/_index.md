---
title: "Resources"
linkTitle: "Resources"
weight: 20
type: "docs"
description: >
  How to configure the CAP Operator Manager resources
---

CAP Operator Manager uses resource `CAPOperator` to install the CAP Operator. The `CAPOperator` resource has the following attributes -

- `subscriptionServer.subDomain` _string_ - Subdomain of the CAP Operator subscription Server
- `subscriptionServer.certificateManager` _string_ - Certificate manager which can be either `Gardener` or `CertManager`
- `subscriptionServer.certificateConfig.gardener` -  Gardener certificate configuration. Relevant only if `subscriptionServer.certificateManager` is set to `Gardener`
- `subscriptionServer.certificateConfig.certManager` -  CertManager certificate configuration. Relevant only if `subscriptionServer.certificateManager` is set to `CertManager`
- `dnsTarget` _string_ - Public ingress URL for the cluster Load Balancer
- `ingressGatewayLabels` - Labels used to identify the istio ingress-gateway component and its corresponding namespace. Usually {“app”:“istio-ingressgateway”,“istio”:“ingressgateway”}
- `controller.detailedOperationalMetrics` _bool_ - Optionally enable detailed opertational metrics for the controller by setting this to true
- `controller.versionMonitoring.prometheusAddress` _string_ - URL of the Prometheus server from which metrics related to managed application versions can be queried
- `controller.versionMonitoring.metricsEvaluationInterval` - The duration (example 2h) after which versions are evaluated for deletion; based on specified workload metrics
- `controller.versionMonitoring.promClientAcquireRetryDelay` - The duration (example 10m) to wait before retrying to acquire Prometheus client and verify connection, after a failed attempt
- `controller.volumes` - Optionally specify list of additional volumes for the controller pod(s)
- `controller.volumeMounts` - Optionally specify list of additional volumeMounts for the controller container(s)
- `monitoring.enabled` _bool_ - Optionally enable Prometheus monitoring for all components
- `webhook.certificateManager` _string_ - Certificate manager which can be either `Default` or `CertManager`
- `webhook.certificateConfig.certManager` -  CertManager certificate configuration. Relevant only if `webhook.certificateManager` is set to `CertManager`

The below example shows a fully configured `CAPOperator` resource:

```yaml
apiVersion: operator.sme.sap.com/v1alpha1
kind: CAPOperator
metadata:
  name: cap-operator
spec:
  subscriptionServer:
    subDomain: cap-op
    certificateManager: Gardener
    certificateConfig:
      gardener:
        issuerName: "gardener-issuer-name"
        issuerNamespace: "gardener-issuer-namespace"
  ingressGatewayLabels:
    - name: istio
      value: ingressgateway
    - name: app
      value: istio-ingressgateway
  monitoring:
    enabled: true
  controller:
    detailedOperationalMetrics: true
    versionMonitoring:
      prometheusAddress: "http://prometheus-operated.monitoring.svc.cluster.local:9090" # <-- example of a Prometheus server running inside the same cluster
      promClientAcquireRetryDelay: "2h"
      metricsEvaluationInterval: "30m"
  webhook:
    certificateManager: CertManager
    certificateConfig:
      certManager:
        issuerGroup: "certManager-issuer-group"
        issuerKind: "certManager-issuer-kind"
        issuerName: "certManager-issuer-name"
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
