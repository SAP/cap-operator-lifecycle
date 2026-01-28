---
title: "Kyma Cluster"
linkTitle: "Kyma Cluster"
weight: 10
type: "docs"
tags: ["setup"]
description: >
  How to install CAP Operator in a Kyma cluster
---

The CAP Operator is available as a [Community module](https://kyma-project.io/external-content/community-modules/docs/user/README) in Kyma clusters.

To enable the CAP Operator module in your Kyma cluster, follow these steps:

1. Open the Kyma Console, navigate to the **Modules** section, and click the **Add** button.

   ![community-module-1](/cap-operator-lifecycle/img/community-module-1.png)

2. Click the **Add** button in the Source YAMLs section to load the list of community modules.

   ![community-module-2](/cap-operator-lifecycle/img/community-module-2.png)

3. In the dialog that opens, you can see the list of available community modules and click the **Add** button.

   ![community-module-3](/cap-operator-lifecycle/img/community-module-3.png)

4. Select the **CAP Operator** module and click the **Add** button.

   ![community-module-4](/cap-operator-lifecycle/img/community-module-4.png)

5. Wait for the automatic installation of the CAP Operator components into the `cap-operator-system` namespace to complete.

   ![community-module-5](/cap-operator-lifecycle/img/community-module-5.png)
