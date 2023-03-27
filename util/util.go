// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/spf13/viper"
)

// Globals type struct exported for global flags
type Globals struct {
	Quiet bool
	Arch  string
}

// GlobalVar global variable exported for global flags
var GlobalVar Globals

// ErrPrefix is a container type for error prefix strings.
type ErrPrefix string

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

// CheckPath checks if path exists. It also optionally checks if it is a directory,
// or if the path is writable
func CheckPath(path string, checkDir bool, checkWritable bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if checkDir && !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	if checkWritable && unix.Access(path, unix.W_OK) != nil {
		return fmt.Errorf("%s is not writable", path)
	}
	return nil
}

// MaybeCreateDir creates a directory with permissions 0775
// Pre-existing directories are left untouched.
func MaybeCreateDir(dirPath string, errPrefix ErrPrefix) error {
	err := os.Mkdir(dirPath, 0775)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("%s: Error '%s' creating %s", errPrefix, err, dirPath)
	}
	return nil
}

// MaybeCreateDirWithParents creates a directory at dirPath if one
// doesn't already exist. It also creates any parent directories.
func MaybeCreateDirWithParents(dirPath string, errPrefix ErrPrefix) error {
	if err := RunSystemCmd("mkdir", "-p", dirPath); err != nil {
		return fmt.Errorf("%sError '%s' trying to create directory %s with parents",
			errPrefix, err, dirPath)
	}
	return nil
}

// RemoveDirs removes the directories dirs
func RemoveDirs(dirs []string, errPrefix ErrPrefix) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("%sError '%s' while removing %s",
				errPrefix, err, dir)
		}
	}
	return nil
}

// CopyToDestDir copies files/dirs specified by srcGlob to destDir
// It is assumed that destDir is present and writable
func CopyToDestDir(
	srcGlob string,
	destDir string,
	errPrefix ErrPrefix) error {

	if err := CheckPath(destDir, true, true); err != nil {
		return fmt.Errorf("%sDirectory %s should be present and writable: %s",
			errPrefix, destDir, err)
	}

	filesToCopy, patternErr := filepath.Glob(srcGlob)
	if patternErr != nil {
		return fmt.Errorf("%sGlob %s returned %s", errPrefix, srcGlob, patternErr)
	}

	for _, file := range filesToCopy {
		insideDestDir := destDir + "/"
		if err := RunSystemCmd("cp", "-rf", file, insideDestDir); err != nil {
			return fmt.Errorf("%scopying %s to %s errored out with '%s'",
				errPrefix, file, insideDestDir, err)
		}
	}
	return nil
}

// GetRepoDir returns the path of the cloned source repo.
// If repo is specified, it's subpath under SrcDir config is
// returned.
// If no repo is specfied, we return current working directory.
func GetRepoDir(repo string) string {
	var repoDir string
	if repo != "" {
		srcDir := viper.GetString("SrcDir")
		repoDir = filepath.Join(srcDir, repo)
	} else {
		repoDir = "."
	}
	return repoDir
}

// MaybeSetDefaultArch sets default architecture if command doesn't specify one
func MaybeSetDefaultArch() error {
	if GlobalVar.Arch == "" {
		var output []byte
		var err error
		if output, err = exec.Command("arch").Output(); err != nil {
			return fmt.Errorf("Running arch returned '%s'", err)
		}
		GlobalVar.Arch = strings.TrimRight(string(output), "\n")
	}
	return nil
}
