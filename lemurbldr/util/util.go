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

// GetMatchingFileNamesFromDir gets list of files matching the regex
func GetMatchingFileNamesFromDir(dirPath string, regexString string) ([]string, error) {
	var fileNames []string
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fileNames, fmt.Errorf("impl.getMatchingFileNamesFromDir: os.ReadDir returned %v", err)
	}

	matchRegexp, err := regexp.Compile(regexString)
	if err != nil {
		return fileNames, fmt.Errorf("impl.getMatchingFileNamesFromDir: regexp.compile returned %v", err)
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
func CopyFilesToDir(fileList []string, src string, dest string) error {
	creatErr := MaybeCreateDir("util.CopyFilesToDir", dest)
	if creatErr != nil {
		return creatErr
	}
	for _, file := range fileList {
		fileErr := RunSystemCmd("mv", "-f", filepath.Join(src, file), dest)
		if fileErr != nil {
			return fmt.Errorf("util.CopyFilesToDir: move file %s errored out with %s", file, fileErr)
		}
	}
	return nil
}

func CheckPath(path string, checkDir bool, checkWritable bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if checkDir && !info.IsDir() {
		return fmt.Errorf("%s is not a directory!", path)
	}

	if checkWritable && unix.Access(path, unix.W_OK) != nil {
		return fmt.Errorf("%s is not writable!", path)
	}
	return nil
}
