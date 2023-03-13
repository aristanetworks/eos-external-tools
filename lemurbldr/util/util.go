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

// GetMatchingFilenamesFromDir gets list of files matching the regex
func GetMatchingFilenamesFromDir(errPrefix string,
	dirPath string, regexString string) ([]string, error) {
	var fileNames []string
	files, readDirErr := os.ReadDir(dirPath)
	if readDirErr != nil {
		retErr := fmt.Errorf("%s: util.GetMatchingFilenamesFromDir: os.ReadDir(%s) returned %s",
			errPrefix, dirPath, readDirErr)
		return nil, retErr
	}

	matchRegexp, regexErr := regexp.Compile(regexString)
	if regexErr != nil {
		retErr := fmt.Errorf("%s: util.GetMatchingFilenamesFromDir: regexp.Compile(%s) returned %s",
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

// CopyFilesToDir copies files in filelist from src to dest dir, creates dest dir if it doesn't exist
func CopyFilesToDir(errPrefix string, fileList []string, src string, dest string, retainInSrc bool) error {
	creatErr := MaybeCreateDir(errPrefix, dest)
	if creatErr != nil {
		return creatErr
	}
	cmd := "cp"
	if !retainInSrc {
		cmd = "mv"
	}
	for _, file := range fileList {
		fileErr := RunSystemCmd(cmd, "-f", filepath.Join(src, file), dest)
		if fileErr != nil {
			return fmt.Errorf("%s: copy/move file %s errored out with %s", errPrefix, file, fileErr)
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

// CreateDirs creates the specified directories
// It calls mkdir, it doesn't create prefixes.
// cleanup indicates whether to cleanup existing directories
func CreateDirs(errPrefix string, dirs []string, cleanup bool) error {
	for _, dir := range dirs {
		if cleanup {
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("%s: Removing %s errored out with %s", errPrefix, dir, err)
			}
		}
		if err := MaybeCreateDir(errPrefix, dir); err != nil {
			return err
		}
	}
	return nil
}
