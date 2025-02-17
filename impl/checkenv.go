// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

var parsedMockCfgTemplate *template.Template

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

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

	// Check SrcDir or default to checking eext.yaml in current directory
	if srcDir != "" {
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			aggError += fmt.Sprintf("\ntrouble with SrcDir: %s", err)
			failed = true
		}
	} else {
		if _, err := os.Stat("./eext.yaml"); os.IsNotExist(err) {
			aggError += fmt.Sprintf("\nNo eext.yaml in current directory. " +
				"SrcDir is unspecified, so it is expected that no --repo will be specified, " +
				"and that the sources are in the current working directory.")
		}
	}

	// Check WorkingDir
	if _, err := os.Stat(workingDir); os.IsNotExist(err) || !isDirectory(workingDir) {
		aggError += fmt.Sprintf("\ntrouble with WorkingDir: %s", err)
		failed = true
	}

	// Check DestDir
	if _, err := os.Stat(destDir); os.IsNotExist(err) || !isDirectory(destDir) {
		aggError += fmt.Sprintf("\ntrouble with DestDir: %s", err)
		failed = true
	}

	// Check MockCfgTemplate
	if _, err := os.Stat(mockCfgTemplate); os.IsNotExist(err) {
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

	// Check DnfConfigFile
	if _, err := os.Stat(dnfConfigFile); os.IsNotExist(err) {
		aggError += fmt.Sprintf("\ntrouble with DnfConfigFile: %s", err)
		failed = true
	}

	// Check SrcConfigFile
	if _, err := os.Stat(srcConfigFile); os.IsNotExist(err) {
		aggError += fmt.Sprintf("\ntrouble with SrcConfigFile: %s", err)
		failed = true
	}

	// Check PkiPath
	if _, err := os.Stat(pkiPath); os.IsNotExist(err) || !isDirectory(pkiPath) {
		aggError += fmt.Sprintf("\ntrouble with PkiPath: %s", err)
		failed = true
	}

	if failed {
		return fmt.Errorf("Environment check failed:%s", aggError)
	}

	return nil
}
