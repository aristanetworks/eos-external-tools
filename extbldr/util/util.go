// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"os"
	"os/exec"
)

// RunSystemCmd runs a command on the shell and pipes to stdout and stderr
func RunSystemCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// MaybeCreateDir creates a directory with permissions 0775
// Pre-existing directories are left untouched.
func MaybeCreateDir(dirPath string) error {
	err := os.Mkdir(dirPath, 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
