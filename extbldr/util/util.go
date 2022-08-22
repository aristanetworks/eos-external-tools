// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Globals type struct exported for global flags
type Globals struct {
	Quiet bool
}

// GlobalVar global variable exported for global flags
var GlobalVar Globals

// RunSystemCmd runs a command on the shell and pipes to stdout and stderr
func RunSystemCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	if !GlobalVar.Quiet {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = io.Discard
	}
	err := cmd.Run()
	return err
}

// MaybeCreateDir creates a directory with permissions 0775
// Pre-existing directories are left untouched.
func MaybeCreateDir(errorPrefix string, dirPath string) error {
	err := os.Mkdir(dirPath, 0775)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("%s: Error '%s' creating %s", errorPrefix, err, dirPath)
	}
	return nil
}

// CopyFile runs cp src dst in the shell
func CopyFile(errorPrefix string, src string, dst string) error {
	var cpErr error
	cpErr = RunSystemCmd("cp", src, dst)
	if cpErr != nil {
		return fmt.Errorf("%s: Error '%s' copying %s to %s", errorPrefix, cpErr, src, dst)
	}
	return nil
}
