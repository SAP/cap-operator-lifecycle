/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/
package operator

import (
	"testing"
)

func TestCheckDirectoryExists(t *testing.T) {

	// invalid directory
	if err := checkDirectoryExists("invalid"); err == nil {
		t.Error("error expected but not returned")
		return
	}

	// valid directory
	if err := checkDirectoryExists("../../chart"); err != nil {
		t.Error("error not expected but returned")
		return
	}

	// File path passed instead of a directory
	if err := checkDirectoryExists("../../chart/values.yaml"); err == nil {
		t.Error("error expected but not returned")
		return
	}
}
