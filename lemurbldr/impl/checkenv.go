// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"text/template"

	"github.com/spf13/viper"

	"lemurbldr/util"
)

var parsedMockCfgTemplate *template.Template

// CheckEnv returns an error if there's a problem with the environment.
func CheckEnv() error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")
	destDir := viper.GetString("DestDir")
	mockCfgTemplate := viper.GetString("MockCfgTemplate")

	var aggError string
	failed := false
	if err := util.CheckPath(srcDir, true, false); err != nil {
		aggError += fmt.Sprintf("\ntrouble with SrcDir: %s", err)
		failed = true
	}

	if err := util.CheckPath(workingDir, true, true); err != nil {
		aggError += fmt.Sprintf("\ntrouble with WorkingDir: %s", err)
		failed = true
	}

	if err := util.CheckPath(destDir, true, true); err != nil {
		aggError += fmt.Sprintf("\ntrouble with DestDir: %s", err)
		failed = true
	}

	if err := util.CheckPath(mockCfgTemplate, false, false); err != nil {
		aggError += fmt.Sprintf("\ntrouble with MockCfgTemplate: %s", err)
	} else if parsedMockCfgTemplate == nil {
		// Only parse once
		var parseErr error
		if parsedMockCfgTemplate, parseErr = template.ParseFiles(mockCfgTemplate); parseErr != nil {
			aggError += fmt.Sprintf("\ntrouble with MockCfgTemplate: %s", parseErr)
		}
	}

	if failed {
		return fmt.Errorf("Environment check failed:%s", aggError)
	}
	return nil
}
