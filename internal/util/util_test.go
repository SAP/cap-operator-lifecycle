/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package util

import "testing"

func TestCheckDirectoryExists(t *testing.T) {

	// invalid directory
	if err := CheckDirectoryExists("invalid"); err == nil {
		t.Error("error expected but not returned")
		return
	}

	// valid directory
	if err := CheckDirectoryExists("../../chart"); err != nil {
		t.Error("error not expected but returned")
		return
	}

	// File path passed instead of a directory
	if err := CheckDirectoryExists("../../chart/values.yaml"); err == nil {
		t.Error("error expected but not returned")
		return
	}
}
