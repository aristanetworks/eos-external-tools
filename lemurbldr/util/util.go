// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Globals type struct exported for global flags
type Globals struct {
	Quiet bool
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

// CreateDirs creates the specified directories
// It calls mkdir, it doesn't create parent dirs.
// cleanup indicates whether to cleanup existing directories
func CreateDirs(dirs []string, cleanup bool, errPrefix ErrPrefix) error {
	for _, dir := range dirs {
		if cleanup {
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("%sRemoving %s errored out with %s", errPrefix, dir, err)
			}
		}
		if err := MaybeCreateDir(dir, errPrefix); err != nil {
			return err
		}
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

// CopyFile runs cp src dest in the shell
func CopyFile(src string, dest string, errPrefix ErrPrefix) error {
	var cpErr error
	cpErr = RunSystemCmd("cp", src, dest)
	if cpErr != nil {
		return fmt.Errorf("%s: Error '%s' copying %s to %s", errPrefix, cpErr, src, dest)
	}
	return nil
}

// GetMatchingFilenamesFromDir gets list of files matching the regex
func GetMatchingFilenamesFromDir(
	dirPath string, regexString string,
	errPrefix ErrPrefix) ([]string, error) {
	var fileNames []string
	files, readDirErr := os.ReadDir(dirPath)
	if readDirErr != nil {
		retErr := fmt.Errorf("%sutil.GetMatchingFilenamesFromDir: os.ReadDir(%s) returned %s",
			errPrefix, dirPath, readDirErr)
		return nil, retErr
	}

	matchRegexp, regexErr := regexp.Compile(regexString)
	if regexErr != nil {
		retErr := fmt.Errorf("%sutil.GetMatchingFilenamesFromDir: regexp.Compile(%s) returned %s",
			errPrefix, regexString, regexErr)
		return nil, retErr
	}

	var matched bool
	for _, file := range files {
		matched = matchRegexp.MatchString(file.Name())
		if matched {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}

// CopyFilesToDir copies files in filelist from srcDir to destDir
// It expects srcDir to be already present, destDir can be created (with
// parents) on demand.
func CopyFilesToDir(fileList []string, srcDir string, destDir string,
	retainInSrc bool,
	errPrefix ErrPrefix) error {
	if err := CheckPath(srcDir, true, false); err != nil {
		return fmt.Errorf("%s: Expected directory %s to be present. (%s)",
			errPrefix, srcDir, err)
	}

	if len(fileList) == 0 {
		return nil
	}

	if err := MaybeCreateDirWithParents(destDir, errPrefix); err != nil {
		return err
	}

	cmd := "cp"
	if !retainInSrc {
		cmd = "mv"
	}
	for _, file := range fileList {
		fileErr := RunSystemCmd(cmd, "-f", filepath.Join(srcDir, file), destDir)
		if fileErr != nil {
			return fmt.Errorf("%scopy/move file %s errored out with %s", errPrefix, file, fileErr)
		}
	}
	return nil
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
