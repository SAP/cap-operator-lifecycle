---
title: Migrate to Kyma Community Module (Optional)
linkTitle: "Migrate to Kyma Community Module (Optional)"
weight: 30
type: "docs"
description: >
  How to migrate CAP Operator installed via Helm or Manifest to the Kyma Community Module without deleting your deployed applications.
---

This page provides an overview of the steps required to migrate the CAP Operator installed via Helm or Manifest to the Kyma Community Module without deleting your deployed applications.

## [Helm Installation](https://sap.github.io/cap-operator/docs/installation/helm/)

If you have installed the CAP Operator using Helm, please follow the steps below to migrate to the Kyma Community Module:

1. Uninstall the CAP Operator Helm release using the following command:

```bash
helm uninstall cap-operator -n cap-operator-system
```

> Note: The release name and namespace may vary based on your installation. Please ensure to use the correct release name and namespace in the command.

2. Install the CAP Operator using the Kyma Community Module by following the instructions in the [Kyma Cluster Installation Documentation](../installation/kyma-cluster.md).

## [Manifest Installation (Using CAP Operator Manager)](https://sap.github.io/cap-operator/docs/installation/cap-operator-manager/)

If you have installed the CAP Operator using a manifest, please follow the steps below to migrate to the Kyma Community Module:

1. Add the annotation `operator.sme.sap.com/retain-resources="true"` to the `CAPOperator` resource to ensure that your deployed applications are not deleted during the migration process. You can do this by running the following command:

```bash
kubectl annotate capoperator cap-operator operator.sme.sap.com/retain-resources="true"
```

2. Wait till the annotation is applied and the `CAPOperator` resource is ready.

3. Delete the `CAPOperator` resource using the following command:

```bash
kubectl delete -f https://github.com/SAP/cap-operator-lifecycle/releases/latest/download/manager_default_CR.yaml
```
4. Delete all the resources created by the manifest installation using the following command:

```bash
kubectl delete -f https://github.com/SAP/cap-operator-lifecycle/releases/latest/download/manager_manifest.yaml
```

5. Install the CAP Operator using the Kyma Community Module by following the instructions in the [Kyma Cluster Installation Documentation](../installation/kyma-cluster.md).
