// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"

	"golang.org/x/sys/unix"
)

var parsedMockCfgTemplate *template.Template

// CheckEnv returns an error if there's a problem with the environment.
func CheckEnv() error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")
	destDir := viper.GetString("DestDir")
	mockCfgTemplate := viper.GetString("MockCfgTemplate")
	dnfConfigFile := viper.GetString("DnfConfigFile")
	srcConfigFile := viper.GetString("SrcConfigFile")
	pkiPath := viper.GetString("PkiPath")

	var aggError string
	failed := false
	if srcDir != "" {
		info, err := os.Stat(srcDir)
		if err != nil {
			aggError += fmt.Sprintf("\ntrouble with SrcDir: %s", err)
			failed = true
		} else if !info.IsDir() {
			aggError += fmt.Sprintf("\nSrcDir is not a directory: %s", srcDir)
			failed = true
		}
	} else {
		if _, pathErr := os.Stat("./eext.yaml"); pathErr != nil {
			aggError += fmt.Sprintf("\nNo eext.yaml in current directory. " +
				"SrcDir is unspecified, so it is expected  that no --repo will be specified, " +
				"and that the sources are in current working directory.")
		}
	}
	info, err := os.Stat(workingDir)
	if err != nil {
		aggError += fmt.Sprintf("\ntrouble with WorkingDir: %s", err)
		failed = true
	} else if !info.IsDir() {
		aggError += fmt.Sprintf("\nWorkingDir is not a directory: %s", workingDir)
		failed = true
	}
	info, err = os.Stat(destDir)
	if err != nil {
		aggError += fmt.Sprintf("\ntrouble with DestDir: %s", err)
		failed = true
	} else if !info.IsDir() {
		aggError += fmt.Sprintf("\nDestDir is not a directory: %s", destDir)
		failed = true
	} else if unix.Access(destDir, unix.W_OK) != nil {
		aggError += fmt.Sprintf("\nDestDir: %s is not writable", destDir)
		failed = true
	}

	if _, err := os.Stat(mockCfgTemplate); err != nil {
		aggError += fmt.Sprintf("\ntrouble with MockCfgTemplate: %s", err)
		failed = true
	} else if parsedMockCfgTemplate == nil {
		// Only parse once
		var parseErr error
		if parsedMockCfgTemplate, parseErr = template.ParseFiles(mockCfgTemplate); parseErr != nil {
			aggError += fmt.Sprintf("\ntrouble with MockCfgTemplate: %s", parseErr)
			failed = true
		}
	}

	if _, err := os.Stat(dnfConfigFile); err != nil {
		aggError += fmt.Sprintf("\ntrouble with DnfConfigFile: %s", err)
		failed = true
	}

	if _, err := os.Stat(srcConfigFile); err != nil {
		aggError += fmt.Sprintf("\ntrouble with SrcConfigFile: %s", err)
		failed = true
	}
	info, err = os.Stat(pkiPath)
	if err != nil {
		aggError += fmt.Sprintf("\ntrouble with PkiPath: %s", err)
		failed = true
	} else if !info.IsDir() {
		aggError += fmt.Sprintf("\nPkiPath is not a directory: %s", srcDir)
		failed = true
	}

	if failed {
		return fmt.Errorf("Environment check failed:%s", aggError)
	}

	return nil
}
