// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.arista.io/eos/tools/eext/executor"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

// Globals type struct exported for global flags
type Globals struct {
	Quiet bool
}

// GlobalVar global variable exported for all global variables
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

// Runs the system command from a specified directory
func RunSystemCmdInDir(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if !GlobalVar.Quiet {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = io.Discard
	}
	err := cmd.Run()
	return err
}

// CheckOutput runs a command on the shell and returns stdout if it is successful
// else it return the error
func CheckOutput(name string, arg ...string) (
	string, error) {
	cmd := exec.Command(name, arg...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output),
				fmt.Errorf("running '%s %s': exited with exit-code %d\nstderr:\n%s",
					name, strings.Join(arg, " "), exitErr.ExitCode(), exitErr.Stderr)
		}
		return string(output),
			fmt.Errorf("running '%s %s' failed with '%s'",
				name, strings.Join(arg, " "), err)
	}
	return string(output), nil
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

// MaybeCreateDirWithParents creates a directory at dirPath if one
// doesn't already exist. It also creates any parent directories.
func MaybeCreateDirWithParents(dirPath string, executor executor.Executor, errPrefix ErrPrefix) error {
	if err := executor.Exec("mkdir", "-p", dirPath); err != nil {
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

// Generate SHA256 hash of file
func GenerateSha256Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("GenerateSha256Hash: %s", err)
	}
	defer file.Close()
	hashComputer := sha256.New()
	if _, err := io.Copy(hashComputer, file); err != nil {
		return "", fmt.Errorf("GenerateSha256Hash: %s", err)
	}
	sha256Hash := fmt.Sprintf("%x", hashComputer.Sum(nil))
	return sha256Hash, nil
}
