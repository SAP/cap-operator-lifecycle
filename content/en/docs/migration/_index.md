---
title: Migrate to Kyma Community Module
linkTitle: "Migrate to Kyma Community Module"
weight: 30
type: "docs"
description: >
  How to migrate CAP Operator installed via Helm, Manifest, or Kyma Module installation to the Kyma Community Module without deleting your deployed applications.
---

This page provides an overview of the steps required to migrate the CAP Operator installed via Helm, Manifest, or Kyma Module installation to the Kyma Community Module without deleting your deployed applications.

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
kubectl annotate capoperator cap-operator -n <namespace> operator.sme.sap.com/retain-resources="true"
```

2. Wait till the annotation is applied and the `CAPOperator` resource is ready.

3. Delete the `CAPOperator` resource using the following command:

```bash
kubectl apply -n cap-operator-system -f https://github.com/SAP/cap-operator-lifecycle/releases/latest/download/manager_default_CR.yaml
```
4. Delete all the resources created by the manifest installation using the following command:

```bash
kubectl apply -f https://github.com/SAP/cap-operator-lifecycle/releases/latest/download/manager_manifest.yaml
```

5. Install the CAP Operator using the Kyma Community Module by following the instructions in the [Kyma Cluster Installation Documentation](../installation/kyma-cluster.md).

## Kyma Module

If you have installed the CAP Operator by adding the Kyma Module, please follow the steps below to migrate to the Kyma Community Module:

1. Add the annotation `operator.sme.sap.com/retain-resources="true"` to the `CAPOperator` resource to ensure that your deployed applications are not deleted during the migration process. You can do this by running the following command:

```bash
kubectl annotate capoperator cap-operator -n kyma-system operator.sme.sap.com/retain-resources="true"
```

2. Wait till the annotation is applied and the `CAPOperator` resource is ready.

3. Uninstall the CAP Operator Kyma Module following the instructions in [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module#deleting-a-kyma-module).

4. Now install the CAP Operator using the Kyma Community Module by following the instructions in the [Kyma Cluster Installation Documentation](../installation/kyma-cluster.md).
