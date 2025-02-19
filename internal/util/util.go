/*
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and cap-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"os"

	"github.com/pkg/errors"
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
