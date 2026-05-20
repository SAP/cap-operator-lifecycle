/*
SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	operatorv1alpha1 "github.com/sap/cap-operator-lifecycle/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CheckDirectoryExists(path string) error {
	fsinfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fsinfo.IsDir() {
		return errors.Errorf("not a directory: %s", path)
	}
	return nil
}

func GetCAPOperator(cl client.Client) (*operatorv1alpha1.CAPOperator, error) {
	capOperatorList := &operatorv1alpha1.CAPOperatorList{}

	err := cl.List(context.TODO(), capOperatorList, &client.ListOptions{Namespace: corev1.NamespaceAll})
	if err != nil {
		return nil, fmt.Errorf("failed to list CAPOperator resources: %w", err)
	}

	if len(capOperatorList.Items) == 0 {
		return nil, fmt.Errorf("no CAPOperator resource found")
	}

	if len(capOperatorList.Items) > 1 {
		return nil, fmt.Errorf("more than one CAPOperator resource found")
	}

	return &capOperatorList.Items[0], nil
}
